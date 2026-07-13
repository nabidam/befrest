package netinfo

import (
	"fmt"
	"net"
	"time"
)

const probeTimeout = 2 * time.Second

// Probe verifies that the advertised LAN listener can be reached locally.
func Probe(address string, port int) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(address, fmt.Sprint(port)), probeTimeout)
	if err != nil {
		return err
	}
	return conn.Close()
}
