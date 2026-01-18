package server

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	t.Run("Serve writes response", func(t *testing.T) {
		srv, err := Serve(0)
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = srv.Close()
		})

		addr := srv.listener.Addr().String()
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Dial returned error: %v", err)
		}
		defer conn.Close()

		data, err := io.ReadAll(conn)
		if err != nil {
			t.Fatalf("ReadAll returned error: %v", err)
		}

		lines := strings.Split(string(data), "\r\n")
		if len(lines) < 2 {
			t.Fatalf("unexpected response: %q", string(data))
		}
		if lines[0] != "HTTP/1.1 200 OK" {
			t.Fatalf("unexpected status line: %q", lines[0])
		}

		headers := make(map[string]string)
		for _, line := range lines[1:] {
			if line == "" {
				break
			}
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) != 2 {
				t.Fatalf("invalid header line: %q", line)
			}
			headers[strings.ToLower(parts[0])] = parts[1]
		}

		if headers["content-length"] != "0" {
			t.Fatalf("unexpected Content-Length: %q", headers["content-length"])
		}
		if headers["connection"] != "close" {
			t.Fatalf("unexpected Connection: %q", headers["connection"])
		}
		if headers["content-type"] != "text/plain" {
			t.Fatalf("unexpected Content-Type: %q", headers["content-type"])
		}
	})

	t.Run("Close stops accepting", func(t *testing.T) {
		srv, err := Serve(0)
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
