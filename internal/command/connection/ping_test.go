package connection

import (
	"avacado/internal/protocol"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPingCommand_Execute_NoMessage(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Ping{}
	storage := mocksstorage.NewMockStorage(controller)

	response := cmd.Execute(context.Background(), storage)

	assert.NoError(t, response.Err)
	assert.Equal(t, "PONG", response.Value.Str)
	assert.Equal(t, protocol.TypeSimpleString, response.Value.Type)
}

func TestPingCommand_Execute_WithMessage(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Ping{Message: "hello"}
	storage := mocksstorage.NewMockStorage(controller)

	response := cmd.Execute(context.Background(), storage)

	assert.NoError(t, response.Err)
	assert.Equal(t, []byte("hello"), response.Value.Bytes)
	assert.Equal(t, protocol.ValueType(protocol.TypeBulkString), response.Value.Type)
}

func TestPingParser_Parse_NoArgs(t *testing.T) {
	parser := NewPingParser()
	msg := &protocol.Message{Command: "PING", Args: []string{}}

	cmd, err := parser.Parse(msg)

	assert.NoError(t, err)
	assert.IsType(t, &Ping{}, cmd)
	assert.Equal(t, "", cmd.(*Ping).Message)
}

func TestPingParser_Parse_WithMessage(t *testing.T) {
	parser := NewPingParser()
	msg := &protocol.Message{Command: "PING", Args: []string{"hello"}}

	cmd, err := parser.Parse(msg)

	assert.NoError(t, err)
	assert.IsType(t, &Ping{}, cmd)
	assert.Equal(t, "hello", cmd.(*Ping).Message)
}
