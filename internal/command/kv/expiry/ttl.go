package expiry

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

// TTL returns the time to live for a command in seconds
type TTL struct {
	Key string
}

func (t *TTL) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	ttl, err := storage.KV().GetTTL(t.Key)
	if err != nil {
		return protocol.NewNumberResponse(-2)
	}
	if ttl == -1 {
		return protocol.NewNumberResponse(-1)
	}
	ttlInSeconds := ttl / 1000
	return protocol.NewNumberResponse(ttlInSeconds)
}

// TTLParser parses protocol message to TTL command
type TTLParser struct {
}

func NewTTLParser() *TTLParser {
	return &TTLParser{}
}

func (t *TTLParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 1 {
		return nil, command.NewInvalidArgumentsCount(t.Name(), 1, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(t.Name(), "key")
	}
	return &TTL{Key: key}, nil
}

func (t *TTLParser) Name() string {
	return "TTL"
}
