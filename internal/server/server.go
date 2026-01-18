package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/glebson1988/httpfromtcp/internal/request"
	"github.com/glebson1988/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	srv := &Server{
		listener: listener,
		handler:  handler,
	}
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		writer := response.NewWriter(conn)
		msg := []byte(err.Error())
		if err := writer.WriteStatusLine(response.StatusBadRequest); err != nil {
			log.Println("Error writing status line:", err)
			return
		}
		headers := response.GetDefaultHeaders(len(msg))
		if err := writer.WriteHeaders(headers); err != nil {
			log.Println("Error writing headers:", err)
			return
		}
		if _, err := writer.WriteBody(msg); err != nil {
			log.Println("Error writing body:", err)
			return
		}
		return
	}

	writer := response.NewWriter(conn)
	s.handler(writer, req)
}
