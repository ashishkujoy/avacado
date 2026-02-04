package server

import (
	mockcommand "avacado/internal/command/mock"
	"avacado/internal/observability"
	protocol2 "avacado/internal/protocol"
	mockprotocol "avacado/internal/protocol/mock"
	mocksstorage "avacado/internal/storage/mock"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestServer_HandlesErrorOnProtocolMessageParsing(t *testing.T) {
	controller := gomock.NewController(t)
	protocol := mockprotocol.NewMockProtocol(controller)
	storage := mocksstorage.NewMockStorage(controller)
	registry := mockcommand.NewMockParserRegistry(controller)
	parser := mockprotocol.NewMockParser(controller)

	server := NewServer(protocol, registry, storage)
	logger := observability.NewNoOutLogger()

	connection := &mockConnection{
		dataToRead:  []byte("input data"),
		readCursor:  0,
		dataWritten: []byte{},
	}

	protocol.EXPECT().CreateParser(connection).Return(parser)
	parser.EXPECT().Parse().Return(nil, fmt.Errorf("protocol parse error"))
	protocol.EXPECT().SerializeError(gomock.Any()).Return([]byte("-protocol parse error\r\n"))

	server.Serve(connection, logger)
	assert.Equal(t, "-protocol parse error\r\n", string(connection.dataWritten))
}

func TestServer_HandlesErrorOnCommandParsing(t *testing.T) {
	controller := gomock.NewController(t)
	protocol := mockprotocol.NewMockProtocol(controller)
	storage := mocksstorage.NewMockStorage(controller)
	registry := mockcommand.NewMockParserRegistry(controller)
	parser := mockprotocol.NewMockParser(controller)

	server := NewServer(protocol, registry, storage)
	logger := observability.NewNoOutLogger()

	connection := &mockConnection{
		dataToRead:  []byte("input data"),
		readCursor:  0,
		dataWritten: []byte{},
	}
	msg := &protocol2.Message{}
	protocol.EXPECT().CreateParser(connection).Return(parser)
	parser.EXPECT().Parse().Return(msg, nil)
	registry.EXPECT().Parse(gomock.Any()).Return(nil, fmt.Errorf("command parse error"))
	protocol.EXPECT().SerializeError(gomock.Any()).Return([]byte("-command parse error\r\n"))

	server.Serve(connection, logger)
	assert.Equal(t, "-command parse error\r\n", string(connection.dataWritten))
}

func TestServer_HandlesCommandExecutionError(t *testing.T) {
	controller := gomock.NewController(t)
	protocol := mockprotocol.NewMockProtocol(controller)
	storage := mocksstorage.NewMockStorage(controller)
	registry := mockcommand.NewMockParserRegistry(controller)
	cmd := mockcommand.NewMockCommand(controller)
	parser := mockprotocol.NewMockParser(controller)
	resp := protocol2.NewErrorResponse(fmt.Errorf("command Execution fail"))

	server := NewServer(protocol, registry, storage)
	logger := observability.NewNoOutLogger()

	connection := &mockConnection{
		dataToRead:  []byte("input data"),
		readCursor:  0,
		dataWritten: []byte{},
	}
	msg := &protocol2.Message{}
	protocol.EXPECT().CreateParser(connection).Return(parser)
	parser.EXPECT().Parse().Return(msg, nil)
	registry.EXPECT().Parse(gomock.Any()).Return(cmd, nil)
	cmd.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(resp)
	protocol.EXPECT().SerializeError(gomock.Any()).Return([]byte("-command Execution fail\r\n"))

	server.Serve(connection, logger)
	assert.Equal(t, "-command Execution fail\r\n", string(connection.dataWritten))
}

func TestServer_HandlesCommandExecutionSuccess(t *testing.T) {
	controller := gomock.NewController(t)
	protocol := mockprotocol.NewMockProtocol(controller)
	registry := mockcommand.NewMockParserRegistry(controller)
	storage := mocksstorage.NewMockStorage(controller)
	cmd := mockcommand.NewMockCommand(controller)
	resp := protocol2.NewSuccessResponse(protocol2.NewStringProtocolValue("OK"))
	parser := mockprotocol.NewMockParser(controller)

	server := NewServer(protocol, registry, storage)
	logger := observability.NewNoOutLogger()

	connection := &mockConnection{
		dataToRead:  []byte("input data"),
		readCursor:  0,
		dataWritten: []byte{},
	}
	msg := &protocol2.Message{}
	protocol.EXPECT().CreateParser(connection).Return(parser)
	parser.EXPECT().Parse().Return(msg, nil)
	registry.EXPECT().Parse(gomock.Any()).Return(cmd, nil)
	cmd.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(resp)
	protocol.EXPECT().Serialize(gomock.Any()).Return([]byte("-command success\r\n"), nil)
	parser.EXPECT().Parse().Return(nil, io.EOF)

	server.Serve(connection, logger)
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
