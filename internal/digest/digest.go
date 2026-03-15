// Package digest implements HTTP Digest Authentication (RFC 7616 / RFC 2617).
package digest

import (
	"bytes"
	"crypto/md5" //nolint:gosec // MD5 is required by digest auth spec
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// DoWithDigest performs an HTTP request with Digest authentication.
// It first sends a request without credentials, parses the WWW-Authenticate
// challenge from the 401 response, computes the digest, and retries.
func DoWithDigest(client *http.Client, method, url string, body io.Reader, username, password string, headers http.Header) (*http.Response, error) {
	// Read body so we can replay it
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("reading body: %w", err)
		}
	}

	// First request to get the challenge
	req1, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating initial request: %w", err)
	}
	copySelectedHeaders(req1.Header, headers)

	resp1, err := client.Do(req1)
	if err != nil {
		return nil, fmt.Errorf("initial request failed: %w", err)
	}

	if resp1.StatusCode != http.StatusUnauthorized {
		// No authentication needed
		return resp1, nil
	}
	resp1.Body.Close()

	// Parse WWW-Authenticate header
	challenge := resp1.Header.Get("WWW-Authenticate")
	if challenge == "" {
		return nil, fmt.Errorf("no WWW-Authenticate header in 401 response")
	}

	params, err := parseChallenge(challenge)
	if err != nil {
		return nil, fmt.Errorf("parsing challenge: %w", err)
	}

	// Compute digest response
	authHeader := computeAuthorization(params, username, password, method, url)

	// Second request with digest auth
	req2, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating authenticated request: %w", err)
	}
	copySelectedHeaders(req2.Header, headers)
	req2.Header.Set("Authorization", authHeader)

	return client.Do(req2)
}

// challenge represents parsed digest auth challenge parameters.
type challenge struct {
	realm     string
	nonce     string
	opaque    string
	qop       string
	algorithm string
}

// parseChallenge parses a WWW-Authenticate: Digest header value.
func parseChallenge(header string) (*challenge, error) {
	header = strings.TrimPrefix(header, "Digest ")
	header = strings.TrimPrefix(header, "digest ")

	c := &challenge{algorithm: "MD5"}
	parts := splitChallenge(header)

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.Trim(strings.TrimSpace(kv[1]), `"`)

		switch strings.ToLower(key) {
		case "realm":
			c.realm = value
		case "nonce":
			c.nonce = value
		case "opaque":
			c.opaque = value
		case "qop":
			c.qop = value
		case "algorithm":
			c.algorithm = value
		}
	}

	if c.realm == "" || c.nonce == "" {
		return nil, fmt.Errorf("incomplete challenge: missing realm or nonce")
	}

	return c, nil
}

// splitChallenge splits a challenge header value by commas, respecting quoted strings.
func splitChallenge(s string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false

	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
			current.WriteRune(r)
		case r == ',' && !inQuote:
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// computeAuthorization builds the Authorization header value for digest auth.
// Supports MD5, MD5-sess, SHA-256, and SHA-256-sess algorithms (RFC 7616).
func computeAuthorization(c *challenge, username, password, method, uri string) string {
	hashFn := newHashFunc(c.algorithm)

	// Extract just the path from the URI for the digest calculation
	digestURI := extractPath(uri)

	cnonce := generateCNonce()
	nc := "00000001"

	// Compute HA1
	ha1 := digestHash(hashFn, fmt.Sprintf("%s:%s:%s", username, c.realm, password))

	// For -sess variants, HA1 = H(H(username:realm:password):nonce:cnonce)
	algoUpper := strings.ToUpper(c.algorithm)
	if algoUpper == "MD5-SESS" || algoUpper == "SHA-256-SESS" {
		ha1 = digestHash(hashFn, fmt.Sprintf("%s:%s:%s", ha1, c.nonce, cnonce))
	}

	// Compute HA2
	ha2 := digestHash(hashFn, fmt.Sprintf("%s:%s", method, digestURI))

	// Compute response
	var response string
	if c.qop == "auth" || c.qop == "auth-int" {
		response = digestHash(hashFn, fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1, c.nonce, nc, cnonce, c.qop, ha2))
	} else {
		response = digestHash(hashFn, fmt.Sprintf("%s:%s:%s", ha1, c.nonce, ha2))
	}

	header := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
		username, c.realm, c.nonce, digestURI, response)

	if c.qop != "" {
		header += fmt.Sprintf(`, qop=%s, nc=%s, cnonce="%s"`, c.qop, nc, cnonce)
	}
	if c.opaque != "" {
		header += fmt.Sprintf(`, opaque="%s"`, c.opaque)
	}
	if c.algorithm != "" {
		header += fmt.Sprintf(`, algorithm=%s`, c.algorithm)
	}

	return header
}

// extractPath extracts the path (and query) portion from a URL string.
func extractPath(rawURL string) string {
	// Find the path after the host
	idx := strings.Index(rawURL, "://")
	if idx == -1 {
		return rawURL
	}
	rest := rawURL[idx+3:]
	slashIdx := strings.Index(rest, "/")
	if slashIdx == -1 {
		return "/"
	}
	return rest[slashIdx:]
}

// newHashFunc returns a hash.Hash constructor for the given digest algorithm.
// Defaults to MD5 for unknown algorithms per RFC 2617 backwards compatibility.
func newHashFunc(algorithm string) func() hash.Hash {
	switch strings.ToUpper(algorithm) {
	case "SHA-256", "SHA-256-SESS":
		return sha256.New
	default:
		// MD5, MD5-sess, or unrecognized — default to MD5
		return md5.New //nolint:gosec
	}
}

// digestHash computes the hex-encoded hash of s using the provided hash constructor.
func digestHash(newHash func() hash.Hash, s string) string {
	h := newHash()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func generateCNonce() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	return fmt.Sprintf("%08x", r.Uint32())
}

func copySelectedHeaders(dst, src http.Header) {
	skipHeaders := map[string]bool{
		"Authorization":       true,
		"Host":                true,
		"Transfer-Encoding":   true,
		"Connection":          true,
		"Keep-Alive":          true,
		"Proxy-Authorization": true,
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
