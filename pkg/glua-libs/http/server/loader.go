// Package http_server implements golang package http_server utility functionality for lua.

package http

import (
	lua "github.com/yuin/gopher-lua"
)

// Preload adds http_server to the given Lua state's package.preload table. After it
// has been preloaded, it can be loaded using require:
//
//	local http_server = require("http_server")
func Preload(L *lua.LState) {
	L.PreloadModule("http_server", Loader)
}

// Loader is the module loader function.
func Loader(L *lua.LState) int {

	httpServerResponseWriterUd := L.NewTypeMetatable(`http_server_response_writer_ud`)
	L.SetGlobal(`http_server_response_writer_ud`, httpServerResponseWriterUd)
	L.SetField(httpServerResponseWriterUd, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"code":     HeaderCode,
		"header":   Header,
		"write":    Write,
		"redirect": Redirect,
		"done":     Done,
	}))

	httpServerUd := L.NewTypeMetatable(`http_server_ud`)
	L.SetGlobal(`http_server_ud`, httpServerUd)
	L.SetField(httpServerUd, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"accept":             Accept,
		"addr":               Addr,
		"do_handle_file":     HandleFile,
		"do_handle_string":   HandleString,
		"do_handle_function": HandleFunction,
	}))

	t := L.NewTable()
	L.SetFuncs(t, api)
	L.Push(t)
	return 1
}

var api = map[string]lua.LGFunction{
	"server":       New,
	"serve_static": ServeStaticFiles,
}
