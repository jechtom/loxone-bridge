# LoxoneBridge

A lightweight, stateless HTTP proxy service that extends the networking capabilities of Loxone Miniserver. It bridges gaps where Loxone's built-in HTTP client falls short.

## Features

1. **Digest Authentication Translation** — Converts Loxone-compatible Basic Authentication into Digest Authentication for third-party devices (e.g., Shelly, Dahua) that require it.
2. **JSON Flattening** — Accepts complex JSON responses and converts them into a flat `key=value` format easily parseable by Loxone.
3. **HTTP-to-UDP Translation** — Receives HTTP requests and forwards the data as raw UDP datagrams to any target.
4. **HTTPS with Certificate Ignoring** — Proxies to HTTPS endpoints while intentionally ignoring invalid or self-signed certificates.

## Philosophy

The application is designed for maximum simplicity. It is **stateless** and requires **no configuration files**. All configuration is declared directly in the request URL — Loxone sends its requests with routing instructions embedded in the path.

## Quick Start

### Docker

```bash
docker run -d -p 8080:8080 ghcr.io/<owner>/loxone-bridge:latest
```

### Build from Source

```bash
go build -o loxone-bridge ./cmd/loxone-bridge
./loxone-bridge
```

The service listens on port `8080` by default. Set the `PORT` environment variable to change it.

### Health Check

```
GET /healthz
```

## Usage Examples

### Basic Auth → Digest Auth (HTTP)

```
GET /digest/http/192.168.1.10/cgi-bin/accessControl.cgi?action=openDoor&channel=1
```

Takes the Basic Authentication credentials sent by Loxone, establishes a Digest Authentication handshake with the target, and proxies the request to `http://192.168.1.10/cgi-bin/accessControl.cgi?action=openDoor&channel=1`.

In Loxone, configure the URL as:
```
http://admin:PASSWORD@loxone-bridge:8080/digest/http/10.0.0.5/cgi-bin/accessControl.cgi?action=openDoor&channel=1
```

### HTTPS with Ignored Certificate Errors

```
GET /https-ignore-cert/192.168.1.10/api/status
```

Proxies to `https://192.168.1.10/api/status` while ignoring invalid or self-signed TLS certificates.

### Send UDP Packets via HTTP

```
GET /udp/192.168.1.10:444/data-to-send
```

Sends the path content (or request body if present) as a raw UDP datagram to `192.168.1.10:444`.

### Flatten JSON Response

```
GET /flatten-json/http/192.168.1.10/api/data
```

Fetches `http://192.168.1.10/api/data` and converts the JSON response:

```json
{
  "data": {
      "volume": 124,
      "error": false
  },
  "name": "device-1",
  "versions": ["a", "b", "c"]
}
```

Into a flat text format:

```
data.error=false
data.volume=124
name=device-1
versions[0]=a
versions[1]=b
versions[2]=c
```

### Combining Modifiers

Modifiers can be chained. For example, Digest Auth + JSON Flattening:

```
GET /digest/flatten-json/https/10.0.0.1/api/sensors
```

---

## URL Format

```
http://loxone-bridge/{modifiers}/{protocol}/{address}/{path-and-query}
       |-----------| |---------| |--------| |--------|
       |             |           |          |
       |             |           |          +-- downstream path forwarded to target
       |             |           |
       |             |           +-- target address (IP or hostname, optional port)
       |             |
       |             +-- zero or more modifiers
       |
       +-- LoxoneBridge address
```

## Modifiers

| Segment        | Description                                                   |
|----------------|---------------------------------------------------------------|
| `/digest`      | Converts Basic Auth to Digest Auth for the upstream request   |
| `/flatten-json`| Converts JSON response to flat `key=value` text format        |

## Protocols

| Segment              | Description                                              |
|----------------------|----------------------------------------------------------|
| `/http`              | Plain HTTP                                               |
| `/https`             | HTTPS with certificate validation                        |
| `/https-ignore-cert` | HTTPS ignoring certificate errors (self-signed, expired) |
| `/udp`               | Send data as UDP datagram (port required in address)     |

## Address

Target address as IP or hostname. Optionally include a port with `:port`. Port is **required** for UDP; HTTP defaults to `80`, HTTPS to `443`.

## Development

### Prerequisites

- Go 1.23+

### Run Tests

```bash
go test -v -race ./...
```

### Build

```bash
go build -o loxone-bridge ./cmd/loxone-bridge
```

### Docker Build

```bash
docker build -t loxone-bridge .
```

## CI/CD

- **CI** (`ci.yml`): Runs tests and builds on every push to `main` and on pull requests.
- **Release** (`release.yml`): On tag push (`v*`), runs tests, builds a multi-arch Docker image (amd64 + arm64), and pushes to GitHub Container Registry (GHCR).

### Creating a Release

```bash
git tag v1.0.0
git push origin v1.0.0
```

## License

MIT