package expiry

import (
	"avacado/internal/protocol"
	"avacado/internal/storage/kv/memory"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTTL_ExecuteHandlesError(t *testing.T) {
	controller := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(controller)
	store := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().GetTTL("key1").Return(int64(-2), memory.NewKeyNotPresentError("key1"))

	ttl := &TTL{Key: "key1"}
	response := ttl.Execute(context.Background(), storage)
	assert.NoError(t, response.Err)
	assert.Equal(t, protocol.NewNumberResponse(int64(-2)), response)
}

func TestTTL_ExecuteHandlesNegativeTTL(t *testing.T) {
	controller := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(controller)
	store := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().GetTTL("key1").Return(int64(-1), nil)

	ttl := &TTL{Key: "key1"}
	response := ttl.Execute(context.Background(), storage)
	assert.NoError(t, response.Err)
	assert.Equal(t, protocol.NewNumberResponse(int64(-1)), response)
}

func TestTTL_ExecutePositiveTTL(t *testing.T) {
	controller := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(controller)
	store := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(store).AnyTimes()
	store.EXPECT().GetTTL("key1").Return(int64(1800), nil)

	ttl := &TTL{Key: "key1"}
	response := ttl.Execute(context.Background(), storage)
	assert.NoError(t, response.Err)
	assert.Equal(t, protocol.NewNumberResponse(int64(1)), response)

	store.EXPECT().GetTTL("key2").Return(int64(800), nil)
	ttl2 := &TTL{Key: "key2"}
	response2 := ttl2.Execute(context.Background(), storage)
	assert.NoError(t, response2.Err)
	assert.Equal(t, protocol.NewNumberResponse(int64(0)), response2)
}

func TestTTLParser_ParseAValidCommand(t *testing.T) {
	parser := TTLParser{}
	msg := &protocol.Message{
		Command: "ttl",
		Args: []protocol.Value{
			protocol.NewStringProtocolValue("key"),
		},
	}
	cmd, err := parser.Parse(msg)
	assert.NoError(t, err)
	assert.Equal(t, "key", (*cmd.(*TTL)).Key)
}

func TestTTLParser_ParseHandleErrorForMissingKey(t *testing.T) {
	parser := TTLParser{}
	msg := &protocol.Message{
		Command: "ttl",
		Args:    []protocol.Value{},
	}
	_, err := parser.Parse(msg)
	assert.Error(t, err)
}

func TestTTLParser_ParseHandleErrorForIncorrectKeyDataType(t *testing.T) {
	parser := TTLParser{}
	msg := &protocol.Message{
		Command: "ttl",
		Args: []protocol.Value{
			protocol.NewNumberProtocolValue(45),
		},
	}
	_, err := parser.Parse(msg)
	assert.Error(t, err)
}
