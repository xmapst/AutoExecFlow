package queues

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/utils"
)

var (
	broker            Broker
	taskRoutingKey    = utils.ServiceName + "Task"
	eventRoutingKey   = utils.ServiceName + "Event"
	managerRoutingKey = utils.ServiceName + "Manager"
)

type Handle func(data string) error

const (
	BROKER_INMEMORY = "inmemory"
	BROKER_AMQP     = "amqp"
)

type Broker interface {
	PublishEvent(data string) error
	PublishTask(node string, data string) error
	PublishManager(node string, data string) error

	SubscribeEvent(ctx context.Context, handler Handle) error
	SubscribeTask(ctx context.Context, node string, handler Handle) error
	SubscribeManager(ctx context.Context, node string, handler Handle) error

	Shutdown(ctx context.Context)
}

func New(rawURL string) error {
	before, _, found := strings.Cut(rawURL, "://")
	if !found {
		return errors.New("invalid message queue url")
	}
	var err error
	switch before {
	case BROKER_AMQP:
		broker, err = newAmqpBroker(rawURL)
		return err
	case BROKER_INMEMORY:
		broker = newInMemoryBroker()
		return nil
	default:
		return errors.New("unknown broker type")
	}
}

func PublishTask(name string, data string) error {
	return broker.PublishTask(name, data)
}

func SubscribeTask(ctx context.Context, name string, handler Handle) error {
	return broker.SubscribeTask(ctx, name, handler)
}

func PublishEvent(data string) error {
	return broker.PublishEvent(data)
}
func SubscribeEvent(ctx context.Context, handler Handle) error {
	return broker.SubscribeEvent(ctx, handler)
}
func PublishManager(node string, data string) error {
	return broker.PublishManager(node, data)
}
func SubscribeManager(ctx context.Context, node string, handler Handle) error {
	return broker.SubscribeManager(ctx, node, handler)
}

func Shutdown(ctx context.Context) {
	broker.Shutdown(ctx)
}
