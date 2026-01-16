package command

import (
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

// Command represent a redis command.
//
//go:generate mockgen -source=command.go -destination=mock/command.go -package=mockcommand
type Command interface {
	Execute(ctx context.Context, storage storage.Storage) *protocol.Response
}

// Parser parses a raw message to a redis command
//
//go:generate mockgen -source=command.go -destination=mock/command.go -package=mockcommand
type Parser interface {
	Parse(msg *protocol.Message) (Command, error)
	Name() string
}

// ParserRegistry represent a registry for parsers
//
//go:generate mockgen -source=command.go -destination=mock/command.go -package=mockcommand
type ParserRegistry interface {
	Register(parser Parser)
	Parse(msg *protocol.Message) (Command, error)
}

// Validator validate the command arguments
type Validator interface {
	Validate(msg protocol.Message) error
}
