package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type PopDirection int

const (
	PopLeft PopDirection = iota
	PopRight
)

type Pop struct {
	Key       string
	Count     int
	HasCount  bool
	Direction PopDirection
}

func (p *Pop) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	var values [][]byte
	var err error
	if p.Direction == PopLeft {
		values, err = storage.Lists().LPop(ctx, p.Key, p.Count)
	} else {
		values, err = storage.Lists().RPop(ctx, p.Key, p.Count)
	}
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	if values == nil {
		return protocol.NewNullBulkStringResponse()
	}
	if !p.HasCount {
		return protocol.NewBulkStringResponse(values[0])
	}
	array := make([]protocol.Value, len(values))
	for i, v := range values {
		array[i] = protocol.NewBulkStringProtocolValue(v)
	}
	return protocol.NewSuccessResponse(protocol.NewArrayProtocolValue(array))
}

func parsePop(name string, direction PopDirection, msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) < 1 || len(msg.Args) > 2 {
		return nil, command.NewInvalidArgumentsCount(name, 1, len(msg.Args))
	}
	key, err := msg.Args[0].AsString()
	if err != nil {
		return nil, command.NewInvalidTypeError(name, "key")
	}
	count := 1
	hasCount := false
	if len(msg.Args) == 2 {
		c, err := msg.Args[1].AsInt64()
		if err != nil {
			return nil, command.NewInvalidTypeError(name, "count")
		}
		count = int(c)
		hasCount = true
	}
	return &Pop{Key: key, Count: count, HasCount: hasCount, Direction: direction}, nil
}

type LPopParser struct{}

func NewLPopParser() *LPopParser {
	return &LPopParser{}
}

func (l *LPopParser) Parse(msg *protocol.Message) (command.Command, error) {
	return parsePop("LPOP", PopLeft, msg)
}

func (l *LPopParser) Name() string {
	return "LPOP"
}

type RPopParser struct{}

func NewRPopParser() *RPopParser {
	return &RPopParser{}
}

func (r *RPopParser) Parse(msg *protocol.Message) (command.Command, error) {
	return parsePop("RPOP", PopRight, msg)
}

func (r *RPopParser) Name() string {
	return "RPOP"
}
