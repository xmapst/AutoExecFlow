package mimecore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"
)

// -----------------------------------------------------------------------------
// mime=require 'mime'
// A,m = mime.qpwrp(0)
// assert.equal("=\r\n", A)
// assert.equal(76, m)
// -----------------------------------------------------------------------------
func TestQpwrp1(t *testing.T) {
	assert := assert.New(t)
	qpsetup()
	L := lua.NewState()
	defer L.Close()

	L.Push(lua.LNumber(0))

	nretargs := qpwrpFn(L)

	assert.Equal(2, nretargs)
	A := L.ToString(-2)
	m := L.ToNumber(-1)
	assert.Equal("=\r\n", A)
	assert.Equal(lua.LNumber(76), m)
}

// -----------------------------------------------------------------------------
// mime=require 'mime'
// A,m = mime.qpwrp(0,'a')
// assert.equal("=\r\na", A)
// assert.equal(75, m)
// -----------------------------------------------------------------------------
func TestQpwrp2(t *testing.T) {
	assert := assert.New(t)
	qpsetup()
	L := lua.NewState()
	defer L.Close()

	L.Push(lua.LNumber(0))
	L.Push(lua.LString("a"))

	nretargs := qpwrpFn(L)

	assert.Equal(2, nretargs)
	A := L.ToString(-2)
	m := L.ToNumber(-1)
	assert.Equal("=\r\na", A)
	assert.Equal(lua.LNumber(75), m)
}

// -----------------------------------------------------------------------------
// mime=require 'mime'
// A,m = mime.qpwrp(0,'abcdefg',4)
// assert.equal("=\r\nabc=\r\ndef=\r\ng", A)
// assert.equal(3, m)
// -----------------------------------------------------------------------------
func TestQpwrp3(t *testing.T) {
	assert := assert.New(t)
	qpsetup()
	L := lua.NewState()
	defer L.Close()

	L.Push(lua.LNumber(0))
	L.Push(lua.LString("abcdefg"))
	L.Push(lua.LNumber(4))

	nretargs := qpwrpFn(L)

	assert.Equal(2, nretargs)
	A := L.ToString(-2)
	m := L.ToNumber(-1)
	assert.Equal("=\r\nabc=\r\ndef=\r\ng", A)
	assert.Equal(lua.LNumber(3), m)
}

// -----------------------------------------------------------------------------
// mime=require 'mime'
// A,m = mime.qpwrp(3,'abcdefg',4)
// assert.equal("ab=\r\ncde=\r\nfg", A)
// assert.equal(2, m)
// -----------------------------------------------------------------------------
func TestQpwrp4(t *testing.T) {
	assert := assert.New(t)
	qpsetup()
	L := lua.NewState()
	defer L.Close()

	L.Push(lua.LNumber(3))
	L.Push(lua.LString("abcdefg"))
	L.Push(lua.LNumber(4))

	nretargs := qpwrpFn(L)

	assert.Equal(2, nretargs)
	A := L.ToString(-2)
	m := L.ToNumber(-1)
	assert.Equal("ab=\r\ncde=\r\nfg", A)
	assert.Equal(lua.LNumber(2), m)
}
