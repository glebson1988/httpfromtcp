package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/glebson1988/httpfromtcp/internal/request"
	"github.com/glebson1988/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

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
		log.Println("Error reading request:", err)
		return
	}

	var body bytes.Buffer
	handlerErr := s.handler(&body, req)
	statusCode := response.StatusOK
	responseBody := body.Bytes()
	if handlerErr != nil {
		statusCode = handlerErr.StatusCode
		responseBody = []byte(handlerErr.Message)
	}

	if err := response.WriteStatusLine(conn, statusCode); err != nil {
		log.Println("Error writing status line:", err)
		return
	}
	if err := response.WriteHeaders(conn, response.GetDefaultHeaders(len(responseBody))); err != nil {
		log.Println("Error writing headers:", err)
		return
	}
	if _, err := conn.Write(responseBody); err != nil {
		log.Println("Error writing body:", err)
		return
	}
}
