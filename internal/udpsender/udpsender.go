// Package udpsender provides UDP datagram sending functionality.
package udpsender

import (
	"fmt"
	"net"
	"strings"
)

// Send sends data as a UDP datagram to the specified address.
// The address must include a port (e.g., "192.168.1.10:444").
// The data string is sent as raw bytes.
func Send(address string, data string) error {
	if !strings.Contains(address, ":") {
		return fmt.Errorf("UDP address must include a port (e.g., 192.168.1.10:444)")
	}

	conn, err := net.Dial("udp", address)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", address, err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("sending UDP data to %s: %w", address, err)
	}

	return nil
}
