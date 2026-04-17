package config

type ClientConfig struct {
	ProtocolVersion int
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{ProtocolVersion: 2}
}
