package queues

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/ksuid"
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
	table := amqp091.NewConnectionProperties()
	table["connection_name"] = utils.HostName()
	r.conn, err = rabbitmq.NewConn(
		rawURL,
		rabbitmq.WithConnectionOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))),
		rabbitmq.WithConnectionOptionsConfig(rabbitmq.Config{
			Properties: table,
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

func (r *amqpBroker) PublishTask(node string, data string) error {
	routingKey := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	logx.Debugln("Publish", queueExchangeName, "queue", node)
	return r.qPublish.Publish(
		[]byte(data), []string{routingKey},
		rabbitmq.WithPublishOptionsExchange(queueExchangeName), // 交换机名称
		rabbitmq.WithPublishOptionsMandatory,                   // 强制发布
		rabbitmq.WithPublishOptionsPersistentDelivery,          // 立即发布
	)
}

func (r *amqpBroker) SubscribeTask(ctx context.Context, node string, handler Handle) error {
	routingKey := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	logx.Debugln("Subscribe", queueExchangeName, "queue", routingKey)
	consumer, err := rabbitmq.NewConsumer(
		r.conn, routingKey,
		rabbitmq.WithConsumerOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithConsumerOptionsExchangeName(queueExchangeName),                            // 交换机名称
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),                                     // routing key
		rabbitmq.WithConsumerOptionsExchangeKind(amqp091.ExchangeDirect),                       // 交换机类型
		rabbitmq.WithConsumerOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithConsumerOptionsExchangeDurable,                                            // 交换机持久化
		rabbitmq.WithConsumerOptionsQueueDurable,                                               // 队列持久化
	)
	if err != nil {
		return err
	}
	r.subscribe(ctx, consumer, handler)
	return nil
}

func (r *amqpBroker) PublishEvent(data string) error {
	routingKey := fmt.Sprintf("%s.*", eventRoutingKey)
	logx.Debugln("Publish", topicExchangeName, "queue", routingKey)
	return r.qPublish.Publish(
		[]byte(data), []string{routingKey},
		rabbitmq.WithPublishOptionsExchange(topicExchangeName), // 交换机名称
		rabbitmq.WithPublishOptionsMandatory,                   // 强制发布
		rabbitmq.WithPublishOptionsPersistentDelivery,          // 立即发布
	)
}

func (r *amqpBroker) SubscribeEvent(ctx context.Context, handler Handle) error {
	routingKey := fmt.Sprintf("%s.*", eventRoutingKey)
	queueName := fmt.Sprintf("%s_%s", eventRoutingKey, ksuid.New().String())
	logx.Debugln("Subscribe", topicExchangeName, "queue", queueName)
	consumer, err := rabbitmq.NewConsumer(
		r.conn, queueName,
		rabbitmq.WithConsumerOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithConsumerOptionsExchangeName(topicExchangeName),                            // 交换机名称
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),                                     // routing key
		rabbitmq.WithConsumerOptionsExchangeKind(amqp091.ExchangeTopic),                        // 交换机类型
		rabbitmq.WithConsumerOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithConsumerOptionsExchangeDurable,                                            // 交换机持久化
		rabbitmq.WithConsumerOptionsQueueAutoDelete,                                            // 队列自动删除
	)
	if err != nil {
		return err
	}
	r.subscribe(ctx, consumer, handler)
	return nil
}

func (r *amqpBroker) PublishManager(node string, data string) error {
	routingKey := fmt.Sprintf("%s.%s", managerRoutingKey, node)
	logx.Debugln("Publish", topicExchangeName, "queue", routingKey)
	return r.tPublish.Publish(
		[]byte(data), []string{routingKey},
		rabbitmq.WithPublishOptionsExchange(topicExchangeName), // 交换机名称
		rabbitmq.WithPublishOptionsMandatory,                   // 强制发布
		rabbitmq.WithPublishOptionsPersistentDelivery,          // 立即发布
	)
}

func (r *amqpBroker) SubscribeManager(ctx context.Context, node string, handler Handle) error {
	routingKey := fmt.Sprintf("%s.%s", managerRoutingKey, node)
	queueName := fmt.Sprintf("%s_%s", managerRoutingKey, node)
	logx.Debugln("Subscribe", topicExchangeName, "queue", queueName)
	consumer, err := rabbitmq.NewConsumer(
		r.conn, queueName,
		rabbitmq.WithConsumerOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithConsumerOptionsExchangeName(topicExchangeName),                            // 交换机名称
		rabbitmq.WithConsumerOptionsRoutingKey(routingKey),                                     // routing key
		rabbitmq.WithConsumerOptionsExchangeKind(amqp091.ExchangeTopic),                        // 交换机类型
		rabbitmq.WithConsumerOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithConsumerOptionsExchangeDurable,                                            // 交换机持久化
		rabbitmq.WithConsumerOptionsQueueAutoDelete,                                            // 队列自动删除
	)
	if err != nil {
		return err
	}
	r.subscribe(ctx, consumer, handler)
	return nil
}

func (r *amqpBroker) subscribe(ctx context.Context, consumer *rabbitmq.Consumer, handler Handle) {
	go func() {
		err := consumer.Run(func(d rabbitmq.Delivery) (action rabbitmq.Action) {
			err := handler(string(d.Body))
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
		logx.Infof("subscribe closed")
		consumer.Close()
	}()
}
