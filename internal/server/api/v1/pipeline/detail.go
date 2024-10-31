package pipeline

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Detail
// @Summary 	详情
// @Description 获取指定流水线详情
// @Tags 		流水线
// @Accept		application/json
// @Produce		application/json
// @Param		pipeline path string true "流水线名称"
// @Success		200 {object} types.SBase[types.SPipelineRes]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pipeline/{pipeline} [get]
func Detail(c *gin.Context) {
	pipelineName := c.Param("pipeline")
	if pipelineName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("pipeline does not exist")))
		return
	}
	res, err := service.Pipeline(pipelineName).Detail()
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, base.WithData(res).WithCode(types.CodeSuccess))
}
