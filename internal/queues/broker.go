package queues

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/utils"
)

var (
	broker            IBroker
	taskRoutingKey    = utils.ServiceName + "Task"
	eventRoutingKey   = utils.ServiceName + "Event"
	managerRoutingKey = utils.ServiceName + "Manager"
)

type HandleFn func(data string) error

const (
	BrokerInMemory = "inmemory"
	BrokerAmqp     = "amqp"
	BrokerRedis    = "redis"
)

type IBroker interface {
	PublishEvent(data string) error
	PublishTask(node string, data string) error
	PublishManager(node string, data string) error

	SubscribeEvent(ctx context.Context, handler HandleFn) error
	SubscribeTask(ctx context.Context, node string, handler HandleFn) error
	SubscribeManager(ctx context.Context, node string, handler HandleFn) error

	Shutdown(ctx context.Context)
}

func New(nodeName, rawURL string) error {
	before, _, found := strings.Cut(rawURL, "://")
	if !found {
		return errors.New("invalid message queue url")
	}
	var err error
	switch before {
	case BrokerAmqp:
		broker, err = newAmqpBroker(nodeName, rawURL)
		return err
	case BrokerRedis:
		broker, err = newRedisBroker(rawURL)
		return err
	case BrokerInMemory:
		broker = newInMemoryBroker()
		return nil
	default:
		return errors.New("unknown broker type")
	}
}

func PublishTask(name string, data string) error {
	return broker.PublishTask(name, data)
}

func SubscribeTask(ctx context.Context, name string, handler HandleFn) error {
	return broker.SubscribeTask(ctx, name, handler)
}

func PublishEvent(data string) error {
	return broker.PublishEvent(data)
}
func SubscribeEvent(ctx context.Context, handler HandleFn) error {
	return broker.SubscribeEvent(ctx, handler)
}
func PublishManager(node string, data string) error {
	return broker.PublishManager(node, data)
}
func SubscribeManager(ctx context.Context, node string, handler HandleFn) error {
	return broker.SubscribeManager(ctx, node, handler)
}

func Shutdown(ctx context.Context) {
	broker.Shutdown(ctx)
}
