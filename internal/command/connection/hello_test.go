package connection

import (
	"avacado/internal/config"
	"avacado/internal/protocol"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHelloCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	command := Hello{Proto: 3}
	ctx := context.WithValue(context.Background(), "clientConfig", config.DefaultClientConfig())

	storage := mocksstorage.NewMockStorage(controller)

	response := command.Execute(ctx, storage)

	assert.NotNil(t, response)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.IsMap())
	assert.False(t, response.Value.Null)
}

func TestHelloParser_Parse(t *testing.T) {
	parser := NewHelloParser()
	msg := &protocol.Message{
		Command: "HELLO",
		Args:    []string{},
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
		Args:    []string{"3"},
	}

	command, err := parser.Parse(msg)

	assert.NoError(t, err)
	assert.NotNil(t, command)
	assert.IsType(t, &Hello{}, command)
}
