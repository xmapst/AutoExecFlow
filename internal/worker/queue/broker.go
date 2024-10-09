package queue

import "context"

type SubHandle func(m any) error

const (
	BROKER_INMEMORY = "inmemory"
)

type Broker interface {
	// queue management
	Publish(qname string, m any) error
	Subscribe(ctx context.Context, qname string, handler SubHandle)
	Shutdown(ctx context.Context)
}

func New(name string) (Broker, error) {
	switch name {
	default:
		return inMemoryBroker(), nil
	}
}
