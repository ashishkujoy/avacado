package server

import (
	mockcommand "avacado/internal/command/mock"
	"avacado/internal/executor"
	"avacado/internal/observability"
	protocol2 "avacado/internal/protocol"
	mockprotocol "avacado/internal/protocol/mock"
	mocksstorage "avacado/internal/storage/mock"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestServer_HandlesErrorOnProtocolMessageParsing(t *testing.T) {
	controller := gomock.NewController(t)
	proto := mockprotocol.NewMockProtocol(controller)
	store := mocksstorage.NewMockStorage(controller)
	registry := mockcommand.NewMockParserRegistry(controller)
	parser := mockprotocol.NewMockParser(controller)

	exec := executor.New(store)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go exec.Run(ctx)

	s := NewServer(proto, registry, exec)
	logger := observability.NewNoOutLogger()

	connection := &mockConnection{
		dataToRead:  []byte("input data"),
		readCursor:  0,
		dataWritten: []byte{},
	}

	proto.EXPECT().CreateParser(connection).Return(parser)
	parser.EXPECT().Parse().Return(nil, fmt.Errorf("protocol parse error"))
	proto.EXPECT().SerializeError(gomock.Any()).Return([]byte("-protocol parse error\r\n"))

	s.Serve(connection, logger)
	assert.Equal(t, "-protocol parse error\r\n", string(connection.dataWritten))
}

func TestServer_HandlesErrorOnCommandParsing(t *testing.T) {
	controller := gomock.NewController(t)
	proto := mockprotocol.NewMockProtocol(controller)
	store := mocksstorage.NewMockStorage(controller)
	registry := mockcommand.NewMockParserRegistry(controller)
	parser := mockprotocol.NewMockParser(controller)

	exec := executor.New(store)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go exec.Run(ctx)

	s := NewServer(proto, registry, exec)
	logger := observability.NewNoOutLogger()

	connection := &mockConnection{
		dataToRead:  []byte("input data"),
		readCursor:  0,
		dataWritten: []byte{},
	}
	msg := &protocol2.Message{}
	proto.EXPECT().CreateParser(connection).Return(parser)
	parser.EXPECT().Parse().Return(msg, nil)
	registry.EXPECT().Parse(gomock.Any()).Return(nil, fmt.Errorf("command parse error"))
	proto.EXPECT().SerializeError(gomock.Any()).Return([]byte("-command parse error\r\n"))

	s.Serve(connection, logger)
	assert.Equal(t, "-command parse error\r\n", string(connection.dataWritten))
}

func TestServer_HandlesCommandExecutionError(t *testing.T) {
	controller := gomock.NewController(t)
	proto := mockprotocol.NewMockProtocol(controller)
	store := mocksstorage.NewMockStorage(controller)
	registry := mockcommand.NewMockParserRegistry(controller)
	cmd := mockcommand.NewMockCommand(controller)
	parser := mockprotocol.NewMockParser(controller)
	resp := protocol2.NewErrorResponse(fmt.Errorf("command Execution fail"))

	exec := executor.New(store)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go exec.Run(ctx)

	s := NewServer(proto, registry, exec)
	logger := observability.NewNoOutLogger()

	connection := &mockConnection{
		dataToRead:  []byte("input data"),
		readCursor:  0,
		dataWritten: []byte{},
	}
	msg := &protocol2.Message{}
	proto.EXPECT().CreateParser(connection).Return(parser)
	parser.EXPECT().Parse().Return(msg, nil)
	registry.EXPECT().Parse(gomock.Any()).Return(cmd, nil)
	cmd.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(resp)
	proto.EXPECT().SerializeError(gomock.Any()).Return([]byte("-command Execution fail\r\n"))

	s.Serve(connection, logger)
	assert.Equal(t, "-command Execution fail\r\n", string(connection.dataWritten))
}

func TestServer_HandlesCommandExecutionSuccess(t *testing.T) {
	controller := gomock.NewController(t)
	proto := mockprotocol.NewMockProtocol(controller)
	registry := mockcommand.NewMockParserRegistry(controller)
	store := mocksstorage.NewMockStorage(controller)
	cmd := mockcommand.NewMockCommand(controller)
	resp := protocol2.NewSuccessResponse(protocol2.NewStringProtocolValue("OK"))
	parser := mockprotocol.NewMockParser(controller)

	exec := executor.New(store)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go exec.Run(ctx)

	s := NewServer(proto, registry, exec)
	logger := observability.NewNoOutLogger()

	connection := &mockConnection{
		dataToRead:  []byte("input data"),
		readCursor:  0,
		dataWritten: []byte{},
	}
	msg := &protocol2.Message{}
	proto.EXPECT().CreateParser(connection).Return(parser)
	parser.EXPECT().Parse().Return(msg, nil)
	registry.EXPECT().Parse(gomock.Any()).Return(cmd, nil)
	cmd.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(resp)
	proto.EXPECT().Serialize(gomock.Any()).Return([]byte("-command success\r\n"), nil)
	parser.EXPECT().Parse().Return(nil, io.EOF)

	s.Serve(connection, logger)
	assert.Equal(t, "-command success\r\n", string(connection.dataWritten))
}

type mockConnection struct {
	dataToRead  []byte
	readCursor  int
	dataWritten []byte
}

func (m *mockConnection) Read(p []byte) (n int, err error) {
	if len(m.dataToRead) == 0 {
		return 0, io.EOF
	}
	n = copy(p, m.dataToRead[m.readCursor:])
	m.readCursor += n
	return n, nil
}

func (m *mockConnection) Write(p []byte) (n int, err error) {
	m.dataWritten = append(m.dataWritten, p...)
	return len(p), nil
}

func (m *mockConnection) Close() error {
	return nil
}
