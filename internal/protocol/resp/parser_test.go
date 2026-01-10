package resp

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser_ParseSimpleString(t *testing.T) {
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader("+OK\r\n")),
	}
	value, err := parser.Parse()
	assert.NoError(t, err)
	assert.Equal(t, "OK", value.Str)
}

func TestParser_ParseIncompleteSimpleStringFails(t *testing.T) {
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader("+OK\n")),
	}
	_, err := parser.Parse()
	assert.Error(t, err)
}

func TestParser_ParseError(t *testing.T) {
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader("-Error message\r\n")),
	}
	value, err := parser.Parse()
	assert.NoError(t, err)
	assert.Equal(t, "Error message", value.Str)
	assert.Equal(t, TypeError, value.Type)
}

func TestParser_ParseInteger(t *testing.T) {
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader(":1123\r\n")),
	}
	value, err := parser.Parse()
	assert.NoError(t, err)
	assert.Equal(t, int64(1123), value.Num)
	assert.Equal(t, TypeInteger, value.Type)
}

func TestParser_ParseBulkString(t *testing.T) {
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader("$6\r\nfoobar\r\n")),
	}
	value, err := parser.Parse()
	assert.NoError(t, err)
	assert.Equal(t, []byte("foobar"), value.Bulk)
	assert.Equal(t, TypeBulkString, value.Type)
}

func TestParser_ParseIncompleteBulkStringFails(t *testing.T) {
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader("$6\r\nfoo\n")),
	}
	_, err := parser.Parse()
	assert.Error(t, err)
}

func TestParser_ParseArray(t *testing.T) {
	data := "*3\r\n" +
		"$3\r\nSET\r\n" +
		"*2\r\n" + ":123\r\n$2\r\nOK\r\n" +
		"-Error Message\r\n"
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader(data)),
	}
	value, err := parser.Parse()
	assert.NoError(t, err)
	assert.Equal(t, TypeArray, value.Type)

	values, err := value.AsArray()
	assert.NoError(t, err)
	assert.Equal(t, TypeBulkString, values[0].Type)
	assert.Equal(t, TypeArray, values[1].Type)
	assert.Equal(t, TypeError, values[2].Type)
}
