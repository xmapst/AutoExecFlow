package mimecore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket/mimecore"
)

func TestQpDecode0(t *testing.T) {
	assert := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("mime.core", mimecore.Loader)
	assert.NoError(L.DoString(`return require 'mime.core'.unqp('=30')`))
	A := L.Get(-2)
	B := L.Get(-1)
	assert.Equal("0", A.String())
	assert.Equal(lua.LTNil, B.Type())
}

func TestQpDecode255(t *testing.T) {
	assert := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("mime.core", mimecore.Loader)
	assert.NoError(L.DoString(`return require 'mime.core'.unqp('=FF')`))
	A := L.Get(-2)
	B := L.Get(-1)
	assert.Equal("\xff", A.String())
	assert.Equal(lua.LTNil, B.Type())
}

func TestQpDecodeMoepsi(t *testing.T) {
	assert := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("mime.core", mimecore.Loader)
	assert.NoError(L.DoString(`return require 'mime.core'.unqp('M=C3=B6psi')`))
	A := L.Get(-2)
	B := L.Get(-1)
	assert.Equal("Möpsi", A.String())
	assert.Equal(lua.LTNil, B.Type())
}
