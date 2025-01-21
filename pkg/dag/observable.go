package dag

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	eventChan  = make(chan string, 15)
	observable = newEvent(eventChan)
)

func emitEvent(format string, args ...interface{}) {
	eventChan <- fmt.Sprintf(format, args...)
}

func SubscribeEvent() (EventStream, int64, error) {
	return observable.Subscribe()
}

func UnSubscribeEvent(id int64) {
	observable.UnSubscribe(id)
}

// EventStream 是一个只接收的通道，表示事件流
type EventStream <-chan string

// Emitter 负责管理订阅者和广播事件
type emitter struct {
	eventStream EventStream
	subscribers sync.Map // 使用sync.Map替代map
	done        bool
	mux         sync.Mutex
	nextID      int64 // 用于生成订阅者的唯一ID
}

// process 处理事件流并将事件发送给所有订阅者
func (e *emitter) process() {
	for event := range e.eventStream {
		e.subscribers.Range(func(_, value any) bool {
			sub := value.(*subscriber)
			sub.sendEvent(event)
			return true
		})
	}
	e.close()
}

// close 关闭所有的订阅者
func (e *emitter) close() {
	e.mux.Lock()
	defer e.mux.Unlock()

	if e.done {
		return
	}
	e.done = true

	// 关闭所有的订阅者
	e.subscribers.Range(func(_, value any) bool {
		sub := value.(*subscriber)
		sub.close()
		return true
	})
}

// Subscribe 创建一个新的订阅者并返回其事件流
func (e *emitter) Subscribe() (EventStream, int64, error) {
	e.mux.Lock()
	defer e.mux.Unlock()

	if e.done {
		return nil, 0, errors.New("emitter has been closed, cannot subscribe")
	}
	// 生成唯一ID
	id := atomic.AddInt64(&e.nextID, 1)
	sub := newEventSubscriber()
	e.subscribers.Store(id, sub)
	return sub.eventChannel(), id, nil
}

// UnSubscribe 移除指定的订阅者并关闭其通道
func (e *emitter) UnSubscribe(id int64) {
	value, exist := e.subscribers.Load(id)
	if !exist {
		return
	}

	sub := value.(*subscriber)
	e.subscribers.Delete(id)
	sub.close()
}

// New 创建一个新的事件发射器
func newEvent(eventStream EventStream) *emitter {
	e := &emitter{
		eventStream: eventStream,
	}
	go e.process()
	return e
}

// subscriber 表示一个事件订阅者
type subscriber struct {
	buffer chan string
	once   sync.Once
}

// SendEvent 将事件发送到订阅者的通道
func (s *subscriber) sendEvent(event string) {
	// 非阻塞发送数据到通道
	select {
	case s.buffer <- event:
	default:
		// 丢弃或处理缓冲区满的情况
	}
}

// eventChannel 返回订阅者的事件通道
func (s *subscriber) eventChannel() EventStream {
	return s.buffer
}

// Close 关闭订阅者的事件通道
func (s *subscriber) close() {
	s.once.Do(func() {
		close(s.buffer)
	})
}

// newEventSubscriber 创建一个新的订阅者
func newEventSubscriber() *subscriber {
	return &subscriber{
		buffer: make(chan string, 200), // 缓冲区大小可以调整为参数
	}
}
