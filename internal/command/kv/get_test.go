package kv

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetCommand_ExecuteForKeyFound(t *testing.T) {
	controller := gomock.NewController(t)
	command := Get{key: "key1"}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	kv := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(kv)
	kv.EXPECT().Get(ctx, "key1").Return([]byte("value1"), nil)

	response := command.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.Equal(t, []byte("value1"), response.Value.Bytes)
}

func TestGetCommand_ExecuteForKeyNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	command := Get{key: "key1"}
	ctx := context.Background()
	storage := mocksstorage.NewMockStorage(controller)
	kv := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(kv)
	kv.EXPECT().Get(ctx, "key1").Return(nil, nil)

	response := command.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
}

func TestGetCommand_ExecuteForKeyError(t *testing.T) {
	controller := gomock.NewController(t)
	command := Get{key: "key1"}
	ctx := context.Background()

	storage := mocksstorage.NewMockStorage(controller)
	kv := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(kv)
	kv.EXPECT().Get(ctx, "key1").Return(nil, fmt.Errorf("some error"))

	response := command.Execute(ctx, storage)
	assert.Nil(t, response.Err)
	assert.True(t, response.Value.Null)
}

func TestGetParser_Parse(t *testing.T) {
	parser := NewGetParser()
	cmd, err := parser.Parse(&protocol.Message{
		Command: "GET",
		Args: []protocol.Value{
			{
				Type:  protocol.TypeBulkString,
				Bytes: []byte("key1"),
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "key1", cmd.(*Get).key)
}
