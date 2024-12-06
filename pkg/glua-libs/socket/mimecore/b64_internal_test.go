package mimecore

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64EncodeMoepsi(t *testing.T) {
	assert := assert.New(t)
	var input, buffer bytes.Buffer

	b64encode('M', &input, &buffer)
	assert.Equal(1, input.Len())
	assert.Equal("M", input.String())
	assert.Equal(0, buffer.Len())

	b64encode('o', &input, &buffer)
	assert.Equal(2, input.Len())
	assert.Equal(0, buffer.Len())

	b64encode('e', &input, &buffer)
	assert.Equal(0, input.Len())
	assert.Equal(4, buffer.Len())
	assert.Equal("TW9l", buffer.String())

	// ----------

	b64encode('p', &input, &buffer)
	assert.Equal(1, input.Len())
	assert.Equal(4, buffer.Len())

	b64encode('s', &input, &buffer)
	assert.Equal(2, input.Len())
	assert.Equal(4, buffer.Len())

	b64encode('i', &input, &buffer)
	assert.Equal(0, input.Len())
	assert.Equal("TW9lcHNp", buffer.String())
}
