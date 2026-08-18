package kv

import (
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAppendParser_Parse(t *testing.T) {
	p := &AppendParser{}
	cmd, err := p.Parse(&protocol.Message{
		Command: "APPEND",
		Args:    []string{"mykey", "hello"},
	})
	assert.NoError(t, err)
	a := cmd.(*Append)
	assert.Equal(t, "mykey", a.Key)
	assert.Equal(t, []byte("hello"), a.Value)
}

func TestAppend_Execute(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Append(gomock.Any(), "mykey", []byte("hello")).Return(int64(5), nil)

	cmd := &Append{Key: "mykey", Value: []byte("hello")}
	resp := cmd.Execute(context.TODO(), storage)
	assert.Equal(t, protocol.NewNumberResponse(5), resp)
}

func TestAppend_ExecuteHandlesError(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().Append(gomock.Any(), "mykey", []byte("hello")).Return(int64(0), assert.AnError)

	cmd := &Append{Key: "mykey", Value: []byte("hello")}
	resp := cmd.Execute(context.TODO(), storage)
	assert.Equal(t, protocol.NewErrorResponse(assert.AnError), resp)
}
