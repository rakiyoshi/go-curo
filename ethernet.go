package main

// nolint:unused
type ethernetHeader struct {
	destAddr  [6]uint8 // destination MAC address
	srcAddr   [6]uint8 // source MAC address
	etherType uint16   // ether type
}

func setMacAddr(macAddrByte []byte) (result [6]uint8) {
	copy(result[:], macAddrByte)
	return
}
