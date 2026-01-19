package main

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/glebson1988/httpfromtcp/internal/request"
	"github.com/glebson1988/httpfromtcp/internal/response"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type chunkedBody struct {
	chunks [][]byte
	index  int
}

func (b *chunkedBody) Read(p []byte) (int, error) {
	if b.index >= len(b.chunks) {
		return 0, io.EOF
	}
	n := copy(p, b.chunks[b.index])
	b.index++
	return n, nil
}

func (b *chunkedBody) Close() error {
	return nil
}

func TestHTTPBinProxyChunked(t *testing.T) {
	t.Run("proxy writes chunked response with headers", func(t *testing.T) {
		origTransport := http.DefaultTransport
		t.Cleanup(func() {
			http.DefaultTransport = origTransport
		})

		chunks := [][]byte{
			[]byte("hello"),
			[]byte(" "),
			[]byte("world"),
		}
		http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Scheme != "https" {
				t.Fatalf("unexpected scheme: %q", r.URL.Scheme)
			}
			if r.URL.Host != "httpbin.org" {
				t.Fatalf("unexpected host: %q", r.URL.Host)
			}
			if r.URL.Path != "/test" {
				t.Fatalf("unexpected path: %q", r.URL.Path)
			}
			return &http.Response{
				StatusCode:    http.StatusOK,
				Status:        "200 OK",
				Header:        http.Header{"Content-Type": {"text/plain"}, "Content-Length": {"11"}},
				Body:          &chunkedBody{chunks: chunks},
				ContentLength: 11,
			}, nil
		})

		handler := newHandler()
		req := &request.Request{
			RequestLine: request.RequestLine{
				RequestTarget: "/httpbin/test",
			},
		}

		var buf bytes.Buffer
		handler(response.NewWriter(&buf), req)

		statusLine, headers, body := splitResponse(t, buf.Bytes())
		if statusLine != "HTTP/1.1 200 OK" {
			t.Fatalf("unexpected status line: %q", statusLine)
		}
		if _, ok := headers["content-length"]; ok {
			t.Fatalf("unexpected Content-Length header: %q", headers["content-length"])
		}
		if headers["transfer-encoding"] != "chunked" {
			t.Fatalf("unexpected Transfer-Encoding: %q", headers["transfer-encoding"])
		}
		if headers["content-type"] != "text/plain" {
			t.Fatalf("unexpected Content-Type: %q", headers["content-type"])
		}

		chunkSizes, payload := parseChunkedBody(t, body)
		expectedBody := bytes.Join(chunks, nil)
		if !bytes.Equal(payload, expectedBody) {
			t.Fatalf("unexpected body: %q", payload)
		}
		if len(chunkSizes) != len(chunks) {
			t.Fatalf("unexpected chunk count: %d", len(chunkSizes))
		}
		for i, size := range chunkSizes {
			if size != len(chunks[i]) {
				t.Fatalf("unexpected chunk size at %d: %d", i, size)
			}
		}
	})
}

func splitResponse(t *testing.T, data []byte) (string, map[string]string, []byte) {
	t.Helper()

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

	return lines[0], headers, []byte(parts[1])
}

func parseChunkedBody(t *testing.T, data []byte) ([]int, []byte) {
	t.Helper()

	var sizes []int
	var payload []byte

	for {
		lineEnd := bytes.Index(data, []byte("\r\n"))
		if lineEnd == -1 {
			t.Fatalf("missing chunk size line")
		}
		sizeStr := string(data[:lineEnd])
		size64, err := strconv.ParseInt(sizeStr, 16, 64)
		if err != nil {
			t.Fatalf("invalid chunk size: %q", sizeStr)
		}
		size := int(size64)
		data = data[lineEnd+2:]
		if size == 0 {
			if len(data) < 2 {
				t.Fatalf("missing chunked terminator")
			}
			if string(data[:2]) == "\r\n" {
				return sizes, payload
			}
			trailerEnd := bytes.Index(data, []byte("\r\n\r\n"))
			if trailerEnd == -1 {
				t.Fatalf("missing chunked terminator")
			}
			return sizes, payload
		}
		if len(data) < size+2 {
			t.Fatalf("chunk length exceeds remaining data")
		}
		payload = append(payload, data[:size]...)
		data = data[size:]
		if string(data[:2]) != "\r\n" {
			t.Fatalf("chunk missing trailing CRLF")
		}
		data = data[2:]
		sizes = append(sizes, size)
	}
}
