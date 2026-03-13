package udpsender

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend_Success(t *testing.T) {
	// Start a UDP listener
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

	err = Send(addr, "test-data")
	require.NoError(t, err)

	data := <-received
	assert.Equal(t, "test-data", data)
}

func TestSend_MissingPort(t *testing.T) {
	err := Send("192.168.1.10", "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must include a port")
}

func TestSend_EmptyData(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer pc.Close()

	addr := pc.LocalAddr().String()
	err = Send(addr, "")
	// Empty data is valid for UDP
	assert.NoError(t, err)
}
