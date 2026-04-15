package hashmap

import (
	"avacado/internal/protocol"
	mockhashmaps "avacado/internal/storage/hashmaps/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHGetCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := hGet{name: "myhash", field: "field1"}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HGet(ctx, "myhash", "field1").Return([]byte("value1"), nil)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, []byte("value1"), response.Value.Bytes)
}

func TestHGetCommand_ExecuteFieldNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := hGet{name: "myhash", field: "missing"}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HGet(ctx, "myhash", "missing").Return(nil, fmt.Errorf("missing field does not exist in myhash map"))
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
	assert.Equal(t, protocol.ValueType(protocol.TypeBulkString), response.Value.Type)
}

func TestHGetCommand_ExecuteHashNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := hGet{name: "nonexistent", field: "field1"}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HGet(ctx, "nonexistent", "field1").Return(nil, fmt.Errorf("nonexistent does not exist"))
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
	assert.Equal(t, protocol.ValueType(protocol.TypeBulkString), response.Value.Type)
}

func TestHGetParser_Parse(t *testing.T) {
	parser := NewHGetParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "HGET",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("myhash")},
			{Type: protocol.TypeBulkString, Bytes: []byte("field1")},
		},
	})
	assert.NoError(t, err)
	hget := cmd.(*hGet)
	assert.Equal(t, "myhash", hget.name)
	assert.Equal(t, "field1", hget.field)
}

func TestHGetParser_ParseTooFewArgs(t *testing.T) {
	parser := NewHGetParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "HGET",
		Args:    []protocol.Value{},
	})
	assert.Error(t, err)
}

func TestHGetParser_ParseTooManyArgs(t *testing.T) {
	parser := NewHGetParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "HGET",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("myhash")},
			{Type: protocol.TypeBulkString, Bytes: []byte("field1")},
			{Type: protocol.TypeBulkString, Bytes: []byte("extra")},
		},
	})
	assert.Error(t, err)
}

func TestHGetParser_Name(t *testing.T) {
	parser := NewHGetParser()
	assert.Equal(t, "HGET", parser.Name())
}

