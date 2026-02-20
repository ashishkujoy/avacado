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

// Execute tests for both directions

func TestPopCommand_ExecuteLeftWithoutCount(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Pop{Key: "mylist", Count: 1, HasCount: false, Direction: PopLeft}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LPop(ctx, "mylist", 1).Return([][]byte{[]byte("val1")}, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, []byte("val1"), response.Value.Bytes)
}

func TestPopCommand_ExecuteLeftWithCountOne(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Pop{Key: "mylist", Count: 1, HasCount: true, Direction: PopLeft}
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

func TestPopCommand_ExecuteLeftWithCountMultiple(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Pop{Key: "mylist", Count: 2, HasCount: true, Direction: PopLeft}
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

func TestPopCommand_ExecuteLeftKeyNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Pop{Key: "mylist", Count: 1, HasCount: false, Direction: PopLeft}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LPop(ctx, "mylist", 1).Return(nil, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
}

func TestPopCommand_ExecuteLeftError(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Pop{Key: "mylist", Count: 1, HasCount: false, Direction: PopLeft}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().LPop(ctx, "mylist", 1).Return(nil, fmt.Errorf("some error"))

	response := cmd.Execute(ctx, storage)
	assert.NotNil(t, response.Err)
}

func TestPopCommand_ExecuteRightWithoutCount(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Pop{Key: "mylist", Count: 1, HasCount: false, Direction: PopRight}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPop(ctx, "mylist", 1).Return([][]byte{[]byte("val1")}, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, []byte("val1"), response.Value.Bytes)
}

func TestPopCommand_ExecuteRightWithCountMultiple(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Pop{Key: "mylist", Count: 2, HasCount: true, Direction: PopRight}
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

func TestPopCommand_ExecuteRightKeyNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Pop{Key: "mylist", Count: 1, Direction: PopRight}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPop(ctx, "mylist", 1).Return(nil, nil)

	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
}

func TestPopCommand_ExecuteRightError(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := Pop{Key: "mylist", Count: 1, Direction: PopRight}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	lists := mocklists.NewMockLists(controller)

	storage.EXPECT().Lists().Return(lists)
	lists.EXPECT().RPop(ctx, "mylist", 1).Return(nil, fmt.Errorf("some error"))

	response := cmd.Execute(ctx, storage)
	assert.NotNil(t, response.Err)
}

// Parser tests

func TestLPopParser_ParseWithoutCount(t *testing.T) {
	parser := NewLPopParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "LPOP",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
		},
	})
	assert.NoError(t, err)
	pop := cmd.(*Pop)
	assert.Equal(t, "mylist", pop.Key)
	assert.Equal(t, 1, pop.Count)
	assert.False(t, pop.HasCount)
	assert.Equal(t, PopLeft, pop.Direction)
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
	pop := cmd.(*Pop)
	assert.Equal(t, "mylist", pop.Key)
	assert.Equal(t, 3, pop.Count)
	assert.True(t, pop.HasCount)
	assert.Equal(t, PopLeft, pop.Direction)
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

func TestRPopParser_ParseWithoutCount(t *testing.T) {
	parser := NewRPopParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "RPOP",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("mylist")},
		},
	})
	assert.NoError(t, err)
	pop := cmd.(*Pop)
	assert.Equal(t, "mylist", pop.Key)
	assert.Equal(t, 1, pop.Count)
	assert.False(t, pop.HasCount)
	assert.Equal(t, PopRight, pop.Direction)
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
	pop := cmd.(*Pop)
	assert.Equal(t, "mylist", pop.Key)
	assert.Equal(t, 3, pop.Count)
	assert.True(t, pop.HasCount)
	assert.Equal(t, PopRight, pop.Direction)
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
