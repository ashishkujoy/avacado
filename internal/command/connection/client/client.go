package client

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

type Client struct {
}

func (c *Client) Execute(ctx context.Context, storage storage.Storage) *protocol.Response {
	return protocol.NewSimpleStringResponse("OK")
}

type Parser struct {
}

func (p *Parser) Parse(msg *protocol.Message) (command.Command, error) {
	return &Client{}, nil
}

func (p *Parser) Name() string {
	return "CLIENT"
}

func NewClientParser() *Parser {
	return &Parser{}
}
