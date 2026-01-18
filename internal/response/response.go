package response

import (
	"fmt"
	"io"

	"github.com/glebson1988/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase := statusReasonPhrase(statusCode)
	var line string
	if reasonPhrase == "" {
		line = fmt.Sprintf("HTTP/1.1 %d\r\n", statusCode)
	} else {
		line = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	}
	_, err := io.WriteString(w, line)
	return err
}

func statusReasonPhrase(statusCode StatusCode) string {
	switch statusCode {
	case StatusOK:
		return "OK"
	case StatusBadRequest:
		return "Bad Request"
	case StatusInternalServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"content-length": fmt.Sprintf("%d", contentLen),
		"connection":     "close",
		"content-type":   "text/plain",
	}
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := io.WriteString(w, headerLine)
		if err != nil {
			return fmt.Errorf("failed to write header %s: %w", key, err)
		}
	}
	_, err := io.WriteString(w, "\r\n")
	if err != nil {
		return fmt.Errorf("failed to write headers terminator: %w", err)
	}
	return nil
}
