package connection

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type Ping struct {
	Message string
}

func (p *Ping) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	if p.Message != "" {
		return protocol.NewBulkStringResponse([]byte(p.Message))
	}
	return protocol.NewSimpleStringResponse("PONG")
}

type PingParser struct{}

func NewPingParser() *PingParser {
	return &PingParser{}
}

func (p *PingParser) Parse(msg *protocol.Message) (command.Command, error) {
	if len(msg.Args) > 0 {
		return &Ping{Message: msg.Args[0]}, nil
	}
	return &Ping{}, nil
}

func (p *PingParser) Name() string {
	return "PING"
}
