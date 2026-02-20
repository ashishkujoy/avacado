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

func TestLPopCommand_ExecuteWithoutCount(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LPop{Key: "mylist", Count: 1, HasCount: false}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LPop(ctx, "mylist", 1).Return([][]byte{[]byte("val1")}, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, []byte("val1"), response.Value.Bytes)
}

func TestLPopCommand_ExecuteWithCountOne(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LPop{Key: "mylist", Count: 1, HasCount: true}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LPop(ctx, "mylist", 1).Return([][]byte{[]byte("val1")}, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, protocol.ValueType('*'), response.Value.Type)
	assert.Len(t, response.Value.Array, 1)
	assert.Equal(t, []byte("val1"), response.Value.Array[0].Bytes)
}

func TestLPopCommand_ExecuteWithCountMultiple(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LPop{Key: "mylist", Count: 2, HasCount: true}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LPop(ctx, "mylist", 2).Return([][]byte{[]byte("val1"), []byte("val2")}, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, protocol.ValueType('*'), response.Value.Type)
	assert.Len(t, response.Value.Array, 2)
	assert.Equal(t, []byte("val1"), response.Value.Array[0].Bytes)
	assert.Equal(t, []byte("val2"), response.Value.Array[1].Bytes)
}

func TestLPopCommand_ExecuteKeyNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LPop{Key: "mylist", Count: 1, HasCount: false}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LPop(ctx, "mylist", 1).Return(nil, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
}

func TestLPopCommand_ExecuteError(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := LPop{Key: "mylist", Count: 1, HasCount: false}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LPop(ctx, "mylist", 1).Return(nil, fmt.Errorf("some error"))

	response := cmd.Execute(ctx, storage)
	assert.NotNil(t, response.Err)
}

func TestLPopParser_ParseWithoutCount(t *testing.T) {
	parser := NewLPopParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "LPOP",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
		},
	})
	assert.NoError(t, err)
	lpop := cmd.(*LPop)
	assert.Equal(t, "mylist", lpop.Key)
	assert.Equal(t, 1, lpop.Count)
	assert.False(t, lpop.HasCount)
}

func TestLPopParser_ParseWithCount(t *testing.T) {
	parser := NewLPopParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "LPOP",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
			{Type: protocol.TypeNumber, Number: 3},
		},
	})
	assert.NoError(t, err)
	lpop := cmd.(*LPop)
	assert.Equal(t, "mylist", lpop.Key)
	assert.Equal(t, 3, lpop.Count)
	assert.True(t, lpop.HasCount)
}

func TestLPopParser_ParseNoArgs(t *testing.T) {
	parser := NewLPopParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LPOP",
		Args:    []protocol.Value{},
	})
	assert.Error(t, err)
}

func TestLPopParser_ParseTooManyArgs(t *testing.T) {
	parser := NewLPopParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "LPOP",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("key")},
			{Type: protocol.TypeNumber, Number: 2},
			{Type: protocol.TypeBulkString, Bytes: []byte("extra")},
		},
	})
	assert.Error(t, err)
}

func TestLPopParser_Name(t *testing.T) {
	parser := NewLPopParser()
	assert.Equal(t, "LPOP", parser.Name())
}
