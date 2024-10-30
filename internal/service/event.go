package service

import (
	"context"

	"github.com/xmapst/AutoExecFlow/internal/queues"
)

type SEventService struct {
}

func Event() *SEventService {
	return &SEventService{}
}

func (e *SEventService) Subscribe(ctx context.Context, event chan string) error {
	return queues.SubscribeEvent(ctx, func(data string) error {
		event <- data
		return nil
	})
}
