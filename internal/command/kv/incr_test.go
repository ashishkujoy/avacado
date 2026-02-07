package kv

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestIncrParser_Parse(t *testing.T) {
	i := &IncrParser{}
	_, err := i.Parse(&protocol.Message{
		Command: "INCR",
		Args: []protocol.Value{
			protocol.NewStringProtocolValue("key"),
		},
	})
	assert.NoError(t, err)
}

func TestIncr_Execute(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Incr(gomock.Any(), "key").Return(int64(1), nil)

	cmd := &Incr{Key: "key"}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewNumberResponse(1), resp)
}

func TestIncr_ExecuteHandlesError(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Incr(gomock.Any(), "key").Return(int64(0), assert.AnError)

	cmd := &Incr{Key: "key"}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewErrorResponse(assert.AnError), resp)
}
