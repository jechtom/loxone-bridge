# LoxoneBridge

A lightweight, stateless HTTP proxy service that extends the networking capabilities of Loxone Miniserver. It bridges gaps where Loxone's built-in HTTP client falls short.

Unlike Home Assistant or Node-RED, LoxoneBridge has no database, no config files, and no UI — just a single binary you deploy and forget. All routing logic lives in the URL itself. Fewer moving parts means fewer things that can break at 2 AM when your alarm system needs to work.

![Loxone Bridge Logo](loxone-bridge-logo-small.png)

## Features

1. **Digest Authentication Translation** — Converts Loxone-compatible Basic Authentication into Digest Authentication for third-party devices (e.g., Shelly, Dahua) that require it.
2. **JSON Flattening** — Accepts complex JSON responses and converts them into a flat `key=value` format easily parseable by Loxone.
3. **HTTP-to-UDP Translation** — Receives HTTP requests and forwards the data as raw UDP datagrams to any target.
4. **HTTPS with Certificate Ignoring** — Proxies to HTTPS endpoints while intentionally ignoring invalid or self-signed certificates.

## Philosophy

The application is designed for maximum simplicity. It is **stateless** and requires **no configuration files**. All configuration is declared directly in the request URL — Loxone sends its requests with routing instructions embedded in the path.

## Usage Examples

- [`docs/dahua-vto-door-unlock.md`](./docs/dahua-vto-door-unlock.md) — Dahua VTO unlock via digest authentication
- [`docs/dahua-event-trigger.md`](./docs/dahua-event-trigger.md) — Dahua HTTP event callbacks forwarded to Loxone UDP
- [`docs/shelly-communication.md`](./docs/shelly-communication.md) — Shelly digest authentication with flattened JSON responses

## Quick Start

### Docker

```bash
docker run -d -p 8080:8080 ghcr.io/jechtom/loxone-bridge:latest
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

This is particularly useful for translating HTTP webhooks into UDP messages that Loxone Miniserver can natively consume via its UDP Virtual Input. For example, Dahua cameras can send HTTP alarm notifications (motion detection, tampering, etc.) to LoxoneBridge, which then forwards them as UDP datagrams directly to Loxone — no additional middleware needed.

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

For more complete Loxone-oriented setup examples, see the dedicated guides in the [`docs/`](./docs/) folder.

---

## URL Format

```
http://loxone-bridge/{modifiers*}/{protocol}/{address}/{path}
```

## Modifiers

Zero or more optional URL segments placed before the protocol. Modifiers alter how LoxoneBridge processes the request or the response. Multiple modifiers can be chained together in any order.

| Segment        | Description                                                   |
|----------------|---------------------------------------------------------------|
| `/digest`      | Converts Basic Auth credentials from the incoming request into Digest Auth for the upstream target |
| `/flatten-json`| Converts the upstream JSON response into a flat `key=value` text format (one line per value) |

## Protocols

The protocol segment determines how LoxoneBridge forwards the request to the target device. It defines the transport scheme and, in the case of UDP, changes the request semantics entirely.

| Segment              | Description                                              |
|----------------------|----------------------------------------------------------|
| `/http`              | Plain HTTP (default port 80)                             |
| `/https`             | HTTPS with standard certificate validation (default port 443) |
| `/https-ignore-cert` | HTTPS that skips TLS certificate verification — useful for self-signed or expired certificates |
| `/udp`               | Sends data as a raw UDP datagram — port is **required** in the address segment |

## Address

The target host as an IP address or DNS hostname. A port can be appended with `:port` (e.g., `192.168.1.10:8080`). The port is **required** for UDP. For HTTP and HTTPS the standard default ports (`80` / `443`) are used when omitted.

## Path

Everything after the address segment is forwarded as-is to the target, including query parameters. For example, in `/http/10.0.0.5/cgi-bin/control?action=open&channel=1`, the path `/cgi-bin/control?action=open&channel=1` is sent to the upstream server unchanged. For the UDP protocol, the path content is used as the UDP payload (unless a request body is provided, which takes precedence).

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