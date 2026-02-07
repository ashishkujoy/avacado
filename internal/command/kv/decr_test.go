package kv

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDecrParser_Parse(t *testing.T) {
	d := &DecrParser{}
	_, err := d.Parse(&protocol.Message{
		Command: "DECR",
		Args: []protocol.Value{
			protocol.NewStringProtocolValue("key"),
		},
	})
	assert.NoError(t, err)
}

func TestDecr_Execute(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Decr(gomock.Any(), "key").Return(int64(-1), nil)

	cmd := &Decr{Key: "key"}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewNumberResponse(-1), resp)
}

func TestDecr_ExecuteHandlesError(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Decr(gomock.Any(), "key").Return(int64(0), assert.AnError)

	cmd := &Decr{Key: "key"}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewErrorResponse(assert.AnError), resp)
}
