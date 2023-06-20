package handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	info "github.com/xmapst/osreapi"
)

type Info struct {
	Version   string   `json:",omitempty"`
	Go        InfoGO   `json:",omitempty"`
	Git       InfoGit  `json:",omitempty"`
	User      InfoUser `json:",omitempty"`
	BuildTime string   `json:",omitempty"`
}

type InfoGO struct {
	Version string `json:",omitempty"`
	OS      string `json:",omitempty"`
	Arch    string `json:",omitempty"`
}

type InfoGit struct {
	Url    string `json:",omitempty"`
	Branch string `json:",omitempty"`
	Commit string `json:",omitempty"`
}

type InfoUser struct {
	Name  string `json:",omitempty"`
	Email string `json:",omitempty"`
}

// Version
// @Summary Version
// @description 当前服务器版本
// @Tags Info
// @Success 200 {object} Info
// @Failure 500 {object} JSONResult
// @Router /version [get]
func Version(c *gin.Context) {
	c.JSON(http.StatusOK, Info{
		Version: info.Version,
		Go: InfoGO{
			Version: runtime.Version(),
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
		Git: InfoGit{
			Url:    info.GitUrl,
			Branch: info.GitBranch,
			Commit: info.GitCommit,
		},
		User: InfoUser{
			Name:  info.UserName,
			Email: info.UserEmail,
		},
		BuildTime: info.BuildTime,
	})
}
