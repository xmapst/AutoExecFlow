package mimecore

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ZGllZ286cGFzc3dvcmQ=
func TestB64DecodeDiegoPassword(t *testing.T) {
	assert := assert.New(t)
	b64setup()
	var input, buffer bytes.Buffer

	b64decode('Z', &input, &buffer)
	assert.Equal(1, input.Len())
	assert.Equal(0, buffer.Len())

	b64decode('G', &input, &buffer)
	assert.Equal(2, input.Len())
	assert.Equal(0, buffer.Len())

	b64decode('l', &input, &buffer)
	assert.Equal(3, input.Len())
	assert.Equal(0, buffer.Len())

	b64decode('l', &input, &buffer)
	assert.Equal(0, input.Len())
	assert.Equal(3, buffer.Len())
	assert.Equal("die", buffer.String())

	// ----------

	b64decode('Z', &input, &buffer)
	assert.Equal(1, input.Len())
	assert.Equal(3, buffer.Len())

	b64decode('2', &input, &buffer)
	assert.Equal(2, input.Len())
	assert.Equal(3, buffer.Len())

	b64decode('8', &input, &buffer)
	assert.Equal(3, input.Len())
	assert.Equal(3, buffer.Len())

	b64decode('6', &input, &buffer)
	assert.Equal(0, input.Len())
	assert.Equal(6, buffer.Len())
	assert.Equal("diego:", buffer.String())

	// ----------

	b64decode('c', &input, &buffer)
	assert.Equal(1, input.Len())
	assert.Equal(6, buffer.Len())

	b64decode('G', &input, &buffer)
	assert.Equal(2, input.Len())
	assert.Equal(6, buffer.Len())

	b64decode('F', &input, &buffer)
	assert.Equal(3, input.Len())
	assert.Equal(6, buffer.Len())

	b64decode('z', &input, &buffer)
	assert.Equal(0, input.Len())
	assert.Equal(9, buffer.Len())
	assert.Equal("diego:pas", buffer.String())

	// ----------

	b64decode('c', &input, &buffer)
	assert.Equal(1, input.Len())
	assert.Equal(9, buffer.Len())

	b64decode('3', &input, &buffer)
	assert.Equal(2, input.Len())
	assert.Equal(9, buffer.Len())

	b64decode('d', &input, &buffer)
	assert.Equal(3, input.Len())
	assert.Equal(9, buffer.Len())

	b64decode('v', &input, &buffer)
	assert.Equal(0, input.Len())
	assert.Equal(12, buffer.Len())
	assert.Equal("diego:passwo", buffer.String())

	// ----------

	b64decode('c', &input, &buffer)
	assert.Equal(1, input.Len())
	assert.Equal(12, buffer.Len())

	b64decode('m', &input, &buffer)
	assert.Equal(2, input.Len())
	assert.Equal(12, buffer.Len())

	b64decode('Q', &input, &buffer)
	assert.Equal(3, input.Len())
	assert.Equal(12, buffer.Len())

	b64decode('=', &input, &buffer)
	assert.Equal(0, input.Len())
	assert.Equal(14, buffer.Len())
	assert.Equal("diego:password", buffer.String())
}
