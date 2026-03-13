// Package handler provides the main HTTP handler for LoxoneBridge.
package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/loxone-bridge/internal/flatten"
	"github.com/loxone-bridge/internal/proxy"
	"github.com/loxone-bridge/internal/udpsender"
	"github.com/loxone-bridge/internal/urlparser"
)

// HealthHandler responds with a simple health check.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

// BridgeHandler is the main request handler that parses the URL,
// determines the protocol and modifiers, and proxies the request accordingly.
func BridgeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "" {
		http.Error(w, "LoxoneBridge is running. Provide a target path.", http.StatusBadRequest)
		return
	}

	parsed, err := urlparser.Parse(r.URL.Path, r.URL.RawQuery)
	if err != nil {
		log.Printf("ERROR: parse URL %s: %v", r.URL.Path, err)
		http.Error(w, fmt.Sprintf("Invalid URL: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("REQ: %s %s -> proto=%s addr=%s path=%s modifiers=%v",
		r.Method, r.URL.Path, parsed.Protocol, parsed.Address, parsed.Path, parsed.Modifiers)

	switch parsed.Protocol {
	case urlparser.ProtoHTTP, urlparser.ProtoHTTPS, urlparser.ProtoHTTPSIgnoreCert:
		handleHTTP(w, r, parsed)
	case urlparser.ProtoUDP:
		handleUDP(w, r, parsed)
	default:
		http.Error(w, fmt.Sprintf("Unsupported protocol: %s", parsed.Protocol), http.StatusBadRequest)
	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request, parsed *urlparser.ParsedURL) {
	// Extract basic auth credentials from the incoming request
	username, password, _ := r.BasicAuth()

	useDigest := parsed.HasModifier(urlparser.ModDigest)
	ignoreCert := parsed.Protocol == urlparser.ProtoHTTPSIgnoreCert

	targetURL := parsed.TargetURL()

	resp, err := proxy.Do(r.Method, targetURL, r.Body, username, password, useDigest, ignoreCert, r.Header)
	if err != nil {
		log.Printf("ERROR: proxy to %s: %v", targetURL, err)
		http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: reading response from %s: %v", targetURL, err)
		http.Error(w, "Error reading upstream response", http.StatusBadGateway)
		return
	}

	// Apply flatten-json modifier if requested
	if parsed.HasModifier(urlparser.ModFlattenJSON) {
		flattened, flatErr := flatten.JSON(body)
		if flatErr != nil {
			log.Printf("ERROR: flatten JSON from %s: %v", targetURL, flatErr)
			http.Error(w, fmt.Sprintf("JSON flatten error: %v", flatErr), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(resp.StatusCode)
		fmt.Fprint(w, flattened)
		return
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func handleUDP(w http.ResponseWriter, r *http.Request, parsed *urlparser.ParsedURL) {
	// For UDP, the path becomes the data to send
	data := strings.TrimPrefix(parsed.Path, "/")

	// Also read body if present
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("ERROR: reading UDP body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		if len(bodyBytes) > 0 {
			data = string(bodyBytes)
		}
	}

	if data == "" {
		http.Error(w, "No data to send via UDP", http.StatusBadRequest)
		return
	}

	if err := udpsender.Send(parsed.Address, data); err != nil {
		log.Printf("ERROR: UDP send to %s: %v", parsed.Address, err)
		http.Error(w, fmt.Sprintf("UDP send error: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "UDP sent to %s", parsed.Address)
}
