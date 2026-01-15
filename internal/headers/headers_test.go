package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		h := Headers{}
		n, done, err := h.Parse([]byte("Host: localhost:42069\r\n"))

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, len("Host: localhost:42069\r\n"), n)
		assert.Equal(t, "localhost:42069", h["Host"])
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		h := Headers{}
		n, done, err := h.Parse([]byte("Host:\t localhost:42069 \r\n"))

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, len("Host:\t localhost:42069 \r\n"), n)
		assert.Equal(t, "localhost:42069", h["Host"])
	})

	t.Run("No CRLF yet", func(t *testing.T) {
		h := Headers{}
		n, done, err := h.Parse([]byte("Host: localhost"))

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, 0, n)
		assert.Empty(t, h)
	})

	t.Run("Valid header with leading whitespace", func(t *testing.T) {
		h := Headers{}
		n, done, err := h.Parse([]byte("   Host: localhost:42069\r\n"))

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, len("   Host: localhost:42069\r\n"), n)
		assert.Equal(t, "localhost:42069", h["Host"])
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		h := Headers{"Connection": "keep-alive"}

		n, done, err := h.Parse([]byte("Host: localhost:42069\r\n"))
		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, len("Host: localhost:42069\r\n"), n)
		assert.Equal(t, "localhost:42069", h["Host"])
		assert.Equal(t, "keep-alive", h["Connection"])

		n, done, err = h.Parse([]byte("User-Agent: test\r\n"))
		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, len("User-Agent: test\r\n"), n)
		assert.Equal(t, "test", h["User-Agent"])
	})

	t.Run("Valid done", func(t *testing.T) {
		h := Headers{}
		n, done, err := h.Parse([]byte("\r\n"))

		require.NoError(t, err)
		assert.True(t, done)
		assert.Equal(t, len("\r\n"), n)
		assert.Empty(t, h)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		h := Headers{}
		n, done, err := h.Parse([]byte("Host : localhost:42069\r\n"))

		require.Error(t, err)
		assert.False(t, done)
		assert.Equal(t, 0, n)
		assert.Empty(t, h)
	})
}
