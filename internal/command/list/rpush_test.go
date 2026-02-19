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

func TestRPushCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := RPush{Key: "mylist", Values: [][]byte{[]byte("a"), []byte("b")}}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPush(ctx, "mylist", []byte("a"), []byte("b")).Return(2, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(2), response.Value.Number)
}

func TestRPushCommand_ExecuteError(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := RPush{Key: "mylist", Values: [][]byte{[]byte("a")}}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPush(ctx, "mylist", []byte("a")).Return(0, fmt.Errorf("some error"))

	response := cmd.Execute(ctx, storage)
	assert.NotNil(t, response.Err)
}

func TestRPushParser_Parse(t *testing.T) {
	parser := RPushParser{}
	cmd, err := parser.Parse(&protocol.Message{
		Command: "RPUSH",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
			{Type: protocol.TypeBulkString, Bytes: []byte("val1")},
			{Type: protocol.TypeBulkString, Bytes: []byte("val2")},
		},
	})
	assert.NoError(t, err)
	rpush := cmd.(*RPush)
	assert.Equal(t, "mylist", rpush.Key)
	assert.Equal(t, [][]byte{[]byte("val1"), []byte("val2")}, rpush.Values)
}

func TestRPushParser_ParseNoArgs(t *testing.T) {
	parser := RPushParser{}
	_, err := parser.Parse(&protocol.Message{
		Command: "RPUSH",
		Args:    []protocol.Value{},
	})
	assert.Error(t, err)
}

func TestRPushParser_ParseOnlyKey(t *testing.T) {
	parser := RPushParser{}
	_, err := parser.Parse(&protocol.Message{
		Command: "RPUSH",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
		},
	})
	assert.Error(t, err)
}

func TestRPushParser_Name(t *testing.T) {
	parser := RPushParser{}
	assert.Equal(t, "RPUSH", parser.Name())
}
