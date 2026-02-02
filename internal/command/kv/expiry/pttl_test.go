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

func TestPTTL_ExecuteHandlesError(t *testing.T) {
	controller := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(controller)
	store := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().GetTTL("key1").Return(int64(-2), memory.NewKeyNotPresentError("key1"))

	pttl := &PTTL{Key: "key1"}
	response := pttl.Execute(context.Background(), storage)
	assert.NoError(t, response.Err)
	assert.Equal(t, protocol.NewNumberResponse(int64(-2)), response)
}

func TestPTTL_ExecuteHandlesNegativeTTL(t *testing.T) {
	controller := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(controller)
	store := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().GetTTL("key1").Return(int64(-1), nil)

	pttl := &PTTL{Key: "key1"}
	response := pttl.Execute(context.Background(), storage)
	assert.NoError(t, response.Err)
	assert.Equal(t, protocol.NewNumberResponse(int64(-1)), response)
}

func TestPTTL_ExecutePositiveTTL(t *testing.T) {
	controller := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(controller)
	store := mockkv.NewMockStore(controller)

	storage.EXPECT().KV().Return(store).AnyTimes()
	store.EXPECT().GetTTL("key1").Return(int64(1800), nil)

	pttl := &PTTL{Key: "key1"}
	response := pttl.Execute(context.Background(), storage)
	assert.NoError(t, response.Err)
	assert.Equal(t, protocol.NewNumberResponse(int64(1800)), response)

	store.EXPECT().GetTTL("key2").Return(int64(800), nil)
	pttl2 := &PTTL{Key: "key2"}
	response2 := pttl2.Execute(context.Background(), storage)
	assert.NoError(t, response2.Err)
	assert.Equal(t, protocol.NewNumberResponse(int64(800)), response2)
}

func TestPTTLParser_ParseAValidCommand(t *testing.T) {
	parser := PTTLParser{}
	msg := &protocol.Message{
		Command: "pttl",
		Args: []protocol.Value{
			protocol.NewStringProtocolValue("key"),
		},
	}
	cmd, err := parser.Parse(msg)
	assert.NoError(t, err)
	assert.Equal(t, "key", (*cmd.(*PTTL)).Key)
}

func TestPTTLParser_ParseHandleErrorForMissingKey(t *testing.T) {
	parser := PTTLParser{}
	msg := &protocol.Message{
		Command: "pttl",
		Args:    []protocol.Value{},
	}
	_, err := parser.Parse(msg)
	assert.Error(t, err)
}

func TestPTTLParser_ParseHandleErrorForIncorrectKeyDataType(t *testing.T) {
	parser := PTTLParser{}
	msg := &protocol.Message{
		Command: "pttl",
		Args: []protocol.Value{
			protocol.NewNumberProtocolValue(45),
		},
	}
	_, err := parser.Parse(msg)
	assert.Error(t, err)
}
