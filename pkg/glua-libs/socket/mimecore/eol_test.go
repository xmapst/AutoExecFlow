package mimecore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket/mimecore"
)

func TestEolDos(t *testing.T) {
	assert := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("mime.core", mimecore.Loader)
	assert.NoError(L.DoString(`return require 'mime.core'.eol(0, 'abc\r\n123', '\n')`))
	A := L.Get(-2)
	B := L.Get(-1)
	assert.Equal("abc\n123", A.String())
	assert.Equal(lua.LTNumber, B.Type())
	assert.Equal(lua.LNumber(0), L.ToNumber(-1))
}
