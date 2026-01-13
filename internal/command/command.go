package command

import (
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
)

// Command represent a redis command.
type Command interface {
	Execute(ctx context.Context, storage storage.Storage) *protocol.Response
}

// Parser parses a raw message to a redis command
type Parser interface {
	Parse(msg protocol.Message) (Command, error)
	Name() string
}

// Validator validate the command arguments
type Validator interface {
	Validate(msg protocol.Message) error
}
