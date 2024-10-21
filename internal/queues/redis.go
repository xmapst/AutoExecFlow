package queues

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisBroker struct {
	client *redis.Client
}

func newRedisBroker(rawURL string) (*redisBroker, error) {
	r := &redisBroker{}
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

func (r *redisBroker) PublishEvent(data string) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisBroker) PublishTask(node string, data string) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisBroker) PublishManager(node string, data string) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisBroker) SubscribeEvent(ctx context.Context, handler Handle) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisBroker) SubscribeTask(ctx context.Context, node string, handler Handle) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisBroker) SubscribeManager(ctx context.Context, node string, handler Handle) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisBroker) Shutdown(ctx context.Context) {
	//TODO implement me
	panic("implement me")
}
