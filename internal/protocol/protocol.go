package protocol

import "io"

// Parser parses the protocol message
type Parser interface {
	Parse(r io.Reader) (*Message, error)
}

// Message represents a protocol message, containing command name and args
type Message struct {
	Command string
	Args    [][]byte
}

// Serializer serializes the protocol message
type Serializer interface {
	Serialize(value interface{}) ([]byte, error)
	SerializeError(e error) []byte
}

type Protocol interface {
	Serializer
	Parser
}
