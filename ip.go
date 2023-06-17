package main

import "fmt"

const IP_ADDRESS_LEN = 4

type IpAddress uint32

// nolint: unused
type ipDevice struct {
	address   IpAddress
	netmask   uint32
	broadcast uint32
}

func (i IpAddress) String() string {
	ipbyte := uint32ToBytes(uint32(i))
	return fmt.Sprintf("%d.%d.%d.%d", ipbyte[0], ipbyte[1], ipbyte[2], ipbyte[3])
}
