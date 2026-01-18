package response

import (
	"fmt"
	"io"

	"github.com/glebson1988/httpfromtcp/internal/headers"
)

type Headers = headers.Headers

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateDone
)

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  writerStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStatusLine {
		return fmt.Errorf("status line must be written first")
	}
	if err := WriteStatusLine(w.writer, statusCode); err != nil {
		return err
	}
	w.state = writerStateHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers Headers) error {
	if w.state != writerStateHeaders {
		return fmt.Errorf("headers must be written after status line")
	}
	if err := WriteHeaders(w.writer, headers); err != nil {
		return err
	}
	w.state = writerStateBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("body must be written after status line and headers")
	}
	n, err := w.writer.Write(p)
	if err != nil {
		return n, err
	}
	w.state = writerStateDone
	return n, nil
}

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

func GetDefaultHeaders(contentLen int) Headers {
	return Headers{
		"content-length": fmt.Sprintf("%d", contentLen),
		"connection":     "close",
		"content-type":   "text/plain",
	}
}

func WriteHeaders(w io.Writer, headers Headers) error {
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
