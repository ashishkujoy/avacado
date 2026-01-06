package main

import (
	"avacado/internal/observability"
	"bufio"
	"flag"
	"fmt"
	"log/slog"
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
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("failed to listen on port", "port", port, "error", err.Error())
		os.Exit(1)
	}
	logger.Info("server started listening", "port", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Debug("failed to accept connection", err)
		}
		go handleConnection(conn, logger)
	}
}

func handleConnection(conn net.Conn, logger *slog.Logger) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	bytes, err := reader.ReadBytes('\n')
	if err != nil {
		logger.Error("failed to read bytes from connection", "error", err.Error())
		return
	}
	_, _ = bufio.NewWriter(conn).Write(bytes)
}
