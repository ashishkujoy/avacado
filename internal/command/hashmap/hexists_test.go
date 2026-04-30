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

func TestHExistsCommand_ExecuteFieldExists(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := HExists{key: "myhash", field: "field1"}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HExists(ctx, "myhash", "field1").Return(1)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(1), response.Value.Number)
}

func TestHExistsCommand_ExecuteFieldNotExists(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := HExists{key: "myhash", field: "missing"}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HExists(ctx, "myhash", "missing").Return(0)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(0), response.Value.Number)
}

func TestHExistsCommand_ExecuteKeyNotExists(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := HExists{key: "nonexistent", field: "field1"}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HExists(ctx, "nonexistent", "field1").Return(0)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(0), response.Value.Number)
}

func TestHExistsParser_Parse(t *testing.T) {
	parser := NewHExistsParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "HEXISTS",
		Args:    []string{"myhash", "field1"},
	})
	assert.NoError(t, err)
	hexists := cmd.(*HExists)
	assert.Equal(t, "myhash", hexists.key)
	assert.Equal(t, "field1", hexists.field)
}

func TestHExistsParser_ParseTooFewArgs(t *testing.T) {
	parser := NewHExistsParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "HEXISTS",
		Args:    []string{"myhash"},
	})
	assert.Error(t, err)
}

func TestHExistsParser_ParseNoArgs(t *testing.T) {
	parser := NewHExistsParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "HEXISTS",
		Args:    []string{},
	})
	assert.Error(t, err)
}

func TestHExistsParser_Name(t *testing.T) {
	parser := NewHExistsParser()
	assert.Equal(t, "HEXISTS", parser.Name())
}
