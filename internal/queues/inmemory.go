package queues

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/xmapst/AutoExecFlow/internal/wildcard"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

const defaultQueueSize = 1000

type memQueue struct {
	name    string
	ch      chan any
	subs    []*qsub
	unacked int32
	closed  bool
	mu      sync.Mutex
	wg      sync.WaitGroup
}

func newMemQueue(name string) *memQueue {
	return &memQueue{
		name:    name,
		ch:      make(chan any, defaultQueueSize),
		subs:    make([]*qsub, 0),
		unacked: 0,
	}
}

type qsub struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (q *memQueue) send(m any) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}
	select {
	case q.ch <- m:
	default:
		// Handle full queue channel scenario, maybe log or drop the message
		logx.Warningln("queue full")
	}
}

func (q *memQueue) size() int {
	return len(q.ch)
}

func (q *memQueue) close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return
	}

	// Mark the queue as closed and stop new messages
	q.closed = true
	close(q.ch)

	for _, sub := range q.subs {
		sub.cancel()
	}

	// Wait for all in-flight message processing to finish
	q.wg.Wait()
}

func (q *memQueue) subscribe(ctx context.Context, sub Handle) {
	q.mu.Lock()
	ctx, cancel := context.WithCancel(ctx)
	q.subs = append(q.subs, &qsub{ctx: ctx, cancel: cancel})
	q.mu.Unlock()

	// Increase waitgroup counter when a new subscriber is added
	q.wg.Add(1)

	go func() {
		defer q.wg.Done() // Mark as done when subscription finishes
		for {
			select {
			case <-ctx.Done():
				if q.size() == 0 {
					return
				}
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
	name       string
	ch         chan any
	subs       []Handle
	terminate  chan any
	terminated chan any
	mu         sync.RWMutex
}

func newMemTopic(name string) *memTopic {
	t := &memTopic{
		name:       name,
		ch:         make(chan any),
		terminate:  make(chan any),
		terminated: make(chan any),
	}
	go func() {
		for {
			select {
			case <-t.terminate:
				close(t.terminated)
				return
			case m := <-t.ch:
				t.mu.RLock()
				for _, sub := range t.subs {
					if err := sub(m); err != nil {
						logx.Errorln("unexpected error occurred while processing task", err)
					}
				}
				t.mu.RUnlock()
			}
		}
	}()

	return t
}

func (t *memTopic) publish(event any) {
	t.ch <- event
}

func (t *memTopic) subscribe(handler Handle) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.subs = append(t.subs, handler)
}

func (t *memTopic) close() {
	t.terminate <- 1
	<-t.terminated
}

// memoryBroker a very simple implementation of the Broker interface
// which uses in-memory channels to exchange messages. Meant for local
// development, tests etc.
type memoryBroker struct {
	queues    sync.Map
	topics    sync.Map
	terminate atomic.Bool
}

func newInMemoryBroker() *memoryBroker {
	return &memoryBroker{}
}

func (b *memoryBroker) Subscribe(ctx context.Context, class string, qname string, handler Handle) {
	logx.Debugf("subscribing to queue %s", qname)
	switch class {
	case TYPE_DIRECT:
		q, _ := b.queues.LoadOrStore(qname, newMemQueue(qname))
		qq, _ := q.(*memQueue)
		qq.subscribe(ctx, handler)
	case TYPE_TOPIC:
		t, _ := b.topics.LoadOrStore(qname, newMemTopic(qname))
		tt, _ := t.(*memTopic)
		tt.subscribe(handler)
	}
}

func (b *memoryBroker) Publish(class string, qname string, m any) error {
	logx.Debugf("publishing to queue %s", qname)
	switch class {
	case TYPE_DIRECT:
		q, _ := b.queues.LoadOrStore(qname, newMemQueue(qname))
		qq, _ := q.(*memQueue)
		qq.send(m)
	case TYPE_TOPIC:
		b.topics.Range(func(key any, value any) bool {
			name := key.(string)
			if wildcard.Match(name, qname) {
				tt, _ := value.(*memTopic)
				tt.publish(m)
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
			logx.Debugf("shutting down channel %s", q.name)
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				q.close()
			}
		}(value.(*memQueue))
		return true
	})
	b.topics.Range(func(_, value any) bool {
		wg.Add(1)
		go func(t *memTopic) {
			logx.Debugf("shutting down topic %s", t.name)
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				t.close()
			}
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
		return
	case <-doneChan:
		return
	}
}
