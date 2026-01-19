package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/glebson1988/httpfromtcp/internal/request"
	"github.com/glebson1988/httpfromtcp/internal/response"
	"github.com/glebson1988/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	handler := newHandler()
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

func newHandler() func(w *response.Writer, req *request.Request) {
	return func(w *response.Writer, req *request.Request) {
		if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			targetPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
			resp, err := http.Get("https://httpbin.org" + targetPath)
			if err != nil {
				body := []byte("failed to reach upstream")
				headers := response.GetDefaultHeaders(len(body))
				if err := w.WriteStatusLine(response.StatusInternalServerError); err != nil {
					return
				}
				if err := w.WriteHeaders(headers); err != nil {
					return
				}
				_, _ = w.WriteBody(body)
				return
			}
			defer resp.Body.Close()

			headers := make(response.Headers)
			for key, values := range resp.Header {
				headers.Set(key, strings.Join(values, ", "))
			}
			delete(headers, "content-length")
			headers.Set("Transfer-Encoding", "chunked")

			if err := w.WriteStatusLine(response.StatusCode(resp.StatusCode)); err != nil {
				return
			}
			if err := w.WriteHeaders(headers); err != nil {
				return
			}

			buf := make([]byte, 1024)
			for {
				n, err := resp.Body.Read(buf)
				log.Println(n)
				if n > 0 {
					if _, err := w.WriteChunkedBody(buf[:n]); err != nil {
						return
					}
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					return
				}
			}
			_, _ = w.WriteChunkedBodyDone()
			return
		}

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
}
