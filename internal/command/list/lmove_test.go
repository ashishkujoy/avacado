package list

import (
	"avacado/internal/protocol"
	"avacado/internal/storage/lists"
	mocklists "avacado/internal/storage/lists/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLMoveParser_Parse(t *testing.T) {
	parser := NewLMoveParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "LMOVE",
		Args:    []string{"src", "dst", "LEFT", "RIGHT"},
	})
	assert.NoError(t, err)
	lmove := cmd.(*LMove)
	assert.Equal(t, "src", lmove.Source)
	assert.Equal(t, "dst", lmove.Destination)
	assert.Equal(t, lists.Left, lmove.SourceDirection)
	assert.Equal(t, lists.Right, lmove.DestinationDirection)
}

func TestLMoveParser_ParseCaseInsensitive(t *testing.T) {
	parser := NewLMoveParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "LMOVE",
		Args:    []string{"src", "dst", "right", "left"},
	})
	assert.NoError(t, err)
	lmove := cmd.(*LMove)
	assert.Equal(t, lists.Right, lmove.SourceDirection)
	assert.Equal(t, lists.Left, lmove.DestinationDirection)
}

func TestLMoveParser_ParseInvalidArgCount(t *testing.T) {
	parser := NewLMoveParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LMOVE",
		Args:    []string{"src"},
	})
	assert.Error(t, err)
}

func TestLMoveParser_ParseInvalidDirection(t *testing.T) {
	parser := NewLMoveParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LMOVE",
		Args:    []string{"src", "dst", "INVALID", "LEFT"},
	})
	assert.Error(t, err)
}

func TestLMoveParser_ParseInvalidDestinationDirection(t *testing.T) {
	parser := NewLMoveParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LMOVE",
		Args:    []string{"src", "dst", "LEFT", "UP"},
	})
	assert.Error(t, err)
}

func TestLMoveParser_Name(t *testing.T) {
	parser := NewLMoveParser()
	assert.Equal(t, "LMOVE", parser.Name())
}

func TestLMoveCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LMove{Source: "src", Destination: "dst", SourceDirection: lists.Left, DestinationDirection: lists.Right}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	listsMock := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(listsMock)
	listsMock.EXPECT().LMove(ctx, "src", "dst", lists.Left, lists.Right).Return([]byte("val1"), nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, []byte("val1"), response.Value.Bytes)
}

func TestLMoveCommand_ExecuteSourceNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LMove{Source: "nonexistent", Destination: "dst", SourceDirection: lists.Left, DestinationDirection: lists.Left}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	listsMock := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(listsMock)
	listsMock.EXPECT().LMove(ctx, "nonexistent", "dst", lists.Left, lists.Left).Return(nil, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
}

func TestLMoveCommand_ExecuteHandlesError(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LMove{Source: "src", Destination: "dst", SourceDirection: lists.Right, DestinationDirection: lists.Right}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	listsMock := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(listsMock)
	listsMock.EXPECT().LMove(ctx, "src", "dst", lists.Right, lists.Right).Return(nil, fmt.Errorf("some error"))

	response := cmd.Execute(ctx, storage)
	assert.NotNil(t, response.Err)
}
