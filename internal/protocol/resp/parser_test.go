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

func TestParser_ParseMap(t *testing.T) {
	data := "%2\r\n+first\r\n:1\r\n+second\r\n$2\r\nHi\r\n"

	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	value, err := parser.Parse()

	assert.NoError(t, err)
	assert.Equal(t, TypeMap, value.Type)
	entries := value.Map

	assert.Equal(t, 2, len(entries))
}

func TestParser_ParseMapWithBulkStringKeys(t *testing.T) {
	// %2\r\n$3\r\nkey\r\n$5\r\nvalue\r\n$4\r\nname\r\n$4\r\njohn\r\n
	data := "%2\r\n$3\r\nkey\r\n$5\r\nvalue\r\n$4\r\nname\r\n$4\r\njohn\r\n"
	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	value, err := parser.Parse()

	assert.NoError(t, err)
	assert.Equal(t, TypeMap, value.Type)
	assert.True(t, value.IsMap())

	entries, err := value.AsMap()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(entries))

	assert.Equal(t, "key", entries[0].Key)
	assert.Equal(t, []byte("value"), entries[0].Val.Bulk)
	assert.Equal(t, TypeBulkString, entries[0].Val.Type)

	assert.Equal(t, "name", entries[1].Key)
	assert.Equal(t, []byte("john"), entries[1].Val.Bulk)
	assert.Equal(t, TypeBulkString, entries[1].Val.Type)
}

func TestParser_ParseEmptyMap(t *testing.T) {
	data := "%0\r\n"
	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	value, err := parser.Parse()

	assert.NoError(t, err)
	assert.Equal(t, TypeMap, value.Type)
	assert.False(t, value.Null)

	entries, err := value.AsMap()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(entries))
}

func TestParser_ParseNullMap(t *testing.T) {
	data := "%-1\r\n"
	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	value, err := parser.Parse()

	assert.NoError(t, err)
	assert.Equal(t, TypeMap, value.Type)
	assert.True(t, value.Null)
	assert.True(t, value.IsMap())

	_, err = value.AsMap()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "null map")
}

func TestParser_ParseMapWithMixedValueTypes(t *testing.T) {
	// Map with string, integer, bulk string, and array values
	data := "%4\r\n" +
		"+key1\r\n+simple\r\n" +
		"+key2\r\n:42\r\n" +
		"+key3\r\n$4\r\ndata\r\n" +
		"+key4\r\n*2\r\n:1\r\n:2\r\n"

	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	value, err := parser.Parse()

	assert.NoError(t, err)
	assert.Equal(t, TypeMap, value.Type)

	entries, err := value.AsMap()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(entries))

	assert.Equal(t, "key1", entries[0].Key)
	assert.Equal(t, TypeSimpleString, entries[0].Val.Type)
	assert.Equal(t, "simple", entries[0].Val.Str)

	assert.Equal(t, "key2", entries[1].Key)
	assert.Equal(t, TypeInteger, entries[1].Val.Type)
	assert.Equal(t, int64(42), entries[1].Val.Num)

	assert.Equal(t, "key3", entries[2].Key)
	assert.Equal(t, TypeBulkString, entries[2].Val.Type)
	assert.Equal(t, []byte("data"), entries[2].Val.Bulk)

	assert.Equal(t, "key4", entries[3].Key)
	assert.Equal(t, TypeArray, entries[3].Val.Type)
}

func TestParser_ParseNestedMap(t *testing.T) {
	// Map with a map value: %1\r\n+outer\r\n%1\r\n+inner\r\n:42\r\n
	data := "%1\r\n+outer\r\n%1\r\n+inner\r\n:42\r\n"

	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	value, err := parser.Parse()

	assert.NoError(t, err)
	assert.Equal(t, TypeMap, value.Type)

	entries, err := value.AsMap()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, "outer", entries[0].Key)

	innerMap := entries[0].Val
	assert.Equal(t, TypeMap, innerMap.Type)
	assert.True(t, innerMap.IsMap())

	innerEntries, err := innerMap.AsMap()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(innerEntries))
	assert.Equal(t, "inner", innerEntries[0].Key)
	assert.Equal(t, int64(42), innerEntries[0].Val.Num)
}

func TestParser_ParseMapPreservesOrder(t *testing.T) {
	data := "%4\r\n" +
		"+first\r\n:1\r\n" +
		"+second\r\n:2\r\n" +
		"+third\r\n:3\r\n" +
		"+fourth\r\n:4\r\n"

	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	value, err := parser.Parse()

	assert.NoError(t, err)
	entries, err := value.AsMap()
	assert.NoError(t, err)

	assert.Equal(t, "first", entries[0].Key)
	assert.Equal(t, "second", entries[1].Key)
	assert.Equal(t, "third", entries[2].Key)
	assert.Equal(t, "fourth", entries[3].Key)
}

func TestParser_ParseMapWithInvalidSize(t *testing.T) {
	data := "%-5\r\n"
	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	_, err := parser.Parse()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid map size")
}

func TestParser_ParseMapWithMissingValue(t *testing.T) {
	// Map declares 2 entries but only provides 1 complete pair
	data := "%2\r\n+key1\r\n:1\r\n+key2\r\n"
	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	_, err := parser.Parse()

	assert.Error(t, err)
}

func TestParser_ParseMapWithNonStringKey(t *testing.T) {
	// Map with integer key (invalid - keys must be strings)
	data := "%1\r\n:123\r\n+value\r\n"
	parser := Parser{reader: bufio.NewReader(strings.NewReader(data))}
	_, err := parser.Parse()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected key type")
}
