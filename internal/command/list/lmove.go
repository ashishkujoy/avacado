package list

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"avacado/internal/storage/lists"
	"context"
	"fmt"
	"strings"
)

type LMove struct {
	Source               string
	Destination          string
	SourceDirection      lists.Direction
	DestinationDirection lists.Direction
}

func (l *LMove) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	element, err := storage.Lists().LMove(ctx, l.Source, l.Destination, l.SourceDirection, l.DestinationDirection)
	if err != nil {
		return protocol.NewErrorResponse(err)
	}
	if element == nil {
		return protocol.NewNullBulkStringResponse()
	}
	return protocol.NewBulkStringResponse(element)
}

type LMoveParser struct{}

func NewLMoveParser() *LMoveParser {
	return &LMoveParser{}
}

func (l *LMoveParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) != 4 {
		return nil, command.NewInvalidArgumentsCount(l.Name(), 4, len(msg.Args))
	}
	srcDir := strings.ToLower(msg.Args[2])
	dstDir := strings.ToLower(msg.Args[3])
	if srcDir != lists.Left && srcDir != lists.Right {
		return nil, fmt.Errorf("ERR syntax error")
	}
	if dstDir != lists.Left && dstDir != lists.Right {
		return nil, fmt.Errorf("ERR syntax error")
	}
	return &LMove{
		Source:               msg.Args[0],
		Destination:          msg.Args[1],
		SourceDirection:      srcDir,
		DestinationDirection: dstDir,
	}, nil
}

func (l *LMoveParser) Name() string {
	return "LMOVE"
}
