# gin-balancer

A lightweight reverse proxy built with [Gin](https://github.com/gin-gonic/gin)
that forwards screenshot requests to a pool of backend endpoints while
randomising the User-Agent header. The service accepts a JSON payload containing
the target URL and responds with the screenshot provider's metadata.

## Quick start

```bash
GOTOOLCHAIN=go1.22.7 go run ./main.go
```

Or use the provided make targets, which default to the module's preferred
toolchain:

```bash
make run
```

### Docker

Build and run the container locally:

```bash
make docker-run PORT=9090
```

Or execute the commands manually:

```bash
docker build -t gin-balancer:latest .
docker run --rm -p 8080:8080 gin-balancer:latest
```

The server listens on `http://localhost:8080` by default and exposes a single
endpoint. To bind a different address set the `PORT` environment variable or
run `GOTOOLCHAIN=go1.22.7 go run ./main.go --port :9090` / `make run PORT=9090`.

- `POST /screenshot` – body: `{ "url": "https://example.com" }`

Example request:

```bash
curl --request POST \
	--url http://localhost:8080/screenshot \
	--header 'Content-Type: application/json' \
	--data '{"url": "https://example.com"}'
```

On success the response mirrors the data returned by the screenshot provider
and includes the endpoint that handled the call. If every backend fails the
service responds with `500` and the last error message.

## Development

1. Install Go 1.22.7 (or newer) or rely on the [toolchain directive](https://go.dev/doc/toolchain) by prefixing commands with `GOTOOLCHAIN=go1.22.7`.
2. Download the dependencies:
	```bash
	GOTOOLCHAIN=go1.22.7 go mod tidy
	```
3. Run the test suite:
	```bash
	make test
	```
4. Run the server with auto-reload (optional) using a tool like
	[`air`](https://github.com/cosmtrek/air) (ensure it respects the `GOTOOLCHAIN`
	setting) or restart manually after code changes.

## Notes

- Each incoming request is forwarded to one backend at a time until one
	succeeds or all fail.
- User-Agent headers are chosen randomly from a fixed list to help distribute
	calls across different browser signatures.