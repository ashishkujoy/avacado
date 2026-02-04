package connection

import (
	"avacado/internal/protocol"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHelloCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	command := Hello{}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)

	response := command.Execute(ctx, storage)

	// Verify response is not nil and has no error
	assert.NotNil(t, response)
	assert.Nil(t, response.Err)

	// Verify response is a map
	assert.True(t, response.Value.IsMap())
	assert.False(t, response.Value.Null)
}

func TestHelloParser_Parse(t *testing.T) {
	parser := NewHelloParser()
	msg := &protocol.Message{
		Command: "HELLO",
		Args:    []protocol.Value{},
	}

	command, err := parser.Parse(msg)

	assert.NoError(t, err)
	assert.NotNil(t, command)
	assert.IsType(t, &Hello{}, command)
}

func TestHelloParser_ParseWithArgs(t *testing.T) {
	parser := NewHelloParser()
	msg := &protocol.Message{
		Command: "HELLO",
		Args: []protocol.Value{
			{
				Type:  protocol.TypeBulkString,
				Bytes: []byte("3"),
			},
		},
	}

	// The parser currently ignores arguments, should still succeed
	command, err := parser.Parse(msg)

	assert.NoError(t, err)
	assert.NotNil(t, command)
	assert.IsType(t, &Hello{}, command)
}
