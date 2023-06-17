package main

import (
	"bytes"
	"fmt"
)

const (
	_ = iota
	ARP_OPERATION_CODE_REQUEST
	ARP_OPERATION_CODE_REPLY
)

const ARP_HTYPE_ETHERNET uint16 = 1

var ArpTableEntryList []arpTableEntry

type arpIPToEthernet struct {
	hardwareType       uint16
	protocolType       uint16
	hardwareLen        uint8
	protocolLen        uint8
	opcode             uint16
	senderHardwareAddr [6]uint8  // sender MAC address
	senderIPAddr       IpAddress // sender IP address
	targetHardwareAddr [6]uint8  // target MAC address
	targetIPAddr       IpAddress // target IP address
}

// nolint unused
type arpTableEntry struct {
	macAddr [6]uint8
	ipAddr  IpAddress
	netdev  *netDevice
}

func (msg arpIPToEthernet) ToPacket() []byte {
	var b bytes.Buffer
	b.Write(uint16ToBytes(msg.hardwareType))
	b.Write(uint16ToBytes(msg.protocolType))
	b.Write([]byte{msg.hardwareLen})
	b.Write([]byte{msg.protocolLen})
	b.Write(uint16ToBytes(msg.opcode))
	b.Write(macToByte(msg.senderHardwareAddr))
	b.Write(uint32ToBytes(uint32(msg.senderIPAddr)))
	b.Write(macToByte(msg.targetHardwareAddr))
	b.Write(uint32ToBytes(uint32(msg.targetIPAddr)))

	return b.Bytes()
}

// arpInput receives the ARP packet
func arpInput(netdev *netDevice, packet []byte) error {
	if len(packet) < 28 {
		return fmt.Errorf("invalid ARP packet: length is too short (length=%d)", len(packet))
	}

	arpMsg := arpIPToEthernet{
		hardwareType:       byteToUint16(packet[0:2]),
		protocolType:       byteToUint16(packet[2:4]),
		hardwareLen:        packet[4],
		protocolLen:        packet[5],
		opcode:             byteToUint16(packet[6:8]),
		senderHardwareAddr: setMacAddr(packet[8:14]),
		senderIPAddr:       IpAddress(byteToUint32(packet[14:18])),
		targetHardwareAddr: setMacAddr(packet[18:24]),
		targetIPAddr:       IpAddress(byteToUint32(packet[24:28])),
	}

	if arpMsg.protocolType != ETHER_TYPE_IP {
		return fmt.Errorf("invalid protocol type: %d", arpMsg.protocolType)
	}
	if arpMsg.hardwareLen != ETHERNET_ADDRESS_LEN {
		return fmt.Errorf("invalid hardware address length: %d", arpMsg.hardwareLen)
	}
	if arpMsg.protocolLen != IP_ADDRESS_LEN {
		return fmt.Errorf("invalid protocol address length: %d", arpMsg.protocolLen)
	}

	switch arpMsg.opcode {
	case ARP_OPERATION_CODE_REQUEST:
		fmt.Printf("received the ARP request packet: %+v\n", arpMsg)
		if err := ReceiveARPRequest(netdev, arpMsg); err != nil {
			return fmt.Errorf("failed to receive ARP request packet: %w", err)
		}
	case ARP_OPERATION_CODE_REPLY:
		fmt.Printf("received the ARP reply packet: %+v\n", arpMsg)
		ReceiveARPReply(netdev, arpMsg)
	}

	return nil
}

// nolint: unused
func getArpTableEntry(ipAddr IpAddress) ([6]uint8, *netDevice) {
	for _, arpTable := range ArpTableEntryList {
		if arpTable.ipAddr == ipAddr {
			return arpTable.macAddr, arpTable.netdev
		}
	}
	return [6]uint8{}, nil
}

// ReceiveARPRequest receives the ARP request packet
func ReceiveARPRequest(netdev *netDevice, arp arpIPToEthernet) error {
	if netdev.ipdev.address == 0 || netdev.ipdev.address != arp.targetIPAddr {
		return fmt.Errorf("invalid address: %s", netdev.ipdev.address)
	}

	fmt.Printf("Sending ARP reply to %s\n", arp.targetIPAddr)
	arpPacket := arpIPToEthernet{
		hardwareType:       ARP_HTYPE_ETHERNET,
		protocolType:       ETHER_TYPE_IP,
		hardwareLen:        ETHERNET_ADDRESS_LEN,
		protocolLen:        IP_ADDRESS_LEN,
		opcode:             ARP_OPERATION_CODE_REPLY,
		senderHardwareAddr: netdev.macaddr,
		senderIPAddr:       netdev.ipdev.address,
		targetHardwareAddr: arp.senderHardwareAddr,
		targetIPAddr:       arp.senderIPAddr,
	}.ToPacket()

	if err := ethernetOutput(netdev, arp.senderHardwareAddr, arpPacket, ETHER_TYPE_ARP); err != nil {
		return fmt.Errorf("failed to output ethernet: %v", err)
	}
	return nil
}

// ReceiveARPReply receives the ARP request packet
func ReceiveARPReply(netdev *netDevice, arp arpIPToEthernet) {

}
