package kv

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestExistsParser_ParseSingleKey(t *testing.T) {
	e := &ExistsParser{}
	cmd, err := e.Parse(&protocol.Message{
		Command: "EXISTS",
		Args: []protocol.Value{
			protocol.NewStringProtocolValue("key1"),
		},
	})
	assert.NoError(t, err)
	existsCmd := cmd.(*Exists)
	assert.Equal(t, []string{"key1"}, existsCmd.Keys)
}

func TestExistsParser_ParseMultipleKeys(t *testing.T) {
	e := &ExistsParser{}
	cmd, err := e.Parse(&protocol.Message{
		Command: "EXISTS",
		Args: []protocol.Value{
			protocol.NewStringProtocolValue("key1"),
			protocol.NewStringProtocolValue("key2"),
			protocol.NewStringProtocolValue("key3"),
		},
	})
	assert.NoError(t, err)
	existsCmd := cmd.(*Exists)
	assert.Equal(t, []string{"key1", "key2", "key3"}, existsCmd.Keys)
}

func TestExists_Execute(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Exists(gomock.Any(), "key1", "key2").Return(int64(2), nil)

	cmd := &Exists{Keys: []string{"key1", "key2"}}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewNumberResponse(2), resp)
}

func TestExists_ExecuteHandlesError(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Exists(gomock.Any(), "key1").Return(int64(0), assert.AnError)

	cmd := &Exists{Keys: []string{"key1"}}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewErrorResponse(assert.AnError), resp)
}
