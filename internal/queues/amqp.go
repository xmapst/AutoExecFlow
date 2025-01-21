package queues

import (
	"context"
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"github.com/xmapst/go-rabbitmq"
	"github.com/xmapst/logx"
	"go.uber.org/zap"

	"github.com/xmapst/AutoExecFlow/internal/utils"
)

var (
	directExchangeName = utils.ServiceName
	topicExchangeName  = utils.ServiceName + "Topic"
)

type sAmqpBroker struct {
	nodeName string
	conn     *rabbitmq.Conn
	// 防止相同生产者重复创建
	publisherMap map[string]*rabbitmq.Publisher
	// 防止相同消费者重复创建
	consumerMap map[string]*rabbitmq.Consumer
	topics      sync.Map
	directs     sync.Map
	mu          sync.Mutex
}

func newAmqpBroker(nodeName, rawURL string) (*sAmqpBroker, error) {
	a := &sAmqpBroker{
		nodeName:     nodeName,
		publisherMap: make(map[string]*rabbitmq.Publisher),
		consumerMap:  make(map[string]*rabbitmq.Consumer),
	}
	var err error
	table := amqp091.NewConnectionProperties()
	table["connection_name"] = nodeName
	a.conn, err = rabbitmq.NewConn(
		rawURL,
		rabbitmq.WithConnectionOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))),
		rabbitmq.WithConnectionOptionsConfig(rabbitmq.Config{
			Properties: table,
		}),
	)
	if err != nil {
		return nil, err
	}
	err = a.newDirectPublisher()
	if err != nil {
		return nil, err
	}
	err = a.newTopicPublisher()
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *sAmqpBroker) PublishTask(node string, data string) error {
	rkey := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	return a.publishDirect(rkey, data)
}

func (a *sAmqpBroker) SubscribeTask(ctx context.Context, node string, handler HandleFn) error {
	qname := fmt.Sprintf("%s_%s", taskRoutingKey, node)
	d, _ := a.directs.LoadOrStore(qname, newMemDirect(qname))
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.consumerMap[qname]; !ok {
		var err error
		a.consumerMap[qname], err = a.newDirectConsumer(qname, func(data string) error {
			d.(*sMemDirect).publish(data)
			return nil
		})
		if err != nil {
			return err
		}
	}
	d.(*sMemDirect).subscribe(ctx, handler)
	return nil
}

func (a *sAmqpBroker) PublishEvent(data string) error {
	rkey := fmt.Sprintf("%s.*", eventRoutingKey)
	return a.publishTopic(rkey, data)
}

func (a *sAmqpBroker) SubscribeEvent(ctx context.Context, handler HandleFn) error {
	rkey := fmt.Sprintf("%s.*", eventRoutingKey)
	t, _ := a.topics.LoadOrStore(rkey, newMemTopic(rkey))
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.consumerMap[rkey]; !ok {
		qname := fmt.Sprintf("%s_%s", eventRoutingKey, a.nodeName)
		var err error
		a.consumerMap[rkey], err = a.newTopicConsumer(rkey, qname, func(data string) error {
			t.(*sMemTopic).publish(data)
			return nil
		})
		if err != nil {
			return err
		}
	}
	t.(*sMemTopic).subscribe(ctx, handler)
	return nil
}

func (a *sAmqpBroker) PublishManager(node string, data string) error {
	routingKey := fmt.Sprintf("%s.%s", managerRoutingKey, node)
	return a.publishTopic(routingKey, data)
}

func (a *sAmqpBroker) SubscribeManager(ctx context.Context, node string, handler HandleFn) error {
	rkey := fmt.Sprintf("%s.%s", managerRoutingKey, node)
	t, _ := a.topics.LoadOrStore(rkey, newMemTopic(rkey))
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.consumerMap[rkey]; !ok {
		qname := fmt.Sprintf("%s_%s", managerRoutingKey, node)
		var err error
		a.consumerMap[rkey], err = a.newTopicConsumer(rkey, qname, func(data string) error {
			t.(*sMemTopic).publish(data)
			return nil
		})
		if err != nil {
			return err
		}
	}
	t.(*sMemTopic).subscribe(ctx, handler)
	return nil
}

