package main

import (
	"fmt"
	"syscall"
)

var IGNORE_INTERFACES = map[string]struct{}{
	"lo":     {},
	"bond0":  {},
	"dummy0": {},
	"tunl0":  {},
	"sit0":   {},
}

type netDevice struct {
	name     string
	macaddr  [6]uint8
	socket   int
	sockaddr syscall.SockaddrLinklayer
	// nolint: unused
	etheHeader ethernetHeader
}

// isIgnoreInterfaces returns true when the name of the interface should be ignored.
func isIgnoreInterfaces(name string) bool {
	_, ok := IGNORE_INTERFACES[name]
	return ok
}

// htons converts a short (uint16) from host-to-network byte order.
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

func (netdev *netDevice) netDevicePoll(mode string) error {
	recvbuffer := make([]byte, 1500)

	n, _, err := syscall.Recvfrom(netdev.socket, recvbuffer, 0)
	if err != nil {
		if n == -1 {
			return nil
		}
		return fmt.Errorf("failed to receive, n = %d, device = %s: %v", n, netdev.name, err)
	}

	switch mode {
	case "ch1":
		fmt.Printf("Received %d bytes from %s: %x\n", n, netdev.name, recvbuffer[:n])
	default:
	}

	return nil
}
