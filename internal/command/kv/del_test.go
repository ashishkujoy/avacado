package kv

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDelParser_ParseSingleKey(t *testing.T) {
	d := &DelParser{}
	cmd, err := d.Parse(&protocol.Message{
		Command: "DEL",
		Args: []protocol.Value{
			protocol.NewStringProtocolValue("key1"),
		},
	})
	assert.NoError(t, err)
	delCmd := cmd.(*Del)
	assert.Equal(t, []string{"key1"}, delCmd.Keys)
}

func TestDelParser_ParseMultipleKeys(t *testing.T) {
	d := &DelParser{}
	cmd, err := d.Parse(&protocol.Message{
		Command: "DEL",
		Args: []protocol.Value{
			protocol.NewStringProtocolValue("key1"),
			protocol.NewStringProtocolValue("key2"),
			protocol.NewStringProtocolValue("key3"),
		},
	})
	assert.NoError(t, err)
	delCmd := cmd.(*Del)
	assert.Equal(t, []string{"key1", "key2", "key3"}, delCmd.Keys)
}

func TestDel_Execute(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Del(gomock.Any(), "key1", "key2").Return(int64(2), nil)

	cmd := &Del{Keys: []string{"key1", "key2"}}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewNumberResponse(2), resp)
}

func TestDel_ExecuteHandlesError(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Del(gomock.Any(), "key1").Return(int64(0), assert.AnError)

	cmd := &Del{Keys: []string{"key1"}}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewErrorResponse(assert.AnError), resp)
}
