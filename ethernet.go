package main

import (
	"bytes"
	"fmt"
)

const (
	ETHER_TYPE_IP        = 0x0800
	ETHER_TYPE_ARP       = 0x0806
	ETHER_TYPE_IPV6      = 0x086dd
	ETHERNET_ADDRESS_LEN = 6
)

var ETHERNET_ADDERSS_BROADCAST = [6]uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

type ethernetHeader struct {
	destAddr  [6]uint8 // destination MAC address
	srcAddr   [6]uint8 // source MAC address
	etherType uint16   // ether type
}

func (header ethernetHeader) ToPacket() []byte {
	var b bytes.Buffer
	b.Write(macToByte(header.destAddr))
	b.Write(macToByte(header.srcAddr))
	b.Write(uint16ToBytes(header.etherType))
	return b.Bytes()
}

func setMacAddr(macAddrByte []byte) (result [6]uint8) {
	copy(result[:], macAddrByte)
	return
}

func macToByte(addr [6]uint8) (b []byte) {
	for _, v := range addr {
		b = append(b, v)
	}
	return
}

// ethernetInput processes the received data in ethernet
func ethernetInput(netdev *netDevice, packet []byte) error {
	// parse data as ethernet frame
	netdev.etheHeader.destAddr = setMacAddr(packet[0:6])
	netdev.etheHeader.srcAddr = setMacAddr(packet[6:12])
	netdev.etheHeader.etherType = byteToUint16(packet[12:14])

	if netdev.macaddr != netdev.etheHeader.destAddr && netdev.etheHeader.destAddr != ETHERNET_ADDERSS_BROADCAST {
		return nil
	}

	// detect protocol of upper layer
	switch netdev.etheHeader.etherType {
	case ETHER_TYPE_ARP:
		if err := arpInput(netdev, packet[14:]); err != nil {
			return fmt.Errorf("failed to input ARP packet: %w", err)
		}
		// case ETHER_TYPE_IP:
		// 	if err := ipInput(netdev, packet[14:]); err != nil {
		// 		return fmt.Errorf("failed to input IP packet: %w", err)
		// 	}
	}

	return nil
}

// ethernetOutput sends ethernet packet
func ethernetOutput(netdev *netDevice, destAddr [6]uint8, packet []byte, ethType uint16) error {
	// create ethernet header packet
	ethHeaderPacket := ethernetHeader{
		destAddr:  destAddr,
		srcAddr:   netdev.macaddr,
		etherType: ethType,
	}.ToPacket()

	ethHeaderPacket = append(ethHeaderPacket, packet...)
	if err := netdev.netDeviceTransmit(ethHeaderPacket); err != nil {
		return fmt.Errorf("failed to output ethernet: %v", err)
	}

	return nil
}
