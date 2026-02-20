package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type LPop struct {
	Key      string
	Count    int
	HasCount bool
}

func (l *LPop) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	values, err := storage.Lists().LPop(ctx, l.Key, l.Count)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	if values == nil {
		return protocol.NewNullBulkStringResponse()
	}
	if !l.HasCount {
		return protocol.NewBulkStringResponse(values[0])
	}
	array := make([]protocol.Value, len(values))
	for i, v := range values {
		array[i] = protocol.NewBulkStringProtocolValue(v)
	}
	return protocol.NewSuccessResponse(protocol.NewArrayProtocolValue(array))
}

type LPopParser struct {
}

func NewLPopParser() *LPopParser {
	return &LPopParser{}
}

func (l *LPopParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 1 || len(msg.Args) > 2 {
		return nil, command.NewInvalidArgumentsCount("LPOP", 1, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError("LPOP", "key")
	}
	count := 1
	hasCount := false
	if len(msg.Args) == 2 {
		c, err := msg.Args[1].AsInt64()
		if err != nil {
			return nil, command.NewInvalidTypeError("LPOP", "count")
		}
		count = int(c)
		hasCount = true
	}
	return &LPop{Key: key, Count: count, HasCount: hasCount}, nil
}

func (l *LPopParser) Name() string {
	return "LPOP"
}
