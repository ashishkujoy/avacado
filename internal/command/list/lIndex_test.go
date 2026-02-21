package list

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocklists "avacado/internal/storage/lists/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLIndexCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LIndex{Key: "mylist", Index: 1}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	kv := mockkv.NewMockStore(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().KV().Return(kv)
	kv.EXPECT().Get(ctx, "mylist").Return(nil, fmt.Errorf("key not found"))
	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LIndex(ctx, "mylist", 1).Return([]byte("value1"), nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, []byte("value1"), response.Value.Bytes)
}

func TestLIndexCommand_ExecuteNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LIndex{Key: "mylist", Index: 99}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	kv := mockkv.NewMockStore(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().KV().Return(kv)
	kv.EXPECT().Get(ctx, "mylist").Return(nil, fmt.Errorf("key not found"))
	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LIndex(ctx, "mylist", 99).Return(nil, fmt.Errorf("index out of range"))

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
}

func TestLIndexCommand_ExecuteWrongType(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LIndex{Key: "mykey", Index: 0}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	kv := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(kv)
	kv.EXPECT().Get(ctx, "mykey").Return([]byte("stringvalue"), nil)

	response := cmd.Execute(ctx, storage)
	assert.NotNil(t, response.Err)
}

func TestLIndexParser_Parse(t *testing.T) {
	parser := NewLIndexParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "LINDEX",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
			{Type: protocol.TypeBulkString, Bytes: []byte("2")},
		},
	})
	assert.NoError(t, err)
	lindex := cmd.(*LIndex)
	assert.Equal(t, "mylist", lindex.Key)
	assert.Equal(t, 2, lindex.Index)
}

func TestLIndexParser_ParseNegativeIndex(t *testing.T) {
	parser := NewLIndexParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "LINDEX",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
			{Type: protocol.TypeBulkString, Bytes: []byte("-1")},
		},
	})
	assert.NoError(t, err)
	lindex := cmd.(*LIndex)
	assert.Equal(t, "mylist", lindex.Key)
	assert.Equal(t, -1, lindex.Index)
}

func TestLIndexParser_ParseNoArgs(t *testing.T) {
	parser := NewLIndexParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LINDEX",
		Args:    []protocol.Value{},
	})
	assert.Error(t, err)
}

func TestLIndexParser_ParseMissingIndex(t *testing.T) {
	parser := NewLIndexParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LINDEX",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
		},
	})
	assert.Error(t, err)
}

func TestLIndexParser_ParseTooManyArgs(t *testing.T) {
	parser := NewLIndexParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LINDEX",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
			{Type: protocol.TypeBulkString, Bytes: []byte("1")},
			{Type: protocol.TypeBulkString, Bytes: []byte("extra")},
		},
	})
	assert.Error(t, err)
}

func TestLIndexParser_ParseInvalidIndex(t *testing.T) {
	parser := NewLIndexParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LINDEX",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
			{Type: protocol.TypeBulkString, Bytes: []byte("notanumber")},
		},
	})
	assert.Error(t, err)
}

func TestLIndexParser_Name(t *testing.T) {
	parser := NewLIndexParser()
	assert.Equal(t, "LINDEX", parser.Name())
}
