package response

import (
	"bytes"
	"strings"
	"testing"
)

func TestChunkedBodyWrites(t *testing.T) {
	t.Run("writes chunks and terminator", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewWriter(&buf)
		if err := writer.WriteStatusLine(StatusOK); err != nil {
			t.Fatalf("WriteStatusLine error: %v", err)
		}
		if err := writer.WriteHeaders(Headers{"transfer-encoding": "chunked"}); err != nil {
			t.Fatalf("WriteHeaders error: %v", err)
		}

		if n, err := writer.WriteChunkedBody([]byte("hello")); err != nil {
			t.Fatalf("WriteChunkedBody error: %v", err)
		} else if n != 5 {
			t.Fatalf("unexpected chunk write size: %d", n)
		}
		if n, err := writer.WriteChunkedBody([]byte("world")); err != nil {
			t.Fatalf("WriteChunkedBody error: %v", err)
		} else if n != 5 {
			t.Fatalf("unexpected chunk write size: %d", n)
		}
		if n, err := writer.WriteChunkedBodyDone(); err != nil {
			t.Fatalf("WriteChunkedBodyDone error: %v", err)
		} else if n == 0 {
			t.Fatalf("expected terminator bytes written")
		}

		parts := strings.SplitN(buf.String(), "\r\n\r\n", 2)
		if len(parts) != 2 {
			t.Fatalf("missing header separator")
		}
		body := parts[1]
		if body != "5\r\nhello\r\n5\r\nworld\r\n0\r\n\r\n" {
			t.Fatalf("unexpected chunked body: %q", body)
		}
	})

	t.Run("writes trailers after chunks", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewWriter(&buf)
		if err := writer.WriteStatusLine(StatusOK); err != nil {
			t.Fatalf("WriteStatusLine error: %v", err)
		}
		if err := writer.WriteHeaders(Headers{"transfer-encoding": "chunked"}); err != nil {
			t.Fatalf("WriteHeaders error: %v", err)
		}
		if _, err := writer.WriteChunkedBody([]byte("hello")); err != nil {
			t.Fatalf("WriteChunkedBody error: %v", err)
		}

		trailers := Headers{
			"x-content-sha256": "deadbeef",
			"x-content-length": "5",
		}
		if err := writer.WriteTrailers(trailers); err != nil {
			t.Fatalf("WriteTrailers error: %v", err)
		}

		parts := strings.SplitN(buf.String(), "\r\n\r\n", 2)
		if len(parts) != 2 {
			t.Fatalf("missing header separator")
		}
		body := parts[1]
		if !strings.Contains(body, "0\r\nx-content-length: 5\r\nx-content-sha256: deadbeef\r\n\r\n") &&
			!strings.Contains(body, "0\r\nx-content-sha256: deadbeef\r\nx-content-length: 5\r\n\r\n") {
			t.Fatalf("unexpected trailer block: %q", body)
		}
	})

	t.Run("requires body state", func(t *testing.T) {
		var buf bytes.Buffer
		writer := NewWriter(&buf)
		if _, err := writer.WriteChunkedBody([]byte("nope")); err == nil {
			t.Fatalf("expected error before headers")
		}
		if _, err := writer.WriteChunkedBodyDone(); err == nil {
			t.Fatalf("expected error before headers")
		}
	})
}
