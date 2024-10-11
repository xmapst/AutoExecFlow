package queues

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/segmentio/ksuid"

	"github.com/xmapst/AutoExecFlow/internal/utils/wildcard"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

const defaultQueueSize = 1000

type memQueue struct {
	name    string
	ch      chan string
	subs    []*qsub
	unacked int32
	closed  atomic.Bool
	mu      sync.RWMutex // 使用读写锁来提高并发效率
	wg      sync.WaitGroup
}

func newMemQueue(name string) *memQueue {
	return &memQueue{
		name: name,
		ch:   make(chan string, defaultQueueSize),
		subs: make([]*qsub, 0),
	}
}

type qsub struct {
	ctx    context.Context
	cancel context.CancelFunc

	// topic only
	cname string
	ch    chan string
}

func (q *memQueue) publish(data string) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.closed.Load() {
		return
	}
	select {
	case q.ch <- data:
		logx.Infof("published message to subscriber queue %s", q.name)
	default:
		logx.Warnln("subscriber queue full, dropping message")
	}
}

func (q *memQueue) size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.ch)
}

func (q *memQueue) close() {
	if !q.closed.CompareAndSwap(false, true) {
		return
	}
	close(q.ch)

	q.mu.Lock()
	defer q.mu.Unlock()
	for _, sub := range q.subs {
		sub.cancel()
	}

	q.wg.Wait()
}

func (q *memQueue) subscribe(ctx context.Context, sub Handle) {
	logx.Infof("subscribing to queue %s", q.name)
	q.mu.Lock()
	ctx, cancel := context.WithCancel(ctx)
	q.subs = append(q.subs, &qsub{ctx: ctx, cancel: cancel})
	q.mu.Unlock()

	q.wg.Add(1)

	go func() {
		defer q.wg.Done()
		for {
			select {
			case <-ctx.Done():
				logx.Infof("subscribe queue closed %s", q.name)
				return
			case m, ok := <-q.ch:
				if !ok {
					logx.Infof("subscribe queue closed %s", q.name)
					return
				}
				atomic.AddInt32(&q.unacked, 1)
				if err := sub(m); err != nil {
					logx.Errorln("unexpected error occurred while processing task", err)
				}
				atomic.AddInt32(&q.unacked, -1)
			}
		}
	}()
}

type memTopic struct {
	name      string
	subs      []*qsub
	terminate chan struct{}
	mu        sync.RWMutex
}

func newMemTopic(name string) *memTopic {
	return &memTopic{
		name:      name,
		terminate: make(chan struct{}),
	}
}

func (t *memTopic) publish(event string) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, sub := range t.subs {
		select {
		case <-sub.ctx.Done():
			continue
		case sub.ch <- event:
			logx.Infof("published message to subscriber topic %s cname %s", t.name, sub.cname)
		default:
			logx.Warnln("subscriber topic full, dropping message")
		}
	}
}

func (t *memTopic) subscribe(ctx context.Context, handler Handle) {
	logx.Infof("subscribing to topic %s", t.name)
	t.mu.Lock()
	defer t.mu.Unlock()

	subCtx, cancel := context.WithCancel(ctx)
	sub := &qsub{
		cname:  ksuid.New().String(),
		ctx:    subCtx,
		cancel: cancel,
		ch:     make(chan string, 100),
	}
	t.subs = append(t.subs, sub)

	go func() {
		defer cancel()
		for {
			select {
			case <-sub.ctx.Done():
				logx.Infof("subscribe topic closed %s", t.name)
				return
			case m := <-sub.ch:
				if err := handler(m); err != nil {
					logx.Errorln("unexpected error occurred while processing task", err)
				}
			}
		}
	}()
}

func (t *memTopic) close() {
	close(t.terminate)

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, sub := range t.subs {
		sub.cancel()
	}
}

type memoryBroker struct {
	queues    sync.Map
	topics    sync.Map
	terminate atomic.Bool
}

func newInMemoryBroker() *memoryBroker {
	return &memoryBroker{}
}

func (m *memoryBroker) PublishTask(node string, data string) error {
	routingKey := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	q, _ := m.queues.LoadOrStore(routingKey, newMemQueue(routingKey))
	q.(*memQueue).publish(data)
	return nil
}

func (m *memoryBroker) SubscribeTask(ctx context.Context, node string, handler Handle) error {
	routingKey := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	q, _ := m.queues.LoadOrStore(routingKey, newMemQueue(routingKey))
	q.(*memQueue).subscribe(ctx, handler)
	return nil
}

func (m *memoryBroker) PublishEvent(data string) error {
	routingKey := fmt.Sprintf("%s.*", eventRoutingKey)
	m.topics.Range(func(key any, value any) bool {
		if wildcard.Match(routingKey, key.(string)) {
			value.(*memTopic).publish(data)
		}
		return true
	})
	return nil
}

func (m *memoryBroker) SubscribeEvent(ctx context.Context, handler Handle) error {
	routingKey := fmt.Sprintf("%s.%s", eventRoutingKey, ksuid.New().String())
	t, _ := m.topics.LoadOrStore(routingKey, newMemTopic(routingKey))
	t.(*memTopic).subscribe(ctx, handler)
	return nil
}

func (m *memoryBroker) PublishManager(node string, data string) error {
	m.topics.Range(func(key any, value any) bool {
		if wildcard.Match(fmt.Sprintf("%s.%s", managerRoutingKey, node), key.(string)) {
			value.(*memTopic).publish(data)
		}
		return true
	})
	return nil
}

func (m *memoryBroker) SubscribeManager(ctx context.Context, node string, handler Handle) error {
	qname := fmt.Sprintf("%s.%s", managerRoutingKey, node)
	t, _ := m.topics.LoadOrStore(qname, newMemTopic(qname))
	t.(*memTopic).subscribe(ctx, handler)
	return nil
}

func (m *memoryBroker) Shutdown(ctx context.Context) {
	if !m.terminate.CompareAndSwap(false, true) {
		return
	}

	var wg sync.WaitGroup
	m.queues.Range(func(_, value any) bool {
		wg.Add(1)
		go func(q *memQueue) {
			defer wg.Done()
			q.close()
		}(value.(*memQueue))
		return true
	})
	m.topics.Range(func(_, value any) bool {
		wg.Add(1)
		go func(t *memTopic) {
			defer wg.Done()
			t.close()
		}(value.(*memTopic))
		return true
	})

	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-ctx.Done():
		logx.Infoln("shutting down broker")
	case <-doneChan:
		logx.Infoln("broker shutdown complete")
	}
}
