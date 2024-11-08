package step

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/types"
)

// Detail
// @Summary		详情
// @Description	获取步骤详情, 支持SSE订阅
// @Tags		步骤
// @Accept		application/json
// @Produce		application/json
// @Param		task path string true "任务名称"
// @Param		step path string true "步骤名称"
// @Success		200 {object} types.SBase[types.SStepRes]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/task/{task}/step/{step} [get]
func Detail(c *gin.Context) {
	var taskName = c.Param("task")
	if taskName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	var stepName = c.Param("step")
	if stepName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("step does not exist")))
		return
	}
	if c.GetHeader("Accept") != base.EventStreamMimeType {
		code, step, err := service.Step(taskName, stepName).Detail()
		if err != nil {
			base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
			return
		}

		base.Send(c, base.WithData(step).WithCode(code).WithError(fmt.Errorf(step.Message)))
		return
	}
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送心跳
	defer ticker.Stop()

	var lastCode types.Code
	var lastError error
	var last *types.SStepRes
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ticker.C:
			c.SSEvent("heartbeat", "keepalive")
			return true
		case <-c.Done():
			return false
		default:
			code, current, err := service.Step(taskName, stepName).Detail()
			if lastCode == code && errors.Is(err, lastError) && reflect.DeepEqual(last, current) {
				time.Sleep(1 * time.Second)
				return true
			}
			c.SSEvent("message", base.WithData(current).WithError(err).WithCode(code))
			lastCode = code
			lastError = err
			last = current
			time.Sleep(1 * time.Second)
			return true
		}
	})
}
