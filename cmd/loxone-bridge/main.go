// Package main is the entry point for the LoxoneBridge service.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/loxone-bridge/internal/handler"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	log.Printf("LoxoneBridge version %s", version)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.BridgeHandler)
	mux.HandleFunc("/healthz", handler.HealthHandler)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("LoxoneBridge starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
