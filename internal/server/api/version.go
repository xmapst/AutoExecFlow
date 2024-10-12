package api

import (
	"runtime"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/pkg/info"
	"github.com/xmapst/AutoExecFlow/types"
)

// version
// @Summary		Version
// @description	Get server version
// @Tags		Default
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Success		200 {object} types.Base[types.Version]
// @Failure		500 {object} types.Base[any]
// @Router		/version [get]
func version(c *gin.Context) {
	base.Send(c, base.WithData(&types.Version{
		Version:   info.Version,
		BuildTime: info.BuildTime,
		Git: types.VersionGit{
			URL:    info.GitUrl,
			Branch: info.GitBranch,
			Commit: info.GitCommit,
		},
		Go: types.VersionGO{
			Version: runtime.Version(),
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
		User: types.VersionUser{
			Name:  info.UserName,
			Email: info.UserEmail,
		},
	}))
}
