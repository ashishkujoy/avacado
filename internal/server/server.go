package server

import (
	"avacado/internal/command"
	"avacado/internal/protocol"
	"avacado/internal/storage"
	"context"
	"io"
	"log/slog"
)

// Server handles the io connections
type Server struct {
	protocol protocol.Protocol
	parser   command.ParserRegistry
	storage  storage.Storage
}

// NewServer creates a new server
func NewServer(
	protocol protocol.Protocol,
	parser command.ParserRegistry,
	storage storage.Storage,
) *Server {
	return &Server{
		protocol: protocol,
		parser:   parser,
		storage:  storage,
	}
}

type Connection interface {
	io.Reader
	io.Writer
	io.Closer
}

func (s *Server) Serve(conn Connection, logger *slog.Logger) error {
	ctx := context.Background()
	defer func() {
		logger.Info("closing connection")
		conn.Close()
	}()
	parser := s.protocol.CreateParser(conn)
	for {
		message, err := parser.Parse()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Error("failed to parse message", "error", err.Error())
			_, _ = conn.Write(s.protocol.SerializeError(err))
			return err
		}
		cmd, err := s.parser.Parse(message)
		if err != nil {
			logger.Error("failed to parse command", "error", err.Error())
			_, _ = conn.Write(s.protocol.SerializeError(err))
			return err
		}
		response := cmd.Execute(ctx, s.storage)
		if response.Err != nil {
			logger.Error("failed to execute command", "error", response.Err.Error())
			_, _ = conn.Write(s.protocol.SerializeError(response.Err))
			return response.Err
		}
		bytes, err := s.protocol.Serialize(response)
		if err != nil {
			logger.Error("failed to serialize response", "error", err.Error())
			return err
		}
		_, _ = conn.Write(bytes)
	}
}
