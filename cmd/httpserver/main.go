package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/glebson1988/httpfromtcp/internal/request"
	"github.com/glebson1988/httpfromtcp/internal/response"
	"github.com/glebson1988/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Your problem is not my problem\n",
			}
		case "/myproblem":
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    "Woopsie, my bad\n",
			}
		default:
			_, err := io.WriteString(w, "All good, frfr\n")
			if err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    "Woopsie, my bad\n",
				}
			}
			return nil
		}
	}

	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
