package expiry

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

// PTTL returns the time to live for a command in seconds
type PTTL struct {
	Key string
}

func (t *PTTL) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	ttl, err := storage.KV().GetTTL(t.Key)
	if err != nil {
		return protocol.NewNumberResponse(-2)
	}
	if ttl == -1 {
		return protocol.NewNumberResponse(-1)
	}
	return protocol.NewNumberResponse(ttl)
}

// PTTLParser parses protocol message to PTTL command
type PTTLParser struct {
}

func NewPTTLParser() *PTTLParser {
	return &PTTLParser{}
}

func (t *PTTLParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 1 {
		return nil, command.NewInvalidArgumentsCount(t.Name(), 1, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(t.Name(), "key")
	}
	return &PTTL{Key: key}, nil
}

func (t *PTTLParser) Name() string {
	return "PTTL"
}
