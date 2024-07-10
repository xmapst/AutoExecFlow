package router

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/middleware/zap"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/internal/api/v1/pool"
	"github.com/xmapst/AutoExecFlow/internal/api/v1/sys"
	"github.com/xmapst/AutoExecFlow/internal/api/v1/task"
	"github.com/xmapst/AutoExecFlow/internal/api/v1/task/step"
	"github.com/xmapst/AutoExecFlow/internal/api/v1/task/workspace"
	taskv2 "github.com/xmapst/AutoExecFlow/internal/api/v2/task"
	stepv2 "github.com/xmapst/AutoExecFlow/internal/api/v2/task/step"
	"github.com/xmapst/AutoExecFlow/pkg/info"
)

//go:embed docs/swagger.yaml
var swaggerFS embed.FS

// New
// @title			Auto Exec Flow
// @version			1.0
// @description		An `API` for cross-platform custom orchestration of execution steps
// @description		without any third-party dependencies. Based on `DAG`, it implements the scheduling
// @description		function of sequential execution of dependent steps and concurrent execution of
// @description		non-dependent steps. <br /><br /> It provides `API` remote operation mode, batch
// @description		execution of `Shell` , `Powershell` , `Python` and other commands, and easily
// @description		completes common management tasks such as running automated operation and maintenance
// @description		scripts, polling processes, installing or uninstalling software, updating applications,
// @description		and installing patches.
// @contact.name	AutoExecFlow
// @contact.url		https://github.com/xmapst/AutoExecFlow/issues
// @license.name	GPL-3.0
// @license.url		https://github.com/xmapst/AutoExecFlow/blob/main/LICENSE
func New(relativePath string) *gin.Engine {
	router := gin.New()
	router.Use(
		zap.Logger,
		zap.Recovery,
		cors.Default(),
		gzip.Gzip(gzip.DefaultCompression),
		func(c *gin.Context) {
			c.Header("Server", "Gin")
			c.Header("X-Server", "Gin")
			c.Header("X-Version", info.Version)
			c.Header("X-Powered-By", info.UserEmail)
		},
	)

	baseGroup := router.Group(relativePath)

	// debug pprof
	pprof.Register(baseGroup)

	// base
	baseGroup.GET("/version", version)
	baseGroup.GET("/healthyz", healthyz)
	baseGroup.GET("/heartbeat", heartbeat)
	baseGroup.HEAD("/heartbeat", heartbeat)
	baseGroup.GET("/swagger.yaml", func(c *gin.Context) {
		content, err := swaggerFS.ReadFile("docs/swagger.yaml")
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Data(http.StatusOK, binding.MIMEYAML2, content)
	})
	baseGroup.GET("/swagger", func(c *gin.Context) {
		var scheme = "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		if c.GetHeader("X-Forwarded-Proto") != "" {
			scheme = c.GetHeader("X-Forwarded-Proto")
		}
		var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, strings.TrimSuffix(relativePath, "/"))
		htmlContent := fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <title>AutoExecFlow API</title>
    <meta charset="utf-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script
      id="api-reference"
      data-configuration="{&quot;hideModels&quot;: true, &quot;pathRouting&quot;: &quot;&quot;, &quot;baseServerURL&quot;: &quot;%s&quot;}"
      type="application/yaml"
      data-url="%s"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest"></script>
  </body>
</html>`,
			uriPrefix,
			fmt.Sprintf("%s/swagger.yaml", uriPrefix),
		)
		_, _ = fmt.Fprintln(c.Writer, htmlContent)
	})
	api := baseGroup.Group("/api")
	// V1
	{
		// task
		api.GET("/v1/task", task.List)
		api.POST("/v1/task", task.Post)
		api.PUT("/v1/task/:task", task.Manager)
		// workspace
		api.GET("/v1/task/:task/workspace", workspace.Get)
		api.DELETE("/v1/task/:task/workspace", workspace.Delete)
		api.POST("/v1/task/:task/workspace", workspace.Post)
		// step
		api.GET("/v1/task/:task", step.List)
		api.GET("/v1/task/:task/step/:step", step.Log)
		api.PUT("/v1/task/:task/step/:step", step.Manager)
		// worker pool
		api.GET("/v1/pool", pool.Detail)
		api.POST("/v1/pool", pool.Post)

		// pty
		api.GET("/v1/pty", sys.PtyWs)
	}
	// V2
	{
		// task
		api.POST("/v2/task", taskv2.Post)
		api.GET("/v2/task/:task", taskv2.Detail)
		api.GET("/v2/task/:task/step/:step", stepv2.Detail)
	}

	// endpoints
	api.Any("/endpoints", func(c *gin.Context) {
		type Endpoint struct {
			Method string `json:"method" yaml:"Method"`
			Path   string `json:"path" yaml:"Path"`
		}
		var res []Endpoint
		var scheme = "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		if c.GetHeader("X-Forwarded-Proto") != "" {
			scheme = c.GetHeader("X-Forwarded-Proto")
		}
		var uriPrefix = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
		for _, v := range router.Routes() {
			res = append(res, Endpoint{
				Method: v.Method,
				Path:   fmt.Sprintf("%s%s", uriPrefix, v.Path),
			})
		}
		base.Send(c, res)
	})

	// no method
	router.NoMethod(func(c *gin.Context) {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("method not allowed")))
	})

	// no route
	router.NoRoute(func(c *gin.Context) {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("the requested path does not exist")))
	})
	return router
}
