package server

import (
	"avacado/internal/command"
	config "avacado/internal/config"
	"avacado/internal/executor"
	"avacado/internal/protocol"
	"context"
	"io"
	"log/slog"
)

// Server handles the io connections
type Server struct {
	protocol protocol.Protocol
	parser   command.ParserRegistry
	executor *executor.Executor
}

// NewServer creates a new server
func NewServer(
	protocol protocol.Protocol,
	parser command.ParserRegistry,
	exec *executor.Executor,
) *Server {
	return &Server{
		protocol: protocol,
		parser:   parser,
		executor: exec,
	}
}

type Connection interface {
	io.Reader
	io.Writer
	io.Closer
}

func (s *Server) Serve(conn Connection, logger *slog.Logger) error {
	clientConfig := config.DefaultClientConfig()
	ctx := context.WithValue(context.Background(), "clientConfig", clientConfig)
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
		response := s.executor.Submit(ctx, cmd)
		if response.Err != nil {
			logger.Error("failed to execute command", "error", response.Err.Error())
			_, _ = conn.Write(s.protocol.SerializeError(response.Err))
			return response.Err
		}

		// Blocking commands (BLPOP/BRPOP) return a non-nil BlockCh when no data
		// is immediately available. Wait for the result on that channel.
		if response.BlockCh != nil {
			select {
			case finalResp := <-response.BlockCh:
				response = finalResp
			case <-ctx.Done():
				return nil
			}
		}

		bytes, err := s.protocol.Serialize(response)
		if err != nil {
			logger.Error("failed to serialize response", "error", err.Error())
			return err
		}
		_, _ = conn.Write(bytes)
	}
}
