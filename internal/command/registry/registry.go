package registry

import (
	"avacado/internal/command"
	"avacado/internal/command/connection"
	"avacado/internal/command/connection/client"
	"avacado/internal/command/kv"
	"avacado/internal/command/kv/expiry"
	"avacado/internal/protocol"
	"strings"
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
	registry.Register(expiry.NewTTLParser())
	registry.Register(expiry.NewPTTLParser())
	registry.Register(connection.NewHelloParser())
	registry.Register(client.NewClientParser())
	registry.Register(kv.NewIncrParser())
	registry.Register(kv.NewDecrParser())

	return registry
}

// Register registers a new parser
func (d *DefaultParserRegistry) Register(parser command.Parser) {
	d.parsers[strings.ToUpper(parser.Name())] = parser
}

// Parse parses a raw message to a redis command
func (d *DefaultParserRegistry) Parse(msg *protocol.Message) (command.Command, error) {
	parser, ok := d.parsers[strings.ToUpper(msg.Command)]
	if !ok {
		return nil, protocol.NewUnknowCommandError(msg.Command)
	}
	return parser.Parse(msg)
}
