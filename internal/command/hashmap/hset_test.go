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

func TestHSetCommand_Execute(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := HSet{name: "myhash", keyValues: []string{"field1", "value1", "field2", "value2"}}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HSet(ctx, "myhash", []string{"field1", "value1", "field2", "value2"}).Return(2)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(2), response.Value.Number)
}

func TestHSetCommand_ExecuteSingleField(t *testing.T) {
	controller := gomock.NewController(t)
	cmd := HSet{name: "myhash", keyValues: []string{"field1", "value1"}}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	maps := mockhashmaps.NewMockHashMaps(controller)
	storage.EXPECT().Maps().Return(maps)
	maps.EXPECT().HSet(ctx, "myhash", []string{"field1", "value1"}).Return(1)
	response := cmd.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, int64(1), response.Value.Number)
}

func TestHSetParser_Parse(t *testing.T) {
	parser := NewHSetParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "HSET",
		Args:    []string{"myhash", "field1", "value1"},
	})
	assert.NoError(t, err)
	hset := cmd.(*HSet)
	assert.Equal(t, "myhash", hset.name)
	assert.Equal(t, []string{"field1", "value1"}, hset.keyValues)
}

func TestHSetParser_ParseMultipleFields(t *testing.T) {
	parser := NewHSetParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "HSET",
		Args:    []string{"myhash", "field1", "value1", "field2", "value2"},
	})
	assert.NoError(t, err)
	hset := cmd.(*HSet)
	assert.Equal(t, "myhash", hset.name)
	assert.Equal(t, []string{"field1", "value1", "field2", "value2"}, hset.keyValues)
}

func TestHSetParser_ParseTooFewArgs(t *testing.T) {
	parser := NewHSetParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "HSET",
		Args:    []string{},
	})
	assert.Error(t, err)
}

func TestHSetParser_ParseMissingValue(t *testing.T) {
	parser := NewHSetParser()
	_, err := parser.Parse(&protocol.Message{
		Command: "HSET",
		Args:    []string{"myhash", "field1"},
	})
	assert.Error(t, err)
}

func TestHSetParser_Name(t *testing.T) {
	parser := NewHSetParser()
	assert.Equal(t, "HSET", parser.Name())
}
