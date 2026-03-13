# Agent Guidelines

## Project Overview

LoxoneBridge is a stateless HTTP proxy service written in Go. It extends Loxone Miniserver's networking capabilities by providing digest auth translation, JSON flattening, HTTPS certificate ignoring, and HTTP-to-UDP forwarding.

## Architecture

```
cmd/loxone-bridge/main.go     -- entry point, HTTP server setup
internal/
  handler/handler.go           -- main HTTP handler, request routing
  urlparser/urlparser.go       -- URL parsing (modifiers, protocol, address, path)
  proxy/proxy.go               -- HTTP/HTTPS proxy with optional digest auth
  digest/digest.go             -- HTTP Digest Authentication (RFC 2617)
  flatten/flatten.go           -- JSON to flat key=value conversion
  udpsender/udpsender.go       -- UDP datagram sender
```

## Key Design Decisions

- **Stateless**: No configuration files. All routing is determined by the request URL.
- **Standard library first**: Minimal external dependencies (only `testify` for tests).
- **Internal packages**: All packages are under `internal/` to prevent external imports.

## Development

- Go 1.23+
- Run tests: `go test -v -race ./...`
- Build: `go build -o loxone-bridge ./cmd/loxone-bridge`

## CI/CD

- CI runs on push to `main` and on PRs (tests + build + Docker build validation).
- Release on tag push (`v*`): tests, multi-arch Docker image build, push to GHCR.

## Conventions

- All code, comments, and documentation in English.
- Tests alongside source files (`*_test.go`).
- Use `log.Printf` for structured logging with `REQ:` / `ERROR:` prefixes.
