package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	state       parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState int

const (
	stateInitialized parserState = iota
	stateDone
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	req := &Request{state: stateInitialized}

	for req.state != stateDone {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read from reader: %w", err)
		}
		if n > 0 {
			readToIndex += n
			consumed, parseErr := req.parse(buf[:readToIndex])
			if parseErr != nil {
				return nil, parseErr
			}
			if consumed > 0 {
				copy(buf, buf[consumed:readToIndex])
				readToIndex -= consumed
			}
		}

		if err == io.EOF {
			req.state = stateDone
			break
		}
	}

	if req.RequestLine.Method == "" {
		return nil, fmt.Errorf("failed to parse request line")
	}
	return req, nil
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	lineEnd := bytes.Index(data, []byte("\r\n"))
	if lineEnd == -1 {
		return RequestLine{}, 0, nil
	}
	line := string(data[:lineEnd])
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, fmt.Errorf("invalid request line: %s", line)
	}
	method, target, version := parts[0], parts[1], parts[2]
	if method == "" {
		return RequestLine{}, 0, fmt.Errorf("invalid HTTP method: %s", method)
	}
	for _, r := range method {
		if r < 'A' || r > 'Z' {
			return RequestLine{}, 0, fmt.Errorf("invalid HTTP method: %s", method)
		}
	}
	if version != "HTTP/1.1" {
		return RequestLine{}, 0, fmt.Errorf("invalid HTTP version: %s", version)
	}
	httpVersion := strings.TrimPrefix(version, "HTTP/")
	return RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   httpVersion,
	}, lineEnd + len("\r\n"), nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case stateInitialized:
		reqLine, consumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil
		}
		r.RequestLine = reqLine
		r.state = stateDone
		return consumed, nil
	case stateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}
