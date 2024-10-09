package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

const defaultQueueSize = 1000

// memoryBroker a very simple implementation of the Broker interface
// which uses in-memory channels to exchange messages. Meant for local
// development, tests etc.
type memoryBroker struct {
	queues    sync.Map
	terminate atomic.Bool
}

type queue struct {
	name    string
	ch      chan any
	subs    []*qsub
	unacked int32
	closed  bool
	mu      sync.Mutex
	wg      sync.WaitGroup
}

func newQueue(name string) *queue {
	return &queue{
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

func (q *queue) send(m any) {
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

func (q *queue) size() int {
	return len(q.ch)
}

func (q *queue) close() {
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

func (q *queue) subscribe(ctx context.Context, sub SubHandle) {
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

func inMemoryBroker() *memoryBroker {
	return &memoryBroker{}
}

func (b *memoryBroker) Subscribe(ctx context.Context, qname string, handler SubHandle) {
	logx.Debugf("subscribing to queue %s", qname)
	q, _ := b.queues.LoadOrStore(qname, newQueue(qname))
	qq, _ := q.(*queue)
	qq.subscribe(ctx, handler)
}

func (b *memoryBroker) Publish(qname string, m any) error {
	logx.Debugf("publishing to queue %s", qname)
	q, _ := b.queues.LoadOrStore(qname, newQueue(qname))
	qq, ok := q.(*queue)
	if !ok {
		return fmt.Errorf("queue %s does not exist", qname)
	}
	qq.send(m)
	return nil
}

func (b *memoryBroker) Shutdown(ctx context.Context) {
	if !b.terminate.CompareAndSwap(false, true) {
		return
	}
	var wg sync.WaitGroup

	b.queues.Range(func(_, value any) bool {
		wg.Add(1)
		go func(q *queue) {
			logx.Debugf("shutting down channel %s", q.name)
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				q.close()
			}
		}(value.(*queue))
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
