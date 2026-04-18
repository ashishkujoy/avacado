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

func TestHDelCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := HDel{key: "myhash", fields: []string{"field1", "field2"}}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HDel(ctx, "myhash", []string{"field1", "field2"}).Return(2, nil)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(2), response.Value.Number)
}

func TestHDelCommand_ExecuteSingleField(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := HDel{key: "myhash", fields: []string{"field1"}}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HDel(ctx, "myhash", []string{"field1"}).Return(1, nil)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(1), response.Value.Number)
}

func TestHDelCommand_ExecuteNonExistentField(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := HDel{key: "myhash", fields: []string{"missing"}}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HDel(ctx, "myhash", []string{"missing"}).Return(0, nil)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(0), response.Value.Number)
}

func TestHDelParser_Parse(t *testing.T) {
	parser := NewHDelParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "HDEL",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("myhash")},
			{Type: protocol.TypeBulkString, Bytes: []byte("field1")},
		},
	})
	assert.NoError(t, err)
	hdel := cmd.(*HDel)
	assert.Equal(t, "myhash", hdel.key)
	assert.Equal(t, []string{"field1"}, hdel.fields)
}

func TestHDelParser_ParseMultipleFields(t *testing.T) {
	parser := NewHDelParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "HDEL",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("myhash")},
			{Type: protocol.TypeBulkString, Bytes: []byte("field1")},
			{Type: protocol.TypeBulkString, Bytes: []byte("field2")},
			{Type: protocol.TypeBulkString, Bytes: []byte("field3")},
		},
	})
	assert.NoError(t, err)
	hdel := cmd.(*HDel)
	assert.Equal(t, "myhash", hdel.key)
	assert.Equal(t, []string{"field1", "field2", "field3"}, hdel.fields)
}

func TestHDelParser_ParseTooFewArgs(t *testing.T) {
	parser := NewHDelParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "HDEL",
		Args:    []protocol.Value{},
	})
	assert.Error(t, err)
}

func TestHDelParser_ParseOnlyKey(t *testing.T) {
	parser := NewHDelParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "HDEL",
		Args: []protocol.Value{
			{Type: protocol.TypeBulkString, Bytes: []byte("myhash")},
		},
	})
	assert.Error(t, err)
}

func TestHDelParser_Name(t *testing.T) {
	parser := NewHDelParser()
	assert.Equal(t, "HDEL", parser.Name())
}
