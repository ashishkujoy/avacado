package kv

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	mockkv "avacado/internal/storage/kv/mock"
	mocksstorage "avacado/internal/storage/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSetRangeParser_Parse(t *testing.T) {
	p := &SetRangeParser{}
	cmd, err := p.Parse(&protocol.Message{
		Command: "SETRANGE",
		Args:    []string{"mykey", "6", "Redis"},
	})
	assert.NoError(t, err)
	s := cmd.(*SetRange)
	assert.Equal(t, "mykey", s.Key)
	assert.Equal(t, 6, s.Offset)
	assert.Equal(t, []byte("Redis"), s.Value)
}

func TestSetRangeParser_ParseWrongArgumentsCount(t *testing.T) {
	p := &SetRangeParser{}
	_, err := p.Parse(&protocol.Message{
		Command: "SETRANGE",
		Args:    []string{"mykey", "6"},
	})
	assert.Equal(t, command.NewInvalidArgumentsCount("SETRANGE", 3, 2), err)
}

func TestSetRangeParser_ParseInvalidOffset(t *testing.T) {
	p := &SetRangeParser{}
	_, err := p.Parse(&protocol.Message{
		Command: "SETRANGE",
		Args:    []string{"mykey", "notanumber", "Redis"},
	})
	assert.Error(t, err)
}

func TestSetRange_Execute(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().SetRange(gomock.Any(), "mykey", 6, []byte("Redis")).Return(11, nil)

	cmd := &SetRange{Key: "mykey", Offset: 6, Value: []byte("Redis")}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewNumberResponse(11), resp)
}

func TestSetRange_ExecuteHandlesError(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().SetRange(gomock.Any(), "mykey", 6, []byte("Redis")).Return(0, assert.AnError)

	cmd := &SetRange{Key: "mykey", Offset: 6, Value: []byte("Redis")}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewErrorResponse(assert.AnError), resp)
}
