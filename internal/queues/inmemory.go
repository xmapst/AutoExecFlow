package queues

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/segmentio/ksuid"
	"github.com/xmapst/logx"

	"github.com/xmapst/AutoExecFlow/pkg/wildcard"
)

const defaultQueueSize = 2 << 15

type sMemDirect struct {
	name    string
	ch      chan string
	subs    []*sSub
	unacked int32
	closed  atomic.Bool
	mu      sync.RWMutex
	wg      sync.WaitGroup
}

func newMemDirect(name string) *sMemDirect {
	return &sMemDirect{
		name: name,
		ch:   make(chan string, defaultQueueSize),
		subs: make([]*sSub, 0),
	}
}

type sSub struct {
	ctx    context.Context
	cancel context.CancelFunc
	cname  string

	// topic only
	ch chan string
}

// Publish messages to all subscribers in a non-blocking manner.
func (d *sMemDirect) publish(data string) {
	if d.closed.Load() {
		return
	}
	select {
	case d.ch <- data:
		logx.Infof("published message to subscriber direct queue %s", d.name)
	default:
		logx.Warnln("subscriber direct queue full, dropping message")
	}
}

func (d *sMemDirect) size() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.ch)
}

func (d *sMemDirect) close() {
	if !d.closed.CompareAndSwap(false, true) {
		return
	}
	close(d.ch)

	d.mu.Lock()
	for _, sub := range d.subs {
		sub.cancel()
	}
	d.mu.Unlock()

	d.wg.Wait()
}

func (d *sMemDirect) subscribe(ctx context.Context, handle HandleFn) {
	subCtx, cancel := context.WithCancel(ctx)
	sub := &sSub{
		cname:  ksuid.New().String(),
		ctx:    subCtx,
		cancel: cancel,
		ch:     make(chan string, 100),
	}
	logx.Infof("subscribed to direct queue %s cname %s", d.name, sub.cname)
	// Add subscription safely
	d.mu.Lock()
	d.subs = append(d.subs, sub)
	d.mu.Unlock()

	d.wg.Add(1)

	// Handle subscription in a separate goroutine
	go func() {
		defer d.wg.Done()
		defer cancel()

		for {
			select {
			case <-sub.ctx.Done():
				logx.Infof("subscriber %s cname %s closed", d.name, sub.cname)
				d.removeSubscriber(sub.cname)
				return
			case msg, ok := <-d.ch:
				if !ok {
					return
				}
				atomic.AddInt32(&d.unacked, 1)
				if err := handle(msg); err != nil {
					logx.Errorln("error processing direct queue:", err)
				}
				atomic.AddInt32(&d.unacked, -1)
			}
		}
	}()
}

func (d *sMemDirect) removeSubscriber(cname string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, sub := range d.subs {
		if sub.cname == cname {
			d.subs = append(d.subs[:i], d.subs[i+1:]...)
			break
		}
	}
}

type sMemTopic struct {
	name      string
	subs      []*sSub
	terminate chan struct{}
	mu        sync.RWMutex
}

func newMemTopic(name string) *sMemTopic {
	return &sMemTopic{
		name:      name,
		terminate: make(chan struct{}),
	}
}

func (t *sMemTopic) publish(event string) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, sub := range t.subs {
		select {
		case <-sub.ctx.Done():
			continue
		case sub.ch <- event:
			logx.Infof("published message to subscriber topic queue %s cname %s", t.name, sub.cname)
		default:
			logx.Warnln("subscriber topic queue full, dropping message")
		}
	}
}

func (t *sMemTopic) subscribe(ctx context.Context, handler HandleFn) {
	subCtx, cancel := context.WithCancel(ctx)
	sub := &sSub{
		cname:  ksuid.New().String(),
		ctx:    subCtx,
		cancel: cancel,
		ch:     make(chan string, 100),
	}
	logx.Infof("subscribed to topic queue %s cname %s", t.name, sub.cname)
	t.mu.Lock()
	t.subs = append(t.subs, sub)
	t.mu.Unlock()

	// Launch subscriber handling in a separate goroutine
	go func() {
		defer cancel()

		for {
			select {
			case <-sub.ctx.Done():
				logx.Infof("subscriber %s cname %s closed", t.name, sub.cname)
				t.removeSubscriber(sub.cname)
				return
			case m := <-sub.ch:
				if err := handler(m); err != nil {
					logx.Errorln("error processing topic queue:", err)
				}
			}
		}
	}()
}

func (t *sMemTopic) removeSubscriber(cname string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, sub := range t.subs {
		if sub.cname == cname {
			t.subs = append(t.subs[:i], t.subs[i+1:]...)
			break
		}
	}
}

func (t *sMemTopic) close() {
	close(t.terminate)

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, sub := range t.subs {
		sub.cancel()
	}
}

type sMemoryBroker struct {
	directs   sync.Map
	topics    sync.Map
	terminate atomic.Bool
}

func newInMemoryBroker() *sMemoryBroker {
	return &sMemoryBroker{}
}

func (m *sMemoryBroker) PublishTask(node string, data string) error {
	routingKey := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	d, _ := m.directs.LoadOrStore(routingKey, newMemDirect(routingKey))
	d.(*sMemDirect).publish(data)
	return nil
}

func (m *sMemoryBroker) SubscribeTask(ctx context.Context, node string, handler HandleFn) error {
	routingKey := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	d, _ := m.directs.LoadOrStore(routingKey, newMemDirect(routingKey))
	d.(*sMemDirect).subscribe(ctx, handler)
	return nil
}

func (m *sMemoryBroker) PublishEvent(data string) error {
	routingKey := fmt.Sprintf("%s.*", eventRoutingKey)
	m.topics.Range(func(key, value any) bool {
		if wildcard.Match(routingKey, key.(string)) {
			value.(*sMemTopic).publish(data)
		}
		return true
	})
	return nil
}

func (m *sMemoryBroker) SubscribeEvent(ctx context.Context, handler HandleFn) error {
	routingKey := fmt.Sprintf("%s.%s", eventRoutingKey, ksuid.New().String())
	t, _ := m.topics.LoadOrStore(routingKey, newMemTopic(routingKey))
	t.(*sMemTopic).subscribe(ctx, handler)
	return nil
}

func (m *sMemoryBroker) PublishManager(node string, data string) error {
	m.topics.Range(func(key, value any) bool {
		if wildcard.Match(fmt.Sprintf("%s.%s", managerRoutingKey, node), key.(string)) {
			value.(*sMemTopic).publish(data)
		}
		return true
	})
	return nil
}

func (m *sMemoryBroker) SubscribeManager(ctx context.Context, node string, handler HandleFn) error {
	qname := fmt.Sprintf("%s.%s", managerRoutingKey, node)
	t, _ := m.topics.LoadOrStore(qname, newMemTopic(qname))
	t.(*sMemTopic).subscribe(ctx, handler)
	return nil
}

func (m *sMemoryBroker) Shutdown(ctx context.Context) {
	if !m.terminate.CompareAndSwap(false, true) {
		return
	}

	var wg sync.WaitGroup
	m.directs.Range(func(_, value any) bool {
		wg.Add(1)
		go func(d *sMemDirect) {
			defer wg.Done()
			d.close()
		}(value.(*sMemDirect))
		return true
	})
	m.topics.Range(func(_, value any) bool {
		wg.Add(1)
		go func(t *sMemTopic) {
			defer wg.Done()
			t.close()
		}(value.(*sMemTopic))
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
