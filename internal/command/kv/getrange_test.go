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

func TestGetRangeParser_Parse(t *testing.T) {
	p := &GetRangeParser{}
	cmd, err := p.Parse(&protocol.Message{
		Command: "GETRANGE",
		Args:    []string{"mykey", "0", "-1"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "mykey", cmd.(*GetRange).Key)
	assert.Equal(t, int64(0), cmd.(*GetRange).Start)
	assert.Equal(t, int64(-1), cmd.(*GetRange).End)
}

func TestGetRangeParser_ParseWrongArgumentsCount(t *testing.T) {
	p := &GetRangeParser{}
	_, err := p.Parse(&protocol.Message{
		Command: "GETRANGE",
		Args:    []string{"mykey", "0"},
	})
	assert.Equal(t, command.NewInvalidArgumentsCount("GETRANGE", 3, 2), err)
}

func TestGetRangeParser_ParseInvalidStart(t *testing.T) {
	p := &GetRangeParser{}
	_, err := p.Parse(&protocol.Message{
		Command: "GETRANGE",
		Args:    []string{"mykey", "notanumber", "-1"},
	})
	assert.Equal(t, command.NewInvalidTypeError("GETRANGE", "start"), err)
}

func TestGetRangeParser_ParseInvalidEnd(t *testing.T) {
	p := &GetRangeParser{}
	_, err := p.Parse(&protocol.Message{
		Command: "GETRANGE",
		Args:    []string{"mykey", "0", "notanumber"},
	})
	assert.Equal(t, command.NewInvalidTypeError("GETRANGE", "end"), err)
}

func TestGetRange_Execute(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().GetRange(gomock.Any(), "mykey", int64(0), int64(4)).Return([]byte("Hello"), nil)

	cmd := &GetRange{Key: "mykey", Start: 0, End: 4}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewBulkStringResponse([]byte("Hello")), resp)
}

func TestGetRange_ExecuteHandlesError(t *testing.T) {
	ctr := gomock.NewController(t)
	storage := mocksstorage.NewMockStorage(ctr)
	store := mockkv.NewMockStore(ctr)

	storage.EXPECT().KV().Return(store)
	store.EXPECT().GetRange(gomock.Any(), "mykey", int64(0), int64(4)).Return(nil, assert.AnError)

	cmd := &GetRange{Key: "mykey", Start: 0, End: 4}
	resp := cmd.Execute(nil, storage)
	assert.Equal(t, protocol.NewErrorResponse(assert.AnError), resp)
}
