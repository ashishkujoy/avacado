package main

import (
	"avacado/internal/command/registry"
	"avacado/internal/observability"
	"avacado/internal/protocol/resp"
	"avacado/internal/server"
	"avacado/internal/storage"
	"flag"
	"fmt"
	"net"
	"os"
)

func main() {
	port := 6379
	flag.IntVar(&port, "port", 6379, "--port")
	flag.Parse()
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
		os.Exit(1)
	}
	logger.Info("server started listening", "port", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Debug("failed to accept connection" + err.Error())
		}
		go s.Serve(conn, logger)
	}
}
