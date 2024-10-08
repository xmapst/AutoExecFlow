package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker/queue"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/pkg/tunny"
)

var qname = fmt.Sprintf("%s-worker-%s", utils.ServiceName, utils.HostName())

var (
	pool      = tunny.NewCallback(1)
	workQueue = queue.NewInMemoryBroker()
)

func init() {
	workQueue.Subscribe(context.Background(), qname, func(m any) {
		name, ok := m.(string)
		if !ok {
			return
		}
		t := newTask(name)
		if t == nil {
			return
		}
		pool.Submit(t.run)
	})
}

func SetSize(n int) {
	pool.SetSize(n)
}

func GetSize() int {
	return pool.GetSize()
}

func StopWait() {
	logx.Info("Waiting for all tasks to complete")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	workQueue.Shutdown(ctx)
}

func Submit(taskName string) error {
	return workQueue.Publish(qname, taskName)
}
