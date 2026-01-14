package request

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoodRequestLine(t *testing.T) {
	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestGoodRequestLineWithPath(t *testing.T) {
	r, err := RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestGoodPOSTRequestWithPath(t *testing.T) {
	r, err := RequestFromReader(strings.NewReader("POST /brew HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/brew", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestInvalidNumberOfPartsInRequestLine(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
}

func TestInvalidMethodRequestLine(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("get / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
	_, err = RequestFromReader(strings.NewReader("123 / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
}

func TestInvalidVersionInRequestLine(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET / HTTP/2.0\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
	_, err = RequestFromReader(strings.NewReader("GET / HTTP/1.0\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
	_, err = RequestFromReader(strings.NewReader("GET / HTTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
}

func TestMissingRequestLine(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
}

func TestExtraSpacesInRequestLine(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET  /  HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	require.Error(t, err)
}

func TestEmptyRequest(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader(""))
	require.Error(t, err)
}
