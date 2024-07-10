package base

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func Send(g *gin.Context, v interface{}) {
	switch g.NegotiateFormat(binding.MIMEJSON, binding.MIMEYAML, binding.MIMEYAML2) {
	case binding.MIMEJSON:
		g.JSON(http.StatusOK, v)
	case binding.MIMEYAML, binding.MIMEYAML2:
		g.YAML(http.StatusOK, v)
	default:
		g.JSON(http.StatusOK, v)
	}
}
