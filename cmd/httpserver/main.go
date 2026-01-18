package main

import (
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
	handler := func(w *response.Writer, req *request.Request) {
		var statusCode response.StatusCode
		var body string
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			statusCode = response.StatusBadRequest
			body = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`
		case "/myproblem":
			statusCode = response.StatusInternalServerError
			body = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`
		default:
			statusCode = response.StatusOK
			body = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`
		}

		bodyBytes := []byte(body)
		headers := response.GetDefaultHeaders(len(bodyBytes))
		headers.Set("Content-Type", "text/html")

		if err := w.WriteStatusLine(statusCode); err != nil {
			return
		}
		if err := w.WriteHeaders(headers); err != nil {
			return
		}
		_, _ = w.WriteBody(bodyBytes)
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
