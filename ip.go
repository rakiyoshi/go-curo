package main

import (
	"fmt"
	"log"
)

const IP_ADDRESS_LEN = 4
const IP_ADDRESS_LIMITED_BROADCAST uint32 = 0xffffffff
const (
	IP_PROTOCOL_NUM_ICMP uint8 = 0x01
	IP_PROTOCOL_NUM_TCP  uint8 = 0x06
	IP_PROTOCOL_NUM_UDP  uint8 = 0x11
)

type IpAddress uint32

type ipDevice struct {
	address IpAddress
	// nolint: unused
	netmask uint32
	// nolint: unused
	broadcast uint32
}

type ipHeader struct {
	version        uint8
	headerLen      uint8
	tos            uint8 // Type of Service
	totalLen       uint16
	identify       uint16
	flagOffset     uint16
	ttl            uint8
	protocol       uint8
	headerChecksum uint16
	srcAddr        uint32
	destAddr       uint32
}

func printIPAddr(ip uint32) string {
	ipbyte := uint32ToBytes(ip)
	return fmt.Sprintf("%d.%d.%d.%d", ipbyte[0], ipbyte[1], ipbyte[2], ipbyte[3])
}

func (i IpAddress) String() string {
	ipbyte := uint32ToBytes(uint32(i))
	return fmt.Sprintf("%d.%d.%d.%d", ipbyte[0], ipbyte[1], ipbyte[2], ipbyte[3])
}

func ipInput(inputdev *netDevice, packet []byte) error {
	if inputdev.ipdev.address == 0 {
		return nil
	}

	if len(packet) < 20 {
		return fmt.Errorf("packet length is too short: name=%s", inputdev.name)
	}
	ipheader := ipHeader{
		version:        packet[0] >> 4,
		headerLen:      packet[0] << 5 >> 5,
		tos:            packet[1],
		totalLen:       byteToUint16(packet[2:4]),
		identify:       byteToUint16(packet[4:6]),
		flagOffset:     byteToUint16(packet[6:8]),
		ttl:            packet[8],
		protocol:       packet[9],
		headerChecksum: byteToUint16(packet[10:12]),
		srcAddr:        byteToUint32(packet[12:16]),
		destAddr:       byteToUint32(packet[16:20]),
	}

	log.Printf("received IP in %s, packetType=%d, from=%s, to=%s",
		inputdev.name,
		ipheader.protocol,
		printIPAddr(ipheader.srcAddr),
		printIPAddr(ipheader.destAddr),
	)

	macaddr, _ := searchArpTableEntry(ipheader.srcAddr)
	if macaddr == [6]uint8{} {
		addArpTableEntry(inputdev, ipheader.srcAddr, inputdev.etheHeader.srcAddr)
	}

	switch ipheader.version {
	case 4:
		break
	case 6:
		// TODO: implement
		return fmt.Errorf("IPv6 is not supported")
	default:
		return fmt.Errorf("invalid IP version: %d", ipheader.version)
	}

	if ipheader.headerLen*4 > 20 {
		return fmt.Errorf("IP header iption is not supported")
	}

	if ipheader.destAddr == IP_ADDRESS_LIMITED_BROADCAST || uint32(inputdev.ipdev.address) == ipheader.destAddr {
		// handle message as this post is destination
		return ipInputToOurs(inputdev, &ipheader, packet[20:])
	}

	for _, dev := range netDeviceList {
		if dev.ipdev.address == IpAddress(ipheader.destAddr) || dev.ipdev.broadcast == ipheader.destAddr {
			return ipInputToOurs(inputdev, &ipheader, packet[20:])
		}
	}

	return nil
}

func ipInputToOurs(inputdev *netDevice, ipheader *ipHeader, packet []byte) error {
	// TODO: implement NAT

	switch ipheader.protocol {
	case IP_PROTOCOL_NUM_ICMP:
		fmt.Println("ICMP received")
	case IP_PROTOCOL_NUM_TCP:
		fmt.Println("TCP received")
	case IP_PROTOCOL_NUM_UDP:
		fmt.Println("UDP received")
	default:
		return fmt.Errorf("Unsupported IP protocol: %d", ipheader.protocol)
	}

	return nil
}
