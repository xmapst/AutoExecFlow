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

func (e *EventService) Subscribe(ctx context.Context, event chan string) error {
	return queues.SubscribeEvent(ctx, func(data string) error {
		event <- data
		return nil
	})
}
