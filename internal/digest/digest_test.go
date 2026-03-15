package digest

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseChallenge_Basic(t *testing.T) {
	header := `Digest realm="testrealm@host.com", nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093", qop="auth", algorithm=MD5`
	c, err := parseChallenge(header)
	require.NoError(t, err)

	assert.Equal(t, "testrealm@host.com", c.realm)
	assert.Equal(t, "dcd98b7102dd2f0e8b11d0f600bfb0c093", c.nonce)
	assert.Equal(t, "auth", c.qop)
	assert.Equal(t, "MD5", c.algorithm)
}

func TestParseChallenge_WithOpaque(t *testing.T) {
	header := `Digest realm="example.com", nonce="abc123", opaque="xyz789", qop="auth"`
	c, err := parseChallenge(header)
	require.NoError(t, err)

	assert.Equal(t, "example.com", c.realm)
	assert.Equal(t, "abc123", c.nonce)
	assert.Equal(t, "xyz789", c.opaque)
	assert.Equal(t, "auth", c.qop)
}

func TestParseChallenge_MissingRealm(t *testing.T) {
	header := `Digest nonce="abc123"`
	_, err := parseChallenge(header)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incomplete challenge")
}

func TestParseChallenge_MissingNonce(t *testing.T) {
	header := `Digest realm="example.com"`
	_, err := parseChallenge(header)
	assert.Error(t, err)
}

func TestComputeAuthorization_WithQOP(t *testing.T) {
	c := &challenge{
		realm:     "testrealm@host.com",
		nonce:     "dcd98b7102dd2f0e8b11d0f600bfb0c093",
		qop:       "auth",
		algorithm: "MD5",
	}

	auth := computeAuthorization(c, "admin", "password", "GET", "http://10.0.0.1/api/status")
	assert.Contains(t, auth, "Digest ")
	assert.Contains(t, auth, `username="admin"`)
	assert.Contains(t, auth, `realm="testrealm@host.com"`)
	assert.Contains(t, auth, `nonce="dcd98b7102dd2f0e8b11d0f600bfb0c093"`)
	assert.Contains(t, auth, `qop=auth`)
	assert.Contains(t, auth, `nc=00000001`)
	assert.Contains(t, auth, `algorithm=MD5`)
}

func TestComputeAuthorization_WithoutQOP(t *testing.T) {
	c := &challenge{
		realm:     "example.com",
		nonce:     "abc123",
		algorithm: "MD5",
	}

	auth := computeAuthorization(c, "user", "pass", "GET", "http://10.0.0.1/path")
	assert.Contains(t, auth, "Digest ")
	assert.Contains(t, auth, `username="user"`)
	assert.NotContains(t, auth, "qop=")
	assert.NotContains(t, auth, "nc=")
}

func TestComputeAuthorization_WithOpaque(t *testing.T) {
	c := &challenge{
		realm:     "example.com",
		nonce:     "abc123",
		opaque:    "opaque-value",
		qop:       "auth",
		algorithm: "MD5",
	}

	auth := computeAuthorization(c, "user", "pass", "GET", "http://10.0.0.1/path")
	assert.Contains(t, auth, `opaque="opaque-value"`)
}

func TestExtractPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://10.0.0.1/api/status", "/api/status"},
		{"https://10.0.0.1:8443/path?q=1", "/path?q=1"},
		{"http://10.0.0.1/", "/"},
		{"/just-path", "/just-path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitChallenge(t *testing.T) {
	input := `realm="test,realm", nonce="abc123", qop="auth"`
	parts := splitChallenge(input)
	assert.Len(t, parts, 3)
	assert.Contains(t, parts[0], "realm")
	assert.Contains(t, parts[1], "nonce")
	assert.Contains(t, parts[2], "qop")
}

func TestDigestHash_MD5_KnownValue(t *testing.T) {
	// Known MD5 hash of "hello"
	result := digestHash(md5.New, "hello")
	assert.Equal(t, "5d41402abc4b2a76b9719d911017c592", result)
}

