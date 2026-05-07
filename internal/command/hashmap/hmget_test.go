package hashmap

import (
	"avacado/internal/protocol"
	mockhashmaps "avacado/internal/storage/hashmaps/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHMGetCommand_Execute_AllFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := hMGet{key: "myhash", fields: []string{"f1", "f2"}}
	ctx := context.Background()
	store := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	store.EXPECT().Maps().Return(maps)
	maps.EXPECT().HMGet(ctx, "myhash", []string{"f1", "f2"}).Return([]any{[]byte("v1"), []byte("v2")})
	response := cmd.Execute(ctx, store)
	assert.Nil(t, response.Err)
	assert.Equal(t, protocol.ValueType(protocol.TypeArray), response.Value.Type)
	assert.Equal(t, 2, len(response.Value.Array))
	assert.Equal(t, []byte("v1"), response.Value.Array[0].Bytes)
	assert.Equal(t, []byte("v2"), response.Value.Array[1].Bytes)
}

func TestHMGetCommand_Execute_SomeNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := hMGet{key: "myhash", fields: []string{"f1", "missing"}}
	ctx := context.Background()
	store := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	store.EXPECT().Maps().Return(maps)
	maps.EXPECT().HMGet(ctx, "myhash", []string{"f1", "missing"}).Return([]any{[]byte("v1"), nil})
	response := cmd.Execute(ctx, store)
	assert.Nil(t, response.Err)
	assert.Equal(t, protocol.ValueType(protocol.TypeArray), response.Value.Type)
	assert.Equal(t, 2, len(response.Value.Array))
	assert.Equal(t, []byte("v1"), response.Value.Array[0].Bytes)
	assert.True(t, response.Value.Array[1].Null)
	assert.Equal(t, protocol.ValueType(protocol.TypeBulkString), response.Value.Array[1].Type)
}

func TestHMGetCommand_Execute_KeyNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := hMGet{key: "nonexistent", fields: []string{"f1", "f2"}}
	ctx := context.Background()
	store := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	store.EXPECT().Maps().Return(maps)
	maps.EXPECT().HMGet(ctx, "nonexistent", []string{"f1", "f2"}).Return([]any{nil, nil})
	response := cmd.Execute(ctx, store)
	assert.Nil(t, response.Err)
	assert.Equal(t, protocol.ValueType(protocol.TypeArray), response.Value.Type)
	assert.Equal(t, 2, len(response.Value.Array))
	assert.True(t, response.Value.Array[0].Null)
	assert.True(t, response.Value.Array[1].Null)
}

func TestHMGetParser_Parse_Valid(t *testing.T) {
	parser := NewHMGetParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "HMGET",
		Args:    []string{"myhash", "f1", "f2"},
	})
	assert.NoError(t, err)
	hmget := cmd.(*hMGet)
	assert.Equal(t, "myhash", hmget.key)
	assert.Equal(t, []string{"f1", "f2"}, hmget.fields)
}

func TestHMGetParser_Parse_TooFewArgs(t *testing.T) {
	parser := NewHMGetParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "HMGET",
		Args:    []string{"myhash"},
	})
	assert.Error(t, err)

	_, err = parser.Parse(&protocol.Message{
		Command: "HMGET",
		Args:    []string{},
	})
	assert.Error(t, err)
}

func TestHMGetParser_Name(t *testing.T) {
	parser := NewHMGetParser()
	assert.Equal(t, "HMGET", parser.Name())
}
