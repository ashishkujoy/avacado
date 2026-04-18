package resp

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		"$3\r\nkey\r\n" +
		"$5\r\nvalue\r\n"
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader(data)),
	}
	value, err := parser.Parse()
	assert.NoError(t, err)
	assert.Equal(t, TypeArray, value.Type)

	values, err := value.AsArray()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(values))
	assert.Equal(t, TypeBulkString, values[0].Type)
	assert.Equal(t, TypeBulkString, values[1].Type)
	assert.Equal(t, TypeBulkString, values[2].Type)
}

func TestParser_RejectsUnsupportedTypes(t *testing.T) {
	unsupported := []string{
		"+OK\r\n",          // simple string
		"-Error\r\n",       // error
		":123\r\n",         // integer
		"%1\r\n$3\r\nkey\r\n$3\r\nval\r\n", // map
	}
	for _, input := range unsupported {
		parser := Parser{reader: bufio.NewReader(strings.NewReader(input))}
		_, err := parser.Parse()
		assert.Error(t, err, "expected error for input: %q", input)
	}
}

func TestParser_ParseNullBulkString(t *testing.T) {
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader("$-1\r\n")),
	}
	value, err := parser.Parse()
	assert.NoError(t, err)
	assert.True(t, value.Null)
	assert.Equal(t, TypeBulkString, value.Type)
}

func TestParser_ParseNullArray(t *testing.T) {
	parser := Parser{
		reader: bufio.NewReader(strings.NewReader("*-1\r\n")),
	}
	value, err := parser.Parse()
	assert.NoError(t, err)
	assert.True(t, value.Null)
	assert.Equal(t, TypeArray, value.Type)
}
