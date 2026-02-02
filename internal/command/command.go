package command

import (
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"fmt"
)

// Command represent a redis command.
//
//go:generate sh -c "rm -f mock/command.go && mockgen -source=command.go -destination=mock/command.go -package=mockcommand"
type Command interface {
	Execute(ctx context.Context, storage storage.Storage) *protocol.Response
}

// Parser parses a raw message to a redis command
type Parser interface {
	Parse(msg *protocol.Message) (Command, error)
	Name() string
}

// ParserRegistry represent a registry for parsers
type ParserRegistry interface {
	Register(parser Parser)
	Parse(msg *protocol.Message) (Command, error)
}

// Validator validate the command arguments
type Validator interface {
	Validate(msg protocol.Message) error
}

// NewInvalidArgumentsCount create a new error to be return when there is arguments counts mismatch while parsing command
func NewInvalidArgumentsCount(name string, expectedCount int, actualCount int) error {
	return fmt.Errorf("%s parse error, expected count <%d> actual count <%d>", name, expectedCount, actualCount)
}

// NewInvalidTypeError create a new error to be return when there is incorrect type of argument
func NewInvalidTypeError(name string, field string) error {
	return fmt.Errorf("%s parse error, incorrect option type %s", name, field)
}
