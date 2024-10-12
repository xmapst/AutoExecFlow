package queues

import (
	"context"
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"github.com/xmapst/go-rabbitmq"
	"go.uber.org/zap"

	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

var (
	directExchangeName = utils.ServiceName
	topicExchangeName  = utils.ServiceName + "Topic"
)

type amqpBroker struct {
	conn *rabbitmq.Conn
	// 防止相同生产者重复创建
	publisherMap map[string]*rabbitmq.Publisher
	// 防止相同消费者重复创建
	consumerMap map[string]*rabbitmq.Consumer
	topics      sync.Map
	directs     sync.Map
	mu          sync.Mutex
}

func newAmqpBroker(rawURL string) (*amqpBroker, error) {
	r := &amqpBroker{
		publisherMap: make(map[string]*rabbitmq.Publisher),
		consumerMap:  make(map[string]*rabbitmq.Consumer),
	}
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
	err = r.newDirectPublisher()
	if err != nil {
		return nil, err
	}
	err = r.newTopicPublisher()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *amqpBroker) PublishTask(node string, data string) error {
	rkey := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	return r.publishDirect(rkey, data)
}

func (r *amqpBroker) SubscribeTask(ctx context.Context, node string, handler Handle) error {
	qname := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	d, _ := r.directs.LoadOrStore(qname, newMemDirect(qname))
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.consumerMap[qname]; !ok {
		var err error
		r.consumerMap[qname], err = r.newDirectConsumer(qname, func(data string) error {
			d.(*memDirect).publish(data)
			return nil
		})
		if err != nil {
			return err
		}
	}
	d.(*memDirect).subscribe(ctx, handler)
	return nil
}

func (r *amqpBroker) PublishEvent(data string) error {
	rkey := fmt.Sprintf("%s.*", eventRoutingKey)
	return r.publishTopic(rkey, data)
}

func (r *amqpBroker) SubscribeEvent(ctx context.Context, handler Handle) error {
	rkey := fmt.Sprintf("%s.*", eventRoutingKey)
	t, _ := r.topics.LoadOrStore(rkey, newMemTopic(rkey))
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.consumerMap[rkey]; !ok {
		qname := fmt.Sprintf("%s_%s", eventRoutingKey, utils.HostName())
		var err error
		r.consumerMap[rkey], err = r.newTopicConsumer(rkey, qname, func(data string) error {
			t.(*memTopic).publish(data)
			return nil
		})
		if err != nil {
			return err
		}
	}
	t.(*memTopic).subscribe(ctx, handler)
	return nil
}

func (r *amqpBroker) PublishManager(node string, data string) error {
	routingKey := fmt.Sprintf("%s.%s", managerRoutingKey, node)
	return r.publishTopic(routingKey, data)
}

func (r *amqpBroker) SubscribeManager(ctx context.Context, node string, handler Handle) error {
	rkey := fmt.Sprintf("%s.%s", managerRoutingKey, node)
	t, _ := r.topics.LoadOrStore(rkey, newMemTopic(rkey))
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.consumerMap[rkey]; !ok {
		qname := fmt.Sprintf("%s_%s", managerRoutingKey, node)
		var err error
		r.consumerMap[rkey], err = r.newTopicConsumer(rkey, qname, func(data string) error {
			t.(*memTopic).publish(data)
			return nil
		})
		if err != nil {
			return err
		}
	}
	t.(*memTopic).subscribe(ctx, handler)
	return nil
}

func (r *amqpBroker) Shutdown(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, publisher := range r.publisherMap {
		publisher.Close()
	}
	for _, consumer := range r.consumerMap {
		consumer.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
	var wg sync.WaitGroup
	r.directs.Range(func(_, value any) bool {
		wg.Add(1)
		go func(t *memDirect) {
			defer wg.Done()
			t.close()
		}(value.(*memDirect))
		return true
	})
	r.topics.Range(func(_, value any) bool {
		wg.Add(1)
		go func(d *memTopic) {
			defer wg.Done()
			d.close()
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
		logx.Infoln("shutting down broker")
	case <-doneChan:
		logx.Infoln("broker shutdown complete")
	}
}

func (r *amqpBroker) subscribe(ctx context.Context, consumer *rabbitmq.Consumer, handler Handle) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
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

func (r *amqpBroker) publishDirect(rkey, data string) error {
	return r.publish(directExchangeName, rkey, data)
}

func (r *amqpBroker) publishTopic(rkey, data string) error {
	return r.publish(topicExchangeName, rkey, data)
}

func (r *amqpBroker) publish(ename, rkey, data string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	publisher, ok := r.publisherMap[ename]
	if !ok {
		return fmt.Errorf("exchange %s publisher not found", ename)
	}
	logx.Infof("published message to exchange %s routingKey %s", ename, rkey)
	return publisher.Publish(
		[]byte(data), []string{rkey},
		rabbitmq.WithPublishOptionsExchange(ename),    // 交换机名称
		rabbitmq.WithPublishOptionsMandatory,          // 强制发布
		rabbitmq.WithPublishOptionsPersistentDelivery, // 立即发布
	)
}

func (r *amqpBroker) newDirectPublisher() (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.publisherMap[directExchangeName], err = r.newPublisher(amqp091.ExchangeDirect, directExchangeName)
	return
}

func (r *amqpBroker) newTopicPublisher() (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.publisherMap[topicExchangeName], err = r.newPublisher(amqp091.ExchangeTopic, topicExchangeName)
	return
}

func (r *amqpBroker) newPublisher(kind, ename string) (*rabbitmq.Publisher, error) {
	return rabbitmq.NewPublisher(
		r.conn,
		rabbitmq.WithPublisherOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithPublisherOptionsExchangeName(ename),                                        // 交换机名称
		rabbitmq.WithPublisherOptionsExchangeKind(kind),                                         // 交换机类型
		rabbitmq.WithPublisherOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithPublisherOptionsExchangeDurable,                                            // 交换机持久化
	)
}

func (r *amqpBroker) newDirectConsumer(qname string, handle Handle) (*rabbitmq.Consumer, error) {
	return r.newConsumer(amqp091.ExchangeDirect, directExchangeName, qname, qname, handle)
}

func (r *amqpBroker) newTopicConsumer(rkey, qname string, handle Handle) (*rabbitmq.Consumer, error) {
	return r.newConsumer(amqp091.ExchangeTopic, topicExchangeName, rkey, qname, handle)
}

func (r *amqpBroker) newConsumer(kind, ename, rkey, qname string, handle Handle) (*rabbitmq.Consumer, error) {
	logx.Debugln("Subscribe", ename, "queue", qname)
	consumer, err := rabbitmq.NewConsumer(
		r.conn, qname,
		rabbitmq.WithConsumerOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithConsumerOptionsExchangeName(ename),                                        // 交换机名称
		rabbitmq.WithConsumerOptionsRoutingKey(rkey),                                           // routing key
		rabbitmq.WithConsumerOptionsExchangeKind(kind),                                         // 交换机类型
		rabbitmq.WithConsumerOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithConsumerOptionsExchangeDurable,                                            // 交换机持久化
		rabbitmq.WithConsumerOptionsQueueDurable,                                               // 队列持久化
	)
	if err != nil {
		return nil, err
	}
	r.subscribe(context.Background(), consumer, func(data string) error {
		return handle(data)
	})
	return consumer, nil
}
