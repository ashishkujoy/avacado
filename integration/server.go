package integration

import (
	"avacado/internal/command/registry"
	"avacado/internal/observability"
	"avacado/internal/protocol/resp"
	"avacado/internal/server"
	"avacado/internal/storage"
	"fmt"
	"net"
)

func StartNewServer(port int64) (func(), error) {
	logger := observability.NewLogger(observability.LoggerConfig{
		Level:  0,
		Format: "json",
	})
	s := server.NewServer(
		resp.NewRespProtocol(),
		registry.SetupDefaultParserRegistry(),
		storage.NewDefaultStorage(),
	)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("failed to listen on port", "port", port, "error", err.Error())
		return func() {}, err
	}
	logger.Info("server started listening", "port", port)
	onServerStart := make(chan interface{})
	go func() {
		onServerStart <- "started"
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Debug("failed to accept connection" + err.Error())
				return
			}
			go s.Serve(conn, logger)
		}
	}()
	<-onServerStart
	return func() {
		listener.Close()
	}, nil
}
