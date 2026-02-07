package kv

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDecrByParser_Parse(t *testing.T) {
	d := &DecrByParser{}
	cmd, err := d.Parse(&protocol.Message{
		Command: "DECRBY",
		Args: []protocol.Value{
			protocol.NewStringProtocolValue("key"),
			protocol.NewStringProtocolValue("5"),
		},
	})
	assert.NoError(t, err)
	decrByCmd := cmd.(*DecrBy)
	assert.Equal(t, "key", decrByCmd.Key)
	assert.Equal(t, int64(5), decrByCmd.Decrement)
}

func TestDecrBy_Execute(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().DecrBy(gomock.Any(), "key", int64(5)).Return(int64(-5), nil)

	cmd := &DecrBy{Key: "key", Decrement: 5}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewNumberResponse(-5), resp)
}

func TestDecrBy_ExecuteHandlesError(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().DecrBy(gomock.Any(), "key", int64(5)).Return(int64(0), assert.AnError)

	cmd := &DecrBy{Key: "key", Decrement: 5}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewErrorResponse(assert.AnError), resp)
}