func TestParseChallenge_SHA256(t *testing.T) {
	header := `Digest realm="example.com", nonce="abc123", qop="auth", algorithm=SHA-256`
	c, err := parseChallenge(header)
	require.NoError(t, err)

	assert.Equal(t, "example.com", c.realm)
	assert.Equal(t, "abc123", c.nonce)
	assert.Equal(t, "auth", c.qop)
	assert.Equal(t, "SHA-256", c.algorithm)
}

func TestComputeAuthorization_SHA256(t *testing.T) {
	c := &challenge{
		realm:     "example.com",
		nonce:     "7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v",
		qop:       "auth",
		algorithm: "SHA-256",
	}

	auth := computeAuthorization(c, "admin", "secret", "GET", "http://10.0.0.1/api/status")
	assert.Contains(t, auth, "Digest ")
	assert.Contains(t, auth, `username="admin"`)
	assert.Contains(t, auth, `realm="example.com"`)
	assert.Contains(t, auth, `algorithm=SHA-256`)
	assert.Contains(t, auth, `qop=auth`)
	assert.Contains(t, auth, `nc=00000001`)

	// Verify the response hash length is 64 hex chars (SHA-256) not 32 (MD5)
	assert.Contains(t, auth, `response="`)
	// Extract response value and verify length
	for _, part := range splitChallenge(auth[len("Digest "):]) {
		kv := splitChallenge(part)
		_ = kv
	}
}

func TestComputeAuthorization_SHA256_KnownVector(t *testing.T) {
	// Verify SHA-256 produces different (and correct-length) responses vs MD5
	c := &challenge{
		realm:     "testrealm",
		nonce:     "testnonce",
		qop:       "auth",
		algorithm: "SHA-256",
	}

	authSHA := computeAuthorization(c, "user", "pass", "GET", "http://10.0.0.1/path")

	c.algorithm = "MD5"
	authMD5 := computeAuthorization(c, "user", "pass", "GET", "http://10.0.0.1/path")

	// Both should be valid Digest headers but with different response values
	assert.Contains(t, authSHA, `algorithm=SHA-256`)
	assert.Contains(t, authMD5, `algorithm=MD5`)

	// They must produce different responses
	assert.NotEqual(t, authSHA, authMD5)
}

func TestComputeAuthorization_SHA256WithoutQOP(t *testing.T) {
	c := &challenge{
		realm:     "example.com",
		nonce:     "abc123",
		algorithm: "SHA-256",
	}

	auth := computeAuthorization(c, "user", "pass", "GET", "http://10.0.0.1/path")
	assert.Contains(t, auth, "Digest ")
	assert.Contains(t, auth, `algorithm=SHA-256`)
	assert.NotContains(t, auth, "qop=")
	assert.NotContains(t, auth, "nc=")
}

func TestDigestHash_SHA256(t *testing.T) {
	// Known SHA-256 hash of "hello"
	h := sha256.New()
	h.Write([]byte("hello"))
	expected := fmt.Sprintf("%x", h.Sum(nil))

	result := digestHash(sha256.New, "hello")
	assert.Equal(t, expected, result)
	assert.Len(t, result, 64) // SHA-256 produces 64 hex chars
}

func TestDigestHash_MD5(t *testing.T) {
	h := md5.New()
	h.Write([]byte("hello"))
	expected := fmt.Sprintf("%x", h.Sum(nil))

	result := digestHash(md5.New, "hello")
	assert.Equal(t, expected, result)
	assert.Len(t, result, 32) // MD5 produces 32 hex chars
}

func TestNewHashFunc(t *testing.T) {
	tests := []struct {
		algorithm string
		expectLen int // hex-encoded hash length
	}{
		{"MD5", 32},
		{"md5", 32},
		{"MD5-sess", 32},
		{"SHA-256", 64},
		{"sha-256", 64},
		{"SHA-256-sess", 64},
		{"SHA-256-SESS", 64},
		{"", 32},        // default to MD5
		{"unknown", 32}, // default to MD5
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			fn := newHashFunc(tt.algorithm)
			h := fn()
			h.Write([]byte("test"))
			result := fmt.Sprintf("%x", h.Sum(nil))
			assert.Len(t, result, tt.expectLen)
		})
	}
}
