// Package proxy provides HTTP/HTTPS proxy functionality with optional
// digest authentication and TLS certificate ignoring.
package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/loxone-bridge/internal/digest"
)

// Do performs an HTTP request to the target URL, optionally using digest
// authentication and/or ignoring TLS certificate errors.
//
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - targetURL: full URL to proxy to
//   - body: request body (may be nil)
//   - username, password: credentials for digest auth (empty strings to skip)
//   - useDigest: whether to use digest authentication
//   - ignoreCert: whether to ignore TLS certificate errors
//   - headers: additional headers to forward
func Do(method, targetURL string, body io.Reader, username, password string, useDigest, ignoreCert bool, headers http.Header) (*http.Response, error) {
	transport := &http.Transport{}
	if ignoreCert {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // intentional for self-signed certs
	}
	client := &http.Client{Transport: transport}

	if useDigest && username != "" {
		return digest.DoWithDigest(client, method, targetURL, body, username, password, headers)
	}

	req, err := http.NewRequest(method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	copyHeaders(req.Header, headers)

	return client.Do(req)
}

// copyHeaders copies selected headers from src to dst, skipping hop-by-hop
// headers and authorization (which is handled separately).
func copyHeaders(dst, src http.Header) {
	skipHeaders := map[string]bool{
		"Authorization":       true,
		"Host":                true,
		"Transfer-Encoding":   true,
		"Connection":          true,
		"Keep-Alive":          true,
		"Proxy-Authorization": true,
		"Te":                  true,
		"Trailers":            true,
		"Upgrade":             true,
	}

	for key, values := range src {
		if skipHeaders[key] {
			continue
		}
		for _, v := range values {
			dst.Add(key, v)
		}
	}
}
