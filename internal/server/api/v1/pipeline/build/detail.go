package build

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
// @Description 获取流水线指定构建任务详情
// @Tags 		构建
// @Accept		application/json
// @Produce		application/json
// @Param		pipeline path string true "流水线名称"
// @Param		build path string true "构建名称"
// @Success		200 {object} types.SBase[types.SPipelineBuildRes]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pipeline/{pipeline}/build/{build} [get]
func Detail(c *gin.Context) {
	pipelineName := c.Param("pipeline")
	if pipelineName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("pipeline does not exist")))
		return
	}
	buildName := c.Param("build")
	if buildName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("build does not exist")))
		return
	}
	code, res, err := service.Pipeline(pipelineName).BuildDetail(buildName)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, base.WithData(res).WithCode(code))
}
