package registry

import (
	"avacado/internal/command"
	"avacado/internal/command/kv"
	"avacado/internal/protocol"
)

// DefaultParserRegistry registers parsers in memory
type DefaultParserRegistry struct {
	parsers map[string]command.Parser
}

func SetupDefaultParserRegistry() *DefaultParserRegistry {
	registry := &DefaultParserRegistry{
		parsers: make(map[string]command.Parser),
	}

	registry.Register(kv.NewSetParser())
	registry.Register(kv.NewGetParser())
	registry.Register(kv.NewTTLParser())

	return registry
}

// Register registers a new parser
func (d *DefaultParserRegistry) Register(parser command.Parser) {
	d.parsers[parser.Name()] = parser
}

// Parse parses a raw message to a redis command
func (d *DefaultParserRegistry) Parse(msg *protocol.Message) (command.Command, error) {
	parser, ok := d.parsers[msg.Command]
	if !ok {
		return nil, protocol.NewUnknowCommandError(msg.Command)
	}
	return parser.Parse(msg)
}
