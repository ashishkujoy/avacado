package config

import "context"

type ClientConfig struct {
	ProtocolVersion int
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{ProtocolVersion: 2}
}

func IsProto3(ctx context.Context) bool {
	cc := ctx.Value("clientConfig")
	config := cc.(*ClientConfig)
	return config.ProtocolVersion == 3
}
