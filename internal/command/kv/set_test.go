package kv

import (
	"avacado/internal/protocol"
	kv2 "avacado/internal/storage/kv"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
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

func TestSetParser_WithNXOption(t *testing.T) {
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
			{
				Type:  protocol.TypeBulkString,
				Bytes: []byte("NX"),
			},
		},
	}
	parser := NewSetParser()
	command, err := parser.Parse(msg)
	assert.NoError(t, err)
	assert.Equal(t, "key", (*command.(*Set)).Key)
	assert.Equal(t, "value", string((*command.(*Set)).Value))
	assert.True(t, (*command.(*Set)).Options.NX)
}

func TestSetParser_WithXXOption(t *testing.T) {
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
			{
				Type:  protocol.TypeBulkString,
				Bytes: []byte("XX"),
			},
		},
	}
	parser := NewSetParser()
	command, err := parser.Parse(msg)
	assert.NoError(t, err)
	assert.Equal(t, "key", (*command.(*Set)).Key)
	assert.Equal(t, "value", string((*command.(*Set)).Value))
	assert.True(t, (*command.(*Set)).Options.XX)
}

func TestSet_ExecuteSuccessfully(t *testing.T) {
	controller := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(controller)
	kv := mockkv.NewMockStore(controller)
	storage.EXPECT().KV().Return(kv)

	ctx := context.Background()
	value := []byte("value")
	kv.EXPECT().Set(ctx, "key", value, kv2.NewSetOptions()).Return(nil)

	command := &Set{
		Key:     "key",
		Value:   value,
		Options: kv2.NewSetOptions(),
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
	kv.EXPECT().Set(ctx, "key", value, gomock.Any()).Return(fmt.Errorf("some error"))

	command := &Set{
		Key:   "key",
		Value: value,
	}
	response := command.Execute(ctx, storage)
	assert.Equal(t, protocol.NewNullBulkStringResponse(), response)
}
