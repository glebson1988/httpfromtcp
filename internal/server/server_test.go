package server

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestServeWritesResponse(t *testing.T) {
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

	body := "Hello World!\n"
	expected := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		len(body),
		body,
	)
	if string(data) != expected {
		t.Fatalf("unexpected response:\n%s", string(data))
	}
}

func TestServerCloseStopsAccepting(t *testing.T) {
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
}
