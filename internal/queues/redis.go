package queues

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type sRedisBroker struct {
	client *redis.Client
}

func newRedisBroker(rawURL string) (*sRedisBroker, error) {
	r := &sRedisBroker{}
	opt, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}
	r.client = redis.NewClient(opt)
	if err = r.client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *sRedisBroker) PublishEvent(data string) error {
	//TODO implement me
	panic("implement me")
}

func (r *sRedisBroker) PublishTask(node string, data string) error {
	//TODO implement me
	panic("implement me")
}

func (r *sRedisBroker) PublishTaskDelayed(node string, data string, delay time.Duration) error {
	panic("implement me")
}

func (r *sRedisBroker) PublishManager(node string, data string) error {
	//TODO implement me
	panic("implement me")
}

func (r *sRedisBroker) SubscribeEvent(ctx context.Context, handler HandleFn) error {
	//TODO implement me
	panic("implement me")
}

func (r *sRedisBroker) SubscribeTask(ctx context.Context, node string, handler HandleFn) error {
	//TODO implement me
	panic("implement me")
}

func (r *sRedisBroker) SubscribeManager(ctx context.Context, node string, handler HandleFn) error {
	panic("implement me")
}

func (r *sRedisBroker) Shutdown(ctx context.Context) {
	//TODO implement me
	panic("implement me")
}
