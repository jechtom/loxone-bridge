package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDo_SimpleGET(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "hello")
	}))
	defer server.Close()

	resp, err := Do("GET", server.URL+"/test", nil, "", "", false, false, http.Header{})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDo_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-value", r.Header.Get("X-Custom-Header"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	headers := http.Header{}
	headers.Set("X-Custom-Header", "test-value")

	resp, err := Do("GET", server.URL+"/test", nil, "", "", false, false, headers)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDo_SkipsAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization header should NOT be forwarded
		assert.Empty(t, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	headers := http.Header{}
	headers.Set("Authorization", "Basic dXNlcjpwYXNz")

	resp, err := Do("GET", server.URL+"/test", nil, "", "", false, false, headers)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDo_DigestAuth(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First request: send digest challenge
			w.Header().Set("WWW-Authenticate", `Digest realm="test", nonce="abc123", qop="auth"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Second request: verify digest auth header is present
		auth := r.Header.Get("Authorization")
		assert.Contains(t, auth, "Digest")
		assert.Contains(t, auth, `username="admin"`)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "authenticated")
	}))
	defer server.Close()

	resp, err := Do("GET", server.URL+"/secure", nil, "admin", "password", true, false, http.Header{})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, callCount)
}
