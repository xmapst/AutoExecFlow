package glualibs

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/plugin"
)

// Preload preload all gopher lua packages
func Preload(L *lua.LState) {
	plugin.PreloadAll(L)
}
