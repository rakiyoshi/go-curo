package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
)

const IpAddressLen = 4
const IpAddressLimitedBroadcast IpAddress = 0xffffffff
const (
	IpProtocolNumICMP uint8 = 0x01
	IpProtocolNumTCP  uint8 = 0x06
	IpProtocolNumUDP  uint8 = 0x11
)
const (
	IpRouteTypeConnected ipRouteType = iota
	IpRouteTypeNetwork
)

type IpAddress uint32

type ipDevice struct {
	address IpAddress
	// nolint: unused
	netmask uint32
	// nolint: unused
	broadcast IpAddress
}

type ipHeader struct {
	version        uint8
	headerLen      uint8
	tos            uint8 // Type of Service
	totalLen       uint16
	identify       uint16
	fragmentOffset uint16
	ttl            uint8
	protocol       uint8
	headerChecksum uint16
	srcAddr        IpAddress
	destAddr       IpAddress
}

type ipRouteEntry struct {
	iptype  ipRouteType
	netdev  *netDevice
	nexthop uint32
}

type ipRouteType uint8

func (i ipHeader) ToPacket(calc bool) (ipHeaderByte []byte) {
	var b bytes.Buffer

	b.Write([]byte{i.version<<4 + i.headerLen})
	b.Write([]byte{i.tos})
	b.Write(uint16ToBytes(i.totalLen))
	b.Write(uint16ToBytes(i.identify))
	b.Write(uint16ToBytes(i.fragmentOffset))
	b.Write([]byte{i.ttl})
	b.Write([]byte{i.protocol})
	b.Write(uint16ToBytes(i.headerChecksum))
	b.Write(uint32ToBytes(uint32(i.srcAddr)))
	b.Write(uint32ToBytes(uint32(i.destAddr)))

	if calc {
		ipHeaderByte = b.Bytes()
		checksum := calcCechksum(ipHeaderByte)
		ipHeaderByte[10] = checksum[0]
		ipHeaderByte[11] = checksum[1]
	}

	return ipHeaderByte
}

func printIPAddr(ip uint32) string {
	ipbyte := uint32ToBytes(ip)
	return fmt.Sprintf("%d.%d.%d.%d", ipbyte[0], ipbyte[1], ipbyte[2], ipbyte[3])
}

// Convert to prefix length from netmask
// example: 0xffffff00 -> 24
func subnetToPrefixLen(netmask uint32) (prefixLen uint32) {
	for ; prefixLen < 32; prefixLen++ {
		if netmask>>(31-prefixLen)&1 == 0 {
			break
		}
	}
	return
}

func getIPDevice(addrs []net.Addr) (*ipDevice, error) {
	ipdev := &ipDevice{}
	for _, addr := range addrs {
		ipaddrstr := addr.String()
		if strings.Contains(ipaddrstr, ":") || strings.Contains(ipaddrstr, ".") {
			continue
		}
		ip, ipnet, err := net.ParseCIDR(ipaddrstr)
		if err != nil {
			return nil, fmt.Errorf("failed to pase CIDR: %w", err)
		}
		ipdev.address = IpAddress(byteToUint32(ip.To4()))
		ipdev.netmask = byteToUint32(ipnet.Mask)
		ipdev.broadcast = IpAddress(uint32(ipdev.address) | (^ipdev.netmask))
	}
	return ipdev, nil
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
		fragmentOffset: byteToUint16(packet[6:8]),
		ttl:            packet[8],
		protocol:       packet[9],
		headerChecksum: byteToUint16(packet[10:12]),
		srcAddr:        IpAddress(byteToUint32(packet[12:16])),
		destAddr:       IpAddress(byteToUint32(packet[16:20])),
	}

	log.Printf("received IP in %s, packetType=%d, from=%s, to=%s",
		inputdev.name,
		ipheader.protocol,
		printIPAddr(uint32(ipheader.srcAddr)),
		printIPAddr(uint32(ipheader.destAddr)),
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
		return fmt.Errorf("IP header option is not supported")
	}

	if ipheader.destAddr == IpAddressLimitedBroadcast || inputdev.ipdev.address == ipheader.destAddr {
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
	case IpProtocolNumICMP:
		fmt.Println("ICMP received")
	case IpProtocolNumTCP:
		fmt.Println("TCP received")
	case IpProtocolNumUDP:
		fmt.Println("UDP received")
	default:
		return fmt.Errorf("Unsupported IP protocol: %d", ipheader.protocol)
	}

	return nil
}

// nolint: unused
func ipPacketEncapsulateOutput(inputdev *netDevice, destAddr, srcAddr IpAddress, payload []byte, protocolType uint8) error {
	var ipPacket []byte

	// IP header length (=20) + packet length
	totalLength := 20 + len(payload)

	ipheader := ipHeader{
		version:        4,
		headerLen:      20 / 4,
		tos:            0,
		totalLen:       uint16(totalLength),
		identify:       0xf80c,
		fragmentOffset: 2 << 13,
		ttl:            0x40,
		protocol:       protocolType,
		headerChecksum: 0,
		srcAddr:        srcAddr,
		destAddr:       destAddr,
	}
	ipPacket = append(ipPacket, ipheader.ToPacket(true)...)
	ipPacket = append(ipPacket, payload...)

	destMacAddr, _ := searchArpTableEntry(destAddr)
	if destMacAddr != [6]uint8{0, 0, 0, 0, 0, 0} {
		if err := ethernetOutput(inputdev, destMacAddr, ipPacket, ETHER_TYPE_IP); err != nil {
			return err
		}
	} else {
		if err := sendArpRequest(inputdev, destAddr); err != nil {
			return err
		}
	}

	return nil
}