func (a *sAmqpBroker) Shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, publisher := range a.publisherMap {
		publisher.Close()
	}
	for _, consumer := range a.consumerMap {
		consumer.Close()
	}
	if a.conn != nil {
		_ = a.conn.Close()
	}
	var wg sync.WaitGroup
	a.directs.Range(func(_, value any) bool {
		wg.Add(1)
		go func(t *sMemDirect) {
			defer wg.Done()
			t.close()
		}(value.(*sMemDirect))
		return true
	})
	a.topics.Range(func(_, value any) bool {
		wg.Add(1)
		go func(d *sMemTopic) {
			defer wg.Done()
			d.close()
		}(value.(*sMemTopic))
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

func (a *sAmqpBroker) subscribe(ctx context.Context, consumer *rabbitmq.Consumer, handler HandleFn) {
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

func (a *sAmqpBroker) publishDirect(rkey, data string) error {
	return a.publish(directExchangeName, rkey, data)
}

func (a *sAmqpBroker) publishTopic(rkey, data string) error {
	return a.publish(topicExchangeName, rkey, data)
}

func (a *sAmqpBroker) publish(ename, rkey, data string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	publisher, ok := a.publisherMap[ename]
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

func (a *sAmqpBroker) newDirectPublisher() (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.publisherMap[directExchangeName], err = a.newPublisher(amqp091.ExchangeDirect, directExchangeName)
	return
}

func (a *sAmqpBroker) newTopicPublisher() (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.publisherMap[topicExchangeName], err = a.newPublisher(amqp091.ExchangeTopic, topicExchangeName)
	return
}

func (a *sAmqpBroker) newPublisher(kind, ename string) (*rabbitmq.Publisher, error) {
	return rabbitmq.NewPublisher(
		a.conn,
		rabbitmq.WithPublisherOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithPublisherOptionsExchangeName(ename),                                        // 交换机名称
		rabbitmq.WithPublisherOptionsExchangeKind(kind),                                         // 交换机类型
		rabbitmq.WithPublisherOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithPublisherOptionsExchangeDurable,                                            // 交换机持久化
	)
}

func (a *sAmqpBroker) newDirectConsumer(qname string, handle HandleFn) (*rabbitmq.Consumer, error) {
	return a.newConsumer(amqp091.ExchangeDirect, directExchangeName, qname, qname, handle)
}

func (a *sAmqpBroker) newTopicConsumer(rkey, qname string, handle HandleFn) (*rabbitmq.Consumer, error) {
	return a.newConsumer(amqp091.ExchangeTopic, topicExchangeName, rkey, qname, handle)
}

func (a *sAmqpBroker) newConsumer(kind, ename, rkey, qname string, handle HandleFn) (*rabbitmq.Consumer, error) {
	logx.Debugln("Subscribe", ename, "queue", qname)
	consumer, err := rabbitmq.NewConsumer(
		a.conn, qname,
		rabbitmq.WithConsumerOptionsLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))), // 日志
		rabbitmq.WithConsumerOptionsExchangeName(ename),                                        // 交换机名称
		rabbitmq.WithConsumerOptionsRoutingKey(rkey),                                           // routing key
		rabbitmq.WithConsumerOptionsExchangeKind(kind),                                         // 交换机类型
		rabbitmq.WithConsumerOptionsExchangeDeclare,                                            // 声明交换机
		rabbitmq.WithConsumerOptionsExchangeDurable,                                            // 交换机持久化
		rabbitmq.WithConsumerOptionsQueueDurable,                                               // 队列持久化
		rabbitmq.WithConsumerOptionsQueueQuorum,                                                // 使用仲裁队列
	)
	if err != nil {
		return nil, err
	}
	a.subscribe(context.Background(), consumer, func(data string) error {
		return handle(data)
	})
	return consumer, nil
}
