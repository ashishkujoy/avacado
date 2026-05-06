package hashmap

import (
	"avacado/internal/protocol"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHIncrByParser_ValidArguments(t *testing.T) {
	parser := NewHIncrByParser()
	msg := &protocol.Message{
		Command: "HINCRBY",
		Args:    []string{"key1", "field1", "10"},
	}

	cmd, err := parser.Parse(msg)
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	hincr, ok := cmd.(*HIncrBy)
	assert.True(t, ok)
	assert.Equal(t, "key1", hincr.key)
	assert.Equal(t, "field1", hincr.field)
	assert.Equal(t, int64(10), hincr.increment)
}

func TestHIncrByParser_NegativeIncrement(t *testing.T) {
	parser := NewHIncrByParser()
	msg := &protocol.Message{
		Command: "HINCRBY",
		Args:    []string{"key1", "field1", "-5"},
	}

	cmd, err := parser.Parse(msg)
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	hincr, ok := cmd.(*HIncrBy)
	assert.True(t, ok)
	assert.Equal(t, int64(-5), hincr.increment)
}

func TestHIncrByParser_MissingArguments(t *testing.T) {
	parser := NewHIncrByParser()
	msg := &protocol.Message{
		Command: "HINCRBY",
		Args:    []string{"key1"},
	}

	cmd, err := parser.Parse(msg)
	assert.Error(t, err)
	assert.Nil(t, cmd)
}

func TestHIncrByParser_TooManyArguments(t *testing.T) {
	parser := NewHIncrByParser()
	msg := &protocol.Message{
		Command: "HINCRBY",
		Args:    []string{"key1", "field1", "10", "extra"},
	}

	cmd, err := parser.Parse(msg)
	assert.Error(t, err)
	assert.Nil(t, cmd)
}

func TestHIncrByParser_InvalidIncrementValue(t *testing.T) {
	parser := NewHIncrByParser()
	msg := &protocol.Message{
		Command: "HINCRBY",
		Args:    []string{"key1", "field1", "notanumber"},
	}

	cmd, err := parser.Parse(msg)
	assert.Error(t, err)
	assert.Nil(t, cmd)
}

func TestHIncrByParser_Name(t *testing.T) {
	parser := NewHIncrByParser()
	assert.Equal(t, "HINCRBY", parser.Name())
}

func TestHIncrByParser_LargeIncrement(t *testing.T) {
	parser := NewHIncrByParser()
	msg := &protocol.Message{
		Command: "HINCRBY",
		Args:    []string{"key1", "field1", "9223372036854775807"},
	}

	cmd, err := parser.Parse(msg)
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	hincr, ok := cmd.(*HIncrBy)
	assert.True(t, ok)
	assert.Equal(t, int64(9223372036854775807), hincr.increment)
}
