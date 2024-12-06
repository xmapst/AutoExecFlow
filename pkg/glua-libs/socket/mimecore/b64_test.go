package mimecore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket/mimecore"
)

func TestB64WithDiegoPassword(t *testing.T) {
	assert := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("mime.core", mimecore.Loader)
	assert.NoError(L.DoString(`return require 'mime.core'.b64('diego:password')`))
	A := L.Get(-2)
	B := L.Get(-1)
	assert.Equal("ZGllZ286cGFzc3dvcmQ=", A.String())
	assert.Equal(lua.LTNil, B.Type())
}

func TestB64WithBinary(t *testing.T) {
	assert := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	L.PreloadModule("mime.core", mimecore.Loader)
	assert.NoError(L.DoString(`return require 'mime.core'.b64(string.char(0x00,0x44,0x1D,0x14,0x0F,0xF4,0xDA,0x11,0xA9,0x78,0x00,0x14,0x38,0x50,0x60,0xCE))`))
	A := L.Get(-2)
	B := L.Get(-1)
	assert.Equal("AEQdFA/02hGpeAAUOFBgzg==", A.String())
	assert.Equal(lua.LTNil, B.Type())
}
