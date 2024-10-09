package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker/queue"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/pkg/tunny"
)

var qname = fmt.Sprintf("%s_Worker_%s", utils.ServiceName, utils.HostName())

var (
	pool      = tunny.NewCallback(1)
	workQueue queue.Broker
)

func Start() (err error) {
	workQueue, err = queue.New(queue.BROKER_INMEMORY)
	if err != nil {
		return err
	}
	workQueue.Subscribe(context.Background(), qname, func(m any) error {
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
	return
}

func SetSize(n int) {
	pool.SetSize(n)
}

func GetSize() int {
	return pool.GetSize()
}

func Shutdown() {
	logx.Info("waiting for all tasks to complete")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	workQueue.Shutdown(ctx)
}

func Submit(taskName string) error {
	return workQueue.Publish(qname, taskName)
}
