package observability

import (
	"log/slog"
	"os"
)

type LoggerConfig struct {
	Level  slog.Level
	Format string
}

func NewLogger(config LoggerConfig) *slog.Logger {
	var handler slog.Handler
	options := slog.HandlerOptions{
		Level: config.Level,
	}
	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &options)
	}
	return slog.New(handler)
}

func NewNoOutLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
