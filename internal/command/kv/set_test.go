package kv

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mocks"
	mocksstorage "avacado/internal/storage/mocks"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSetParser_Parse(t *testing.T) {
	msg := &protocol.Message{
		Command: "SET",
		Args: []protocol.Value{
			{
				Type:  protocol.TypeBulkString,
				Bytes: []byte("key"),
			},
			{
				Type:  protocol.TypeBulkString,
				Bytes: []byte("value"),
			},
		},
	}
	parser := NewSetParser()
	command, err := parser.Parse(msg)
	assert.NoError(t, err)
	assert.Equal(t, "key", (*command.(*Set)).Key)
	assert.Equal(t, "value", string((*command.(*Set)).Value))
}

func TestSet_ExecuteSuccessfully(t *testing.T) {
	controller := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(controller)
	kv := mockkv.NewMockStore(controller)
	storage.EXPECT().KV().Return(kv)

	ctx := context.Background()
	value := []byte("value")
	kv.EXPECT().Set(ctx, "key", value).Return(nil)

	command := &Set{
		Key:   "key",
		Value: value,
	}
	response := command.Execute(ctx, storage)
	assert.Equal(t, protocol.NewSimpleStringResponse("OK"), response)
}

func TestSet_ExecuteWithError(t *testing.T) {
	controller := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(controller)
	kv := mockkv.NewMockStore(controller)
	storage.EXPECT().KV().Return(kv)

	ctx := context.Background()
	value := []byte("value")
	kv.EXPECT().Set(ctx, "key", value).Return(fmt.Errorf("some error"))

	command := &Set{
		Key:   "key",
		Value: value,
	}
	response := command.Execute(ctx, storage)
	assert.Equal(t, protocol.NewNullBulkStringResponse(), response)
}
