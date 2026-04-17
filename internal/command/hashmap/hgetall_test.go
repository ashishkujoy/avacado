package hashmap

import (
	"avacado/internal/config"
	"avacado/internal/protocol"
	mockhashmaps "avacado/internal/storage/hashmaps/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func proto2Ctx() context.Context {
	return context.WithValue(context.Background(), "clientConfig", config.DefaultClientConfig())
}

func proto3Ctx() context.Context {
	return context.WithValue(context.Background(), "clientConfig", &config.ClientConfig{ProtocolVersion: 3})
}

func TestHGetAllCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := hGetAll{key: "myhash"}
	ctx := proto2Ctx()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HGetAll(ctx, "myhash").Return(map[string]string{"field1": "value1"}, nil)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, protocol.ValueType(protocol.TypeArray), response.Value.Type)
	assert.Len(t, response.Value.Array, 2)
}

func TestHGetAllCommand_ExecuteProto3(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := hGetAll{key: "myhash"}
	ctx := proto3Ctx()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HGetAll(ctx, "myhash").Return(map[string]string{"field1": "value1"}, nil)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, protocol.ValueType(protocol.TypeMap), response.Value.Type)
	assert.Len(t, response.Value.Map, 1)
	assert.Equal(t, "field1", response.Value.Map[0].Key)
	assert.Equal(t, "value1", response.Value.Map[0].Val.Str)
}

func TestHGetAllCommand_ExecuteHashNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := hGetAll{key: "nonexistent"}
	ctx := proto2Ctx()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HGetAll(ctx, "nonexistent").Return(nil, fmt.Errorf("nonexistent does not exist"))
	response := cmd.Execute(ctx, storage)
	assert.NotNil(t, response.Err)
}

func TestHGetAllParser_Parse(t *testing.T) {
	parser := &HGetAllParser{}
	cmd, err := parser.Parse(&protocol.Message{
		Command: "HGETALL",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("myhash")},
		},
	})
	assert.NoError(t, err)
	hgetall := cmd.(*hGetAll)
	assert.Equal(t, "myhash", hgetall.key)
}

func TestHGetAllParser_ParseTooFewArgs(t *testing.T) {
	parser := &HGetAllParser{}
	_, err := parser.Parse(&protocol.Message{
		Command: "HGETALL",
		Args:    []protocol.Value{},
	})
	assert.Error(t, err)
}

func TestHGetAllParser_ParseTooManyArgs(t *testing.T) {
	parser := &HGetAllParser{}
	_, err := parser.Parse(&protocol.Message{
		Command: "HGETALL",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("myhash")},
			{Type: protocol.TypeBulkString, Bytes: []byte("extra")},
		},
	})
	assert.Error(t, err)
}

func TestHGetAllParser_Name(t *testing.T) {
	parser := &HGetAllParser{}
	assert.Equal(t, "HGETALL", parser.Name())
}

