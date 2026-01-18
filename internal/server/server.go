package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/glebson1988/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	srv := &Server{listener: listener}
	go srv.listen()
	return srv, nil
}

func (s *Server) Close() error {
	if s == nil {
		return nil
	}
	if s.closed.Swap(true) {
		return nil
	}
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Println("Error accepting connection:", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		log.Println("Error writing status line:", err)
		return
	}
	if err := response.WriteHeaders(conn, response.GetDefaultHeaders(0)); err != nil {
		log.Println("Error writing headers:", err)
		return
	}
}
