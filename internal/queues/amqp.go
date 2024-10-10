package queues

import (
	"context"
	"errors"

	"github.com/rabbitmq/amqp091-go"
	"github.com/xmapst/go-rabbitmq"
	"go.uber.org/zap"

	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

var (
	queueExchangeName = utils.ServiceName
	topicExchangeName = utils.ServiceName + "Topic"
)

type amqpBroker struct {
	conn     *rabbitmq.Conn
	qPublish *rabbitmq.Publisher
	tPublish *rabbitmq.Publisher
}

func newAmqpBroker(rawURL string) (*amqpBroker, error) {
	r := new(amqpBroker)
	var err error
	r.conn, err = rabbitmq.NewConn(
		rawURL,
		rabbitmq.WithConnectionOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))),
		rabbitmq.WithConnectionOptionsConfig(rabbitmq.Config{
			Properties: amqp091.Table{
				"connection_name": utils.HostName(),
				"platform":        "Golang",
				"version":         "0.9.1",
				"product":         "RabbitMQ",
			},
		}),
	)
	if err != nil {
		return nil, err
	}
	r.qPublish, err = rabbitmq.NewPublisher(
		r.conn,
		rabbitmq.WithPublisherOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithPublisherOptionsExchangeName(queueExchangeName),                            // 交换机名称
		rabbitmq.WithPublisherOptionsExchangeKind(amqp091.ExchangeDirect),                       // 交换机类型
		rabbitmq.WithPublisherOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithPublisherOptionsExchangeDurable,                                            // 交换机持久化
	)
	if err != nil {
		return nil, err
	}
	r.tPublish, err = rabbitmq.NewPublisher(
		r.conn,
		rabbitmq.WithPublisherOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithPublisherOptionsExchangeName(topicExchangeName),                            // 交换机名称
		rabbitmq.WithPublisherOptionsExchangeKind(amqp091.ExchangeTopic),                        // 交换机类型
		rabbitmq.WithPublisherOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithPublisherOptionsExchangeDurable,                                            // 交换机持久化
	)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *amqpBroker) Shutdown(ctx context.Context) {
	if r.qPublish != nil {
		r.qPublish.Close()
	}
	if r.tPublish != nil {
		r.tPublish.Close()
	}
	if r.conn == nil {
		return
	}
	_ = r.conn.Close()
}

func (r *amqpBroker) Subscribe(ctx context.Context, class string, qname string, handler Handle) error {
	switch class {
	case TYPE_DIRECT:
		return r.subscribe(ctx, class, queueExchangeName, qname, handler)
	case TYPE_TOPIC:
		return r.subscribe(ctx, class, topicExchangeName, qname, handler)
	default:
		return errors.New("unknown class")
	}
}

func (r *amqpBroker) subscribe(ctx context.Context, class string, name, routingKey string, handler Handle) error {
	logx.Debugln("Subscribe", name, "queue", routingKey)
	consumer, err := rabbitmq.NewConsumer(
		r.conn, routingKey,
		rabbitmq.WithConsumerOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithConsumerOptionsExchangeName(name),                                         // 交换机名称
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),                                     // routing key
		rabbitmq.WithConsumerOptionsExchangeKind(class),                                        // 交换机类型
		rabbitmq.WithConsumerOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithConsumerOptionsExchangeDurable,                                            // 交换机持久化
	)
	if err != nil {
		return err
	}
	go func() {
		err = consumer.Run(func(d rabbitmq.Delivery) (action rabbitmq.Action) {
			err = handler(string(d.Body))
			if err != nil {
				logx.Errorln("unexpected error occurred while processing task", err)
				return rabbitmq.NackDiscard
			}
			return rabbitmq.Ack
		})
		if err != nil {
			logx.Errorln("unexpected error occurred while processing task", err)
		}
	}()
	go func() {
		<-ctx.Done()
		logx.Infof("subscribe %s closed", routingKey)
		consumer.Close()
	}()
	return nil
}

func (r *amqpBroker) Publish(class string, qname string, data string) error {
	switch class {
	case TYPE_DIRECT:
		return r.publish(r.qPublish, queueExchangeName, qname, data)
	case TYPE_TOPIC:
		return r.publish(r.tPublish, topicExchangeName, qname, data)
	default:
		return errors.New("unknown class")
	}
}

func (r *amqpBroker) publish(pub *rabbitmq.Publisher, name, routingKey string, data string) error {
	logx.Debugln("Publish", name, "queue", routingKey)
	return pub.Publish(
		[]byte(data), []string{routingKey},
		rabbitmq.WithPublishOptionsExchange(name),     // 交换机名称
		rabbitmq.WithPublishOptionsMandatory,          // 强制发布
		rabbitmq.WithPublishOptionsPersistentDelivery, // 立即发布
	)
}
