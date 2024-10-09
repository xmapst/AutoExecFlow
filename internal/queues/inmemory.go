package queues

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/xmapst/AutoExecFlow/internal/utils/wildcard"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

const defaultQueueSize = 1000

type memQueue struct {
	name    string
	ch      chan any
	subs    []*qsub
	unacked int32
	closed  atomic.Bool
	mu      sync.RWMutex // 使用读写锁来提高并发效率
	wg      sync.WaitGroup
}

func newMemQueue(name string) *memQueue {
	return &memQueue{
		name: name,
		ch:   make(chan any, defaultQueueSize),
		subs: make([]*qsub, 0),
	}
}

type qsub struct {
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan any
}

func (q *memQueue) send(m any) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.closed.Load() {
		return
	}
	select {
	case q.ch <- m:
	default:
		logx.Warnln("subscriber channel full, dropping message")
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
				return
			case m, ok := <-q.ch:
				if !ok {
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

func (t *memTopic) publish(event any) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, sub := range t.subs {
		select {
		case sub.ch <- event:
		case <-sub.ctx.Done():
			continue
		default:
			logx.Warnln("subscriber channel full, dropping message")
		}
	}
}

func (t *memTopic) subscribe(ctx context.Context, handler Handle) {
	t.mu.Lock()
	defer t.mu.Unlock()

	subCtx, cancel := context.WithCancel(ctx)
	sub := &qsub{
		ctx:    subCtx,
		cancel: cancel,
		ch:     make(chan any, 100),
	}
	t.subs = append(t.subs, sub)

	go func() {
		defer cancel()
		for {
			select {
			case <-sub.ctx.Done():
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

func (b *memoryBroker) Subscribe(ctx context.Context, class string, qname string, handler Handle) {
	switch class {
	case TYPE_DIRECT:
		q, _ := b.queues.LoadOrStore(qname, newMemQueue(qname))
		q.(*memQueue).subscribe(ctx, handler)
	case TYPE_TOPIC:
		t, _ := b.topics.LoadOrStore(qname, newMemTopic(qname))
		t.(*memTopic).subscribe(ctx, handler)
	}
}

func (b *memoryBroker) Publish(class string, qname string, m any) error {
	switch class {
	case TYPE_DIRECT:
		q, _ := b.queues.LoadOrStore(qname, newMemQueue(qname))
		q.(*memQueue).send(m)
	case TYPE_TOPIC:
		b.topics.Range(func(key any, value any) bool {
			if wildcard.Match(key.(string), qname) {
				value.(*memTopic).publish(m)
			}
			return true
		})
	}
	return nil
}

func (b *memoryBroker) Shutdown(ctx context.Context) {
	if !b.terminate.CompareAndSwap(false, true) {
		return
	}

	var wg sync.WaitGroup
	b.queues.Range(func(_, value any) bool {
		wg.Add(1)
		go func(q *memQueue) {
			defer wg.Done()
			q.close()
		}(value.(*memQueue))
		return true
	})
	b.topics.Range(func(_, value any) bool {
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
	case <-doneChan:
	}
}
