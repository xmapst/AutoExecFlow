package service

import (
	"context"

	"github.com/xmapst/AutoExecFlow/internal/queues"
)

type EventService struct {
}

func Event() *EventService {
	return &EventService{}
}

func (e *EventService) Subscribe(ctx context.Context, handle queues.Handle) error {
	queues.Subscribe(ctx, queues.TYPE_TOPIC, "event.*", handle)
	return nil
}
