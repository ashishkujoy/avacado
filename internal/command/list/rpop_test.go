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

func TestRPopCommand_ExecuteWithoutCount(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := RPop{Key: "mylist", Count: 1, HasCount: false}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPop(ctx, "mylist", 1).Return([][]byte{[]byte("val1")}, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, []byte("val1"), response.Value.Bytes)
}

func TestRPopCommand_ExecuteWithCountOne(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := RPop{Key: "mylist", Count: 1, HasCount: true}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPop(ctx, "mylist", 1).Return([][]byte{[]byte("val1")}, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, protocol.ValueType('*'), response.Value.Type)
	assert.Len(t, response.Value.Array, 1)
	assert.Equal(t, []byte("val1"), response.Value.Array[0].Bytes)
}

func TestRPopCommand_ExecuteWithCountMultiple(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := RPop{Key: "mylist", Count: 2, HasCount: true}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPop(ctx, "mylist", 2).Return([][]byte{[]byte("val2"), []byte("val1")}, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, protocol.ValueType('*'), response.Value.Type)
	assert.Len(t, response.Value.Array, 2)
	assert.Equal(t, []byte("val2"), response.Value.Array[0].Bytes)
	assert.Equal(t, []byte("val1"), response.Value.Array[1].Bytes)
}

func TestRPopCommand_ExecuteKeyNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := RPop{Key: "mylist", Count: 1}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPop(ctx, "mylist", 1).Return(nil, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
}

func TestRPopCommand_ExecuteError(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := RPop{Key: "mylist", Count: 1}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPop(ctx, "mylist", 1).Return(nil, fmt.Errorf("some error"))

	response := cmd.Execute(ctx, storage)
	assert.NotNil(t, response.Err)
}

func TestRPopParser_ParseWithoutCount(t *testing.T) {
	parser := NewRPopParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "RPOP",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
		},
	})
	assert.NoError(t, err)
	rpop := cmd.(*RPop)
	assert.Equal(t, "mylist", rpop.Key)
	assert.Equal(t, 1, rpop.Count)
	assert.False(t, rpop.HasCount)
}

func TestRPopParser_ParseWithCount(t *testing.T) {
	parser := NewRPopParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "RPOP",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
			{Type: protocol.TypeNumber, Number: 3},
		},
	})
	assert.NoError(t, err)
	rpop := cmd.(*RPop)
	assert.Equal(t, "mylist", rpop.Key)
	assert.Equal(t, 3, rpop.Count)
	assert.True(t, rpop.HasCount)
}

func TestRPopParser_ParseNoArgs(t *testing.T) {
	parser := NewRPopParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "RPOP",
		Args:    []protocol.Value{},
	})
	assert.Error(t, err)
}

func TestRPopParser_ParseTooManyArgs(t *testing.T) {
	parser := NewRPopParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "RPOP",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("key")},
			{Type: protocol.TypeNumber, Number: 2},
			{Type: protocol.TypeBulkString, Bytes: []byte("extra")},
		},
	})
	assert.Error(t, err)
}

func TestRPopParser_Name(t *testing.T) {
	parser := NewRPopParser()
	assert.Equal(t, "RPOP", parser.Name())
}
