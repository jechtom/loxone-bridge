// Package urlparser provides URL parsing for LoxoneBridge requests.
// It extracts modifiers, protocol, target address, and downstream path
// from incoming request URLs.
package urlparser

import (
	"fmt"
	"strings"
)

// Modifier constants.
const (
	ModDigest      = "digest"
	ModFlattenJSON = "flatten-json"
)

// Protocol constants.
const (
	ProtoHTTP            = "http"
	ProtoHTTPS           = "https"
	ProtoHTTPSIgnoreCert = "https-ignore-cert"
	ProtoUDP             = "udp"
)

// ParsedURL holds the parsed components of a LoxoneBridge request URL.
type ParsedURL struct {
	Modifiers []string // e.g., ["digest", "flatten-json"]
	Protocol  string   // e.g., "http", "https", "https-ignore-cert", "udp"
	Address   string   // e.g., "192.168.1.10" or "192.168.1.10:8080"
	Path      string   // remaining path + query, e.g., "/cgi-bin/foo?bar=1"
}

// IsModifier returns true if the segment is a known modifier.
func IsModifier(s string) bool {
	switch s {
	case ModDigest, ModFlattenJSON:
		return true
	}
	return false
}

// IsProtocol returns true if the segment is a known protocol.
func IsProtocol(s string) bool {
	switch s {
	case ProtoHTTP, ProtoHTTPS, ProtoHTTPSIgnoreCert, ProtoUDP:
		return true
	}
	return false
}

// Parse parses the incoming request path into its components.
// The expected format is: /{modifiers...}/{protocol}/{address}/{path...}
//
// Examples:
//
//	/digest/http/192.168.1.10/cgi-bin/foo?bar=1
//	/https-ignore-cert/192.168.1.10/some/path
//	/flatten-json/http/192.168.1.10/api/data
//	/udp/192.168.1.10:444/data
func Parse(requestPath string, rawQuery string) (*ParsedURL, error) {
	// Remove leading slash and split
	path := strings.TrimPrefix(requestPath, "/")
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}

	segments := strings.SplitN(path, "/", -1)
	if len(segments) < 2 {
		return nil, fmt.Errorf("path too short: need at least protocol and address")
	}

	result := &ParsedURL{}
	idx := 0

	// Consume modifiers
	for idx < len(segments) && IsModifier(segments[idx]) {
		result.Modifiers = append(result.Modifiers, segments[idx])
		idx++
	}

	// Next must be protocol
	if idx >= len(segments) {
		return nil, fmt.Errorf("missing protocol segment")
	}
	if !IsProtocol(segments[idx]) {
		return nil, fmt.Errorf("unknown protocol: %s", segments[idx])
	}
	result.Protocol = segments[idx]
	idx++

	// Next must be address
	if idx >= len(segments) {
		return nil, fmt.Errorf("missing address segment")
	}
	result.Address = segments[idx]
	idx++

	// Remaining segments form the downstream path
	if idx < len(segments) {
		result.Path = "/" + strings.Join(segments[idx:], "/")
	} else {
		result.Path = "/"
	}

	// Append query string if present
	if rawQuery != "" {
		result.Path += "?" + rawQuery
	}

	return result, nil
}

// TargetURL constructs the full target URL from the parsed components.
func (p *ParsedURL) TargetURL() string {
	scheme := p.Protocol
	if scheme == ProtoHTTPSIgnoreCert {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, p.Address, p.Path)
}

// HasModifier checks if a specific modifier is present.
func (p *ParsedURL) HasModifier(mod string) bool {
	for _, m := range p.Modifiers {
		if m == mod {
			return true
		}
	}
	return false
}
