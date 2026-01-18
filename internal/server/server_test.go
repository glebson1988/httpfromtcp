package server

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glebson1988/httpfromtcp/internal/request"
	"github.com/glebson1988/httpfromtcp/internal/response"
)

func TestServer(t *testing.T) {
	handler := func(w *response.Writer, req *request.Request) {
		var statusCode response.StatusCode
		var body string
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			statusCode = response.StatusBadRequest
			body = "Your problem is not my problem\n"
		case "/myproblem":
			statusCode = response.StatusInternalServerError
			body = "Woopsie, my bad\n"
		default:
			statusCode = response.StatusOK
			body = "All good, frfr\n"
		}

		bodyBytes := []byte(body)
		headers := response.GetDefaultHeaders(len(bodyBytes))
		if err := w.WriteStatusLine(statusCode); err != nil {
			return
		}
		if err := w.WriteHeaders(headers); err != nil {
			return
		}
		_, _ = w.WriteBody(bodyBytes)
	}

	t.Run("Serve writes success response", func(t *testing.T) {
		srv, err := Serve(0, handler)
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = srv.Close()
		})

		statusLine, headers, body := sendRequest(t, srv.listener.Addr().String(), "/")
		if statusLine != "HTTP/1.1 200 OK" {
			t.Fatalf("unexpected status line: %q", statusLine)
		}
		if body != "All good, frfr\n" {
			t.Fatalf("unexpected body: %q", body)
		}

		if headers["content-length"] != strconv.Itoa(len(body)) {
			t.Fatalf("unexpected Content-Length: %q", headers["content-length"])
		}
		if headers["connection"] != "close" {
			t.Fatalf("unexpected Connection: %q", headers["connection"])
		}
		if headers["content-type"] != "text/plain" {
			t.Fatalf("unexpected Content-Type: %q", headers["content-type"])
		}
	})

	t.Run("Serve writes handler error responses", func(t *testing.T) {
		srv, err := Serve(0, handler)
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = srv.Close()
		})

		statusLine, headers, body := sendRequest(t, srv.listener.Addr().String(), "/yourproblem")
		if statusLine != "HTTP/1.1 400 Bad Request" {
			t.Fatalf("unexpected status line: %q", statusLine)
		}
		if body != "Your problem is not my problem\n" {
			t.Fatalf("unexpected body: %q", body)
		}
		if headers["content-length"] != strconv.Itoa(len(body)) {
			t.Fatalf("unexpected Content-Length: %q", headers["content-length"])
		}

		statusLine, headers, body = sendRequest(t, srv.listener.Addr().String(), "/myproblem")
		if statusLine != "HTTP/1.1 500 Internal Server Error" {
			t.Fatalf("unexpected status line: %q", statusLine)
		}
		if body != "Woopsie, my bad\n" {
			t.Fatalf("unexpected body: %q", body)
		}
		if headers["content-length"] != strconv.Itoa(len(body)) {
			t.Fatalf("unexpected Content-Length: %q", headers["content-length"])
		}
	})

	t.Run("Close stops accepting", func(t *testing.T) {
		srv, err := Serve(0, handler)
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}

		addr := srv.listener.Addr().String()
		if err := srv.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
		if err := srv.Close(); err != nil {
			t.Fatalf("Close returned error on second call: %v", err)
		}

		time.Sleep(10 * time.Millisecond)
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			t.Fatalf("expected dial to fail after Close")
		}
	})
}

func sendRequest(t *testing.T, addr, path string) (string, map[string]string, string) {
	t.Helper()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial returned error: %v", err)
	}
	defer conn.Close()

	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\n\r\n", path, addr)
	if _, err := io.WriteString(conn, req); err != nil {
		t.Fatalf("WriteString returned error: %v", err)
	}

	data, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	parts := strings.SplitN(string(data), "\r\n\r\n", 2)
	if len(parts) != 2 {
		t.Fatalf("unexpected response: %q", string(data))
	}

	lines := strings.Split(parts[0], "\r\n")
	if len(lines) == 0 {
		t.Fatalf("missing status line: %q", string(data))
	}

	headers := make(map[string]string)
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		kv := strings.SplitN(line, ": ", 2)
		if len(kv) != 2 {
			t.Fatalf("invalid header line: %q", line)
		}
		headers[strings.ToLower(kv[0])] = kv[1]
	}

	return lines[0], headers, parts[1]
}
