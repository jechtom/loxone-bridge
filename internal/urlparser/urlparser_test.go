package urlparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_DigestHTTP(t *testing.T) {
	parsed, err := Parse("/digest/http/192.168.1.10/cgi-bin/accessControl.cgi", "action=openDoor&channel=1")
	require.NoError(t, err)

	assert.Equal(t, []string{"digest"}, parsed.Modifiers)
	assert.Equal(t, "http", parsed.Protocol)
	assert.Equal(t, "192.168.1.10", parsed.Address)
	assert.Equal(t, "/cgi-bin/accessControl.cgi?action=openDoor&channel=1", parsed.Path)
}

func TestParse_HTTPSIgnoreCert(t *testing.T) {
	parsed, err := Parse("/https-ignore-cert/192.168.1.10/aaa", "")
	require.NoError(t, err)

	assert.Empty(t, parsed.Modifiers)
	assert.Equal(t, "https-ignore-cert", parsed.Protocol)
	assert.Equal(t, "192.168.1.10", parsed.Address)
	assert.Equal(t, "/aaa", parsed.Path)
}

func TestParse_UDP(t *testing.T) {
	parsed, err := Parse("/udp/192.168.1.10:444/data", "")
	require.NoError(t, err)

	assert.Empty(t, parsed.Modifiers)
	assert.Equal(t, "udp", parsed.Protocol)
	assert.Equal(t, "192.168.1.10:444", parsed.Address)
	assert.Equal(t, "/data", parsed.Path)
}

func TestParse_FlattenJSON(t *testing.T) {
	parsed, err := Parse("/flatten-json/http/192.168.1.10/aaaa", "")
	require.NoError(t, err)

	assert.Equal(t, []string{"flatten-json"}, parsed.Modifiers)
	assert.Equal(t, "http", parsed.Protocol)
	assert.Equal(t, "192.168.1.10", parsed.Address)
	assert.Equal(t, "/aaaa", parsed.Path)
	assert.True(t, parsed.HasModifier(ModFlattenJSON))
}

func TestParse_MultipleModifiers(t *testing.T) {
	parsed, err := Parse("/digest/flatten-json/https/10.0.0.1/api/data", "")
	require.NoError(t, err)

	assert.Equal(t, []string{"digest", "flatten-json"}, parsed.Modifiers)
	assert.Equal(t, "https", parsed.Protocol)
	assert.Equal(t, "10.0.0.1", parsed.Address)
	assert.Equal(t, "/api/data", parsed.Path)
}

func TestParse_SimpleHTTP(t *testing.T) {
	parsed, err := Parse("/http/10.0.0.5/status", "")
	require.NoError(t, err)

	assert.Empty(t, parsed.Modifiers)
	assert.Equal(t, "http", parsed.Protocol)
	assert.Equal(t, "10.0.0.5", parsed.Address)
	assert.Equal(t, "/status", parsed.Path)
}

func TestParse_NoPath(t *testing.T) {
	parsed, err := Parse("/http/10.0.0.5", "")
	require.NoError(t, err)

	assert.Equal(t, "http", parsed.Protocol)
	assert.Equal(t, "10.0.0.5", parsed.Address)
	assert.Equal(t, "/", parsed.Path)
}

func TestParse_EmptyPath(t *testing.T) {
	_, err := Parse("", "")
	assert.Error(t, err)
}

func TestParse_UnknownProtocol(t *testing.T) {
	_, err := Parse("/ftp/10.0.0.1/file", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown protocol")
}

func TestTargetURL_HTTP(t *testing.T) {
	parsed := &ParsedURL{Protocol: "http", Address: "10.0.0.5", Path: "/api/data?key=val"}
	assert.Equal(t, "http://10.0.0.5/api/data?key=val", parsed.TargetURL())
}

func TestTargetURL_HTTPSIgnoreCert(t *testing.T) {
	parsed := &ParsedURL{Protocol: "https-ignore-cert", Address: "10.0.0.5:8443", Path: "/api"}
	assert.Equal(t, "https://10.0.0.5:8443/api", parsed.TargetURL())
}

func TestHasModifier(t *testing.T) {
	parsed := &ParsedURL{Modifiers: []string{"digest", "flatten-json"}}
	assert.True(t, parsed.HasModifier("digest"))
	assert.True(t, parsed.HasModifier("flatten-json"))
	assert.False(t, parsed.HasModifier("unknown"))
}

func TestIsModifier(t *testing.T) {
	assert.True(t, IsModifier("digest"))
	assert.True(t, IsModifier("flatten-json"))
	assert.False(t, IsModifier("http"))
	assert.False(t, IsModifier("unknown"))
}

func TestIsProtocol(t *testing.T) {
	assert.True(t, IsProtocol("http"))
	assert.True(t, IsProtocol("https"))
	assert.True(t, IsProtocol("https-ignore-cert"))
	assert.True(t, IsProtocol("udp"))
	assert.False(t, IsProtocol("ftp"))
	assert.False(t, IsProtocol("digest"))
}

func TestParse_DeepPath(t *testing.T) {
	parsed, err := Parse("/digest/http/10.0.0.5/cgi-bin/accessControl.cgi", "action=openDoor&channel=1")
	require.NoError(t, err)

	assert.Equal(t, "http://10.0.0.5/cgi-bin/accessControl.cgi?action=openDoor&channel=1", parsed.TargetURL())
}

func TestParse_AddressWithPort(t *testing.T) {
	parsed, err := Parse("/http/10.0.0.5:8080/api", "")
	require.NoError(t, err)

	assert.Equal(t, "10.0.0.5:8080", parsed.Address)
	assert.Equal(t, "http://10.0.0.5:8080/api", parsed.TargetURL())
}
