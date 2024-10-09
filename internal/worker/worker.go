package worker

import (
	"context"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/queues"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/tunny"
)

var pool = tunny.NewCallback(1)

func Start() (err error) {
	queues.Subscribe(context.Background(), queues.TYPE_DIRECT, queues.TaskQueueName+utils.HostName(), func(m any) error {
		name, ok := m.(string)
		if !ok {
			return errors.New("invalid task name")
		}
		t := newTask(name)
		if t == nil {
			return errors.New("task not found")
		}
		return pool.Submit(t.run)
	})
	event, _, err := dag.SubscribeEvent()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case e := <-event:
				_ = queues.Publish(queues.TYPE_TOPIC, queues.EventQueueName, e)
			}
		}
	}()
	return
}

func SetSize(n int) {
	pool.SetSize(n)
}

func GetSize() int {
	return pool.GetSize()
}
