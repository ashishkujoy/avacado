package connection

import (
	"avacado/internal/command"
	"avacado/internal/config"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"errors"
	"fmt"
	"strconv"
)

type Hello struct {
	Proto int
}

func (h *Hello) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	clientConfig := ctx.Value("clientConfig")
	if clientConfig == nil {
		fmt.Printf("missing client config")
		return protocol.NewErrorResponse(errors.New("internal server error"))
	}
	cc := clientConfig.(*config.ClientConfig)
	cc.ProtocolVersion = h.Proto
	return protocol.NewMapResponse([]protocol.MapEntry{
		{Key: "server", Val: protocol.NewBulkStringProtocolValue([]byte("Avacado"))},
		{Key: "version", Val: protocol.NewBulkStringProtocolValue([]byte("0.1.0"))},
		{Key: "proto", Val: protocol.NewNumberProtocolValue(int64(h.Proto))},
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
	proto := 2
	if len(msg.Args) > 0 {
		protoStr, err := msg.Args[0].AsString()
		if err != nil {
			return nil, command.NewInvalidTypeError(h.Name(), "Proto")
		}
		p, err := strconv.Atoi(protoStr)
		if err != nil {
			return nil, command.NewInvalidTypeError(h.Name(), "Proto")
		}
		if p > 3 {
			return nil, command.NewInvalidTypeError(h.Name(), "Proto")
		}
		proto = p
	}
	return &Hello{Proto: proto}, nil
}

func (h *HelloParser) Name() string {
	return "HELLO"
}
