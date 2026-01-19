# httpfromtcp

A tiny HTTP/1.1 server built from raw TCP primitives. It includes a minimal request parser and response writer, plus a small demo app that proxies httpbin and serves a local video file.

## What we implemented

- Basic HTTP/1.1 request parsing (request line, headers, optional body).
- HTTP response writer with status line + headers + body helpers.
- Chunked transfer encoding support, including trailers.
- Demo handler that:
  - Proxies `/httpbin/*` to https://httpbin.org with chunked encoding and trailers.
  - Serves `/video` from `assets/vim.mp4` with `Content-Type: video/mp4`.
  - Serves `/yourproblem`, `/myproblem`, and a default success page.

## Quick start

```bash
go run ./cmd/httpserver
```

The server listens on `http://127.0.0.1:42069`.

## Example requests

```bash
# Default page
curl http://127.0.0.1:42069/

# HTTPBin proxy (chunked + trailers)
curl -i http://127.0.0.1:42069/httpbin/html

# Video file
curl -I http://127.0.0.1:42069/video
```

## Tests

```bash
go test ./...
```
