package list

import (
	"avacado/internal/protocol"
	mocklists "avacado/internal/storage/lists/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLLenCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LLen{Key: "mylist"}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().Len(ctx, "mylist").Return(5, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(5), response.Value.Number)
}

func TestLLenCommand_ExecuteEmptyList(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LLen{Key: "mylist"}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().Len(ctx, "mylist").Return(0, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(0), response.Value.Number)
}

func TestLLenCommand_ExecuteError(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LLen{Key: "mylist"}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().Len(ctx, "mylist").Return(0, fmt.Errorf("some error"))

	response := cmd.Execute(ctx, storage)
	assert.NotNil(t, response.Err)
}

func TestLLenParser_Parse(t *testing.T) {
	parser := NewLLenParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "LLEN",
		Args: []protocol.Value{
			{
				Type:  protocol.TypeBulkString,
				Bytes: []byte("mylist"),
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "mylist", cmd.(*LLen).Key)
}

func TestLLenParser_ParseNoArgs(t *testing.T) {
	parser := NewLLenParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LLEN",
		Args:    []protocol.Value{},
	})
	assert.Error(t, err)
}

func TestLLenParser_ParseTooManyArgs(t *testing.T) {
	parser := NewLLenParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LLEN",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("key1")},
			{Type: protocol.TypeBulkString, Bytes: []byte("key2")},
		},
	})
	assert.Error(t, err)
}

func TestLLenParser_Name(t *testing.T) {
	parser := NewLLenParser()
	assert.Equal(t, "LLEN", parser.Name())
}
