package server

import (
	"avacado/internal/protocol"
	"net"
)

// Server handles the io connections
type Server struct {
	protocol protocol.Protocol
}

// NewServer creates a new server
func NewServer(protocol protocol.Protocol) *Server {
	return &Server{protocol: protocol}
}

func (s *Server) Serve(conn net.Conn) error {
	//message, err := s.protocol.Parse(conn)
	//if err != nil {
	//	return err
	//}
	//
	return nil
}
