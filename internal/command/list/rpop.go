package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type RPop struct {
	Key      string
	Count    int
	HasCount bool
}

func (r *RPop) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	values, err := storage.Lists().RPop(ctx, r.Key, r.Count)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	if values == nil {
		return protocol.NewNullBulkStringResponse()
	}
	if !r.HasCount {
		return protocol.NewBulkStringResponse(values[0])
	}
	array := make([]protocol.Value, len(values))
	for i, v := range values {
		array[i] = protocol.NewBulkStringProtocolValue(v)
	}
	return protocol.NewSuccessResponse(protocol.NewArrayProtocolValue(array))
}

type RPopParser struct {
}

func NewRPopParser() *RPopParser {
	return &RPopParser{}
}

func (r *RPopParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 1 || len(msg.Args) > 2 {
		return nil, command.NewInvalidArgumentsCount("RPOP", 1, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError("RPOP", "key")
	}
	count := 1
	if len(msg.Args) == 2 {
		c, err := msg.Args[1].AsInt64()
		if err != nil {
			return nil, command.NewInvalidTypeError("RPOP", "count")
		}
		count = int(c)
	}
	return &RPop{Key: key, Count: count, HasCount: len(msg.Args) == 2}, nil
}

func (r *RPopParser) Name() string {
	return "RPOP"
}
