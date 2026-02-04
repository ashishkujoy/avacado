package connection

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type Hello struct {
}

func (h *Hello) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	return protocol.NewMapResponse([]protocol.MapEntry{
		{Key: "server", Val: protocol.NewBulkStringProtocolValue([]byte("Avacado"))},
		{Key: "version", Val: protocol.NewBulkStringProtocolValue([]byte("0.1.0"))},
		{Key: "proto", Val: protocol.NewNumberProtocolValue(2)},
		{Key: "mode", Val: protocol.NewBulkStringProtocolValue([]byte("standalone"))},
		{Key: "role", Val: protocol.NewBulkStringProtocolValue([]byte("master"))},
	})
}

type HelloParser struct {
}

func NewHelloParser() *HelloParser {
	return &HelloParser{}
}

func (h *HelloParser) Parse(msg *protocol.Message) (command.Command, error) {
	return &Hello{}, nil
}

func (h *HelloParser) Name() string {
	return "HELLO"
}
