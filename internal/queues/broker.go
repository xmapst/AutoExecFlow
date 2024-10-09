package queues

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/utils"
)

var (
	broker         Broker
	TaskQueueName  = fmt.Sprintf("%s_Worker_", utils.ServiceName)
	EventQueueName = fmt.Sprintf("%s_Event", utils.ServiceName)
)

type Handle func(m any) error

const (
	BROKER_INMEMORY = "inmemory"
	BROKER_RABBITMQ = "rabbitmq"

	TYPE_DIRECT = "direct"
	TYPE_TOPIC  = "topic"
)

type Broker interface {
	// queue management
	Publish(class string, qname string, m any) error
	Subscribe(ctx context.Context, class string, qname string, handler Handle)
	Shutdown(ctx context.Context)
}

func New(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case BROKER_INMEMORY:
		broker = newInMemoryBroker()
		return nil
	default:
		return errors.New("unknown broker type")
	}
}

func Publish(class string, name string, m any) error {
	return broker.Publish(class, name, m)
}

func Subscribe(ctx context.Context, class string, name string, handler Handle) {
	broker.Subscribe(ctx, class, name, handler)
}

func Shutdown(ctx context.Context) {
	broker.Shutdown(ctx)
}
