package kv

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestStrlenParser_Parse(t *testing.T) {
	p := &StrlenParser{}
	cmd, err := p.Parse(&protocol.Message{
		Command: "STRLEN",
		Args:    []string{"mykey"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "mykey", cmd.(*Strlen).Key)
}

func TestStrlen_Execute(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Len(gomock.Any(), "mykey").Return(int64(5), nil)

	cmd := &Strlen{Key: "mykey"}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewNumberResponse(5), resp)
}

func TestStrlen_ExecuteHandlesError(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Len(gomock.Any(), "mykey").Return(int64(0), assert.AnError)

	cmd := &Strlen{Key: "mykey"}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewErrorResponse(assert.AnError), resp)
}
