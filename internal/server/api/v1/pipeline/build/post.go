package build

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Post
// @Summary 	创建
// @Description 创建构建任务
// @Tags 		构建
// @Accept		application/json
// @Produce		application/json
// @Param		pipeline path string true "流水线名称"
// @Param		build body types.SPipelineBuildReq true "构建参数"
// @Success		200 {object} types.SBase[types.STaskCreateRes]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pipeline/{pipeline}/build [post]
func Post(c *gin.Context) {
	pipelineName := c.Param("pipeline")
	if pipelineName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("pipeline does not exist")))
		return
	}
	var req = new(types.SPipelineBuildReq)
	if err := c.ShouldBind(req); err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	build, err := service.Pipeline(pipelineName).BuildCreate(req)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	c.Request.Header.Set(types.XTaskName, build)
	c.Header(types.XTaskName, build)

	base.Send(c, base.WithData(&types.STaskCreateRes{
		Name: build,
	}))
}

// ReRun
// @Summary 	重新运行
// @Description 重新执行指定构建任务
// @Tags 		构建
// @Accept		application/json
// @Produce		application/json
// @Param		pipeline path string true "流水线名称"
// @Param		build path string true "构建名称"
// @Success		200 {object} types.SBase[any]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pipeline/{pipeline}/build/{build} [post]
func ReRun(c *gin.Context) {
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
	err := service.Pipeline(pipelineName).BuildReRun(buildName)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, base.WithCode[any](types.CodeSuccess))
}
