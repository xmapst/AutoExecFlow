package mimecore

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"
)

func TestQpQuote0(t *testing.T) {
	assert := assert.New(t)
	var buffer bytes.Buffer
	qpquote('0', &buffer)
	assert.Equal("=30", buffer.String())
}

func TestQpQuote255(t *testing.T) {
	assert := assert.New(t)
	var buffer bytes.Buffer
	qpquote(0xff, &buffer)
	assert.Equal("=FF", buffer.String())
}

// -----------------------------------------------------------------------------
// mime=require 'mime'
// A,B = mime.qp('Möpsi')
// assert.equal('M=C3=B6psi', A)
// assert.equal(nil, B)
// -----------------------------------------------------------------------------
func TestQpWithMoepsi(t *testing.T) {
	assert := assert.New(t)
	qpsetup()

	L := lua.NewState()
	defer L.Close()

	L.Push(lua.LString("Möpsi"))
	retargs := qpFn(L)

	assert.Equal(2, retargs)
	A := L.ToString(-2)
	B := L.Get(-1)
	assert.Equal("M=C3=B6psi", A)
	assert.Equal(lua.LTNil, B.Type())
}

// -----------------------------------------------------------------------------
// mime=require 'mime'
// A,B = mime.qp('Möpsi ', 'Pepsi')
// assert.equal('M=C3=B6psi Pepsi", A)
// assert.equal(”, B)
// -----------------------------------------------------------------------------
func TestQpWithMoepsiPepsi(t *testing.T) {
	assert := assert.New(t)
	qpsetup()

	L := lua.NewState()
	defer L.Close()

	L.Push(lua.LString("Möpsi "))
	L.Push(lua.LString("Pepsi"))
	retargs := qpFn(L)

	assert.Equal(2, retargs)
	A := L.ToString(-2)
	B := L.ToString(-1)
	assert.Equal("M=C3=B6psi Pepsi", A)
	assert.Equal("", B)
}

// -----------------------------------------------------------------------------
// mime=require 'mime'
// A,B = mime.qp('abc\r\ndef', nil, '_')
// assert.equal('abc_def', A)
// assert.equal(niul, B)
// -----------------------------------------------------------------------------
func TestQpMarker(t *testing.T) {
	assert := assert.New(t)
	qpsetup()

	L := lua.NewState()
	defer L.Close()

	L.Push(lua.LString("abc\r\ndef"))
	L.Push(lua.LNil)
	L.Push(lua.LString("_"))
	retargs := qpFn(L)

	assert.Equal(2, retargs)
	A := L.ToString(-2)
	B := L.Get(-1)
	assert.Equal("abc_def", A)
	assert.Equal(lua.LTNil, B.Type())
}
