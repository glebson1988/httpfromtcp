package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}
	reqStr := string(data)
	reqLineEnd := strings.Index(reqStr, "\r\n")
	if reqLineEnd == -1 {
		return nil, fmt.Errorf("invalid request: missing CRLF")
	}
	reqLineStr := reqStr[:reqLineEnd]
	reqLine, err := parseRequestLine(reqLineStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request line: %w", err)
	}
	return &Request{
		RequestLine: *reqLine,
	}, nil
}

func parseRequestLine(line string) (*RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", line)
	}
	method, target, version := parts[0], parts[1], parts[2]
	if method == "" {
		return nil, fmt.Errorf("invalid HTTP method: %s", method)
	}
	for _, r := range method {
		if r < 'A' || r > 'Z' {
			return nil, fmt.Errorf("invalid HTTP method: %s", method)
		}
	}
	if version != "HTTP/1.1" {
		return nil, fmt.Errorf("invalid HTTP version: %s", version)
	}
	httpVersion := strings.TrimPrefix(version, "HTTP/")
	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   httpVersion,
	}, nil
}
