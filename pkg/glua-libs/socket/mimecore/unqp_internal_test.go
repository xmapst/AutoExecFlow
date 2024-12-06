package mimecore

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQpDecode0(t *testing.T) {
	assert := assert.New(t)
	var input, buffer bytes.Buffer

	qpdecode('=', &input, &buffer)
	assert.Equal(1, input.Len())
	assert.Equal(0, buffer.Len())

	qpdecode('3', &input, &buffer)
	assert.Equal(2, input.Len())
	assert.Equal(0, buffer.Len())

	qpdecode('0', &input, &buffer)
	assert.Equal(0, input.Len())
	assert.Equal(1, buffer.Len())
	assert.Equal("0", buffer.String())
}
