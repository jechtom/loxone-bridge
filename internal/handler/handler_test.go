package handler

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	HealthHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestBridgeHandler_RootPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	BridgeHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBridgeHandler_InvalidPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ftp/10.0.0.1/test", nil)
	rec := httptest.NewRecorder()

	BridgeHandler(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBridgeHandler_HTTPProxy(t *testing.T) {
	// Create a test upstream server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "ok"}`)
	}))
	defer upstream.Close()

	// Extract host from upstream URL
	host := strings.TrimPrefix(upstream.URL, "http://")

	req := httptest.NewRequest(http.MethodGet, "/http/"+host+"/api/status", nil)
	rec := httptest.NewRecorder()

	BridgeHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status": "ok"`)
}

func TestBridgeHandler_FlattenJSON(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"data": {"volume": 124, "error": false}, "name": "device-1"}`)
	}))
	defer upstream.Close()

	host := strings.TrimPrefix(upstream.URL, "http://")

	req := httptest.NewRequest(http.MethodGet, "/flatten-json/http/"+host+"/api/data", nil)
	rec := httptest.NewRecorder()

	BridgeHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "data.volume=124")
	assert.Contains(t, body, "data.error=false")
	assert.Contains(t, body, "name=device-1")
}

func TestBridgeHandler_UDP(t *testing.T) {
	// Start a UDP listener
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer pc.Close()

	addr := pc.LocalAddr().String()

	// Channel to receive data
	received := make(chan string, 1)
	go func() {
		buf := make([]byte, 1024)
		n, _, _ := pc.ReadFrom(buf)
		received <- string(buf[:n])
	}()

	req := httptest.NewRequest(http.MethodGet, "/udp/"+addr+"/hello-data", nil)
	rec := httptest.NewRecorder()

	BridgeHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "UDP sent to")

	data := <-received
	assert.Equal(t, "hello-data", data)
}

func TestBridgeHandler_UDPWithBody(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer pc.Close()

	addr := pc.LocalAddr().String()

	received := make(chan string, 1)
	go func() {
		buf := make([]byte, 1024)
		n, _, _ := pc.ReadFrom(buf)
		received <- string(buf[:n])
	}()

	body := strings.NewReader("custom-body-data")
	req := httptest.NewRequest(http.MethodPost, "/udp/"+addr+"/ignored-path", body)
	rec := httptest.NewRecorder()

	BridgeHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	data := <-received
	assert.Equal(t, "custom-body-data", data)
}

func TestBridgeHandler_QueryStringPassthrough(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/cgi-bin/control", r.URL.Path)
		assert.Equal(t, "action=open&channel=1", r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	}))
	defer upstream.Close()

	host := strings.TrimPrefix(upstream.URL, "http://")

	req := httptest.NewRequest(http.MethodGet, "/http/"+host+"/cgi-bin/control?action=open&channel=1", nil)
	rec := httptest.NewRecorder()

	BridgeHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
