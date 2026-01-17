package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestFromReader(t *testing.T) {
	t.Run("Good request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Good request line with path", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 1,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Good POST request with path", func(t *testing.T) {
		data := "POST /brew HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
		reader := &chunkReader{
			data:            data,
			numBytesPerRead: len(data),
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "POST", r.RequestLine.Method)
		assert.Equal(t, "/brew", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	})

	t.Run("Invalid number of parts in request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 2,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Invalid method request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "get / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 2,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
		reader = &chunkReader{
			data:            "123 / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 2,
		}
		_, err = RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Invalid version in request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/2.0\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 2,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
		reader = &chunkReader{
			data:            "GET / HTTP/1.0\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 2,
		}
		_, err = RequestFromReader(reader)
		require.Error(t, err)
		reader = &chunkReader{
			data:            "GET / HTTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 2,
		}
		_, err = RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Missing request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 2,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Extra spaces in request line", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET  /  HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 2,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Empty request", func(t *testing.T) {
		reader := &chunkReader{
			data:            "",
			numBytesPerRead: 2,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})
}

func TestRequestParseHeaders(t *testing.T) {
	t.Run("Standard Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
		assert.Equal(t, "*/*", r.Headers["accept"])
	})

	t.Run("Empty Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\n\r\n",
			numBytesPerRead: 2,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Empty(t, r.Headers)
	})

	t.Run("Malformed Header", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			numBytesPerRead: 3,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Duplicate Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nSet-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n",
			numBytesPerRead: 4,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", r.Headers["set-person"])
	})

	t.Run("Case Insensitive Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHoST: localhost:42069\r\nUser-AGENT: curl/7.81.0\r\n\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	})

	t.Run("Missing End of Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\n",
			numBytesPerRead: 3,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Standard Body", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: 5,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "hello world!\n", string(r.Body))
	})

	t.Run("Body shorter than reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"Partial content\n",
			numBytesPerRead: 5,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	t.Run("Empty Body, 0 reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 0\r\n" +
				"\r\n",
			numBytesPerRead: 5,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))
	})

	t.Run("Empty Body, no reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 5,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))
	})

	t.Run("No Content-Length but Body Exists", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n" +
				"Body without length",
			numBytesPerRead: 5,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))
	})
}
