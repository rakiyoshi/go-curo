package main

import "encoding/binary"

func byteToUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func byteToUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func uint16ToBytes(i uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return b
}

func uint32ToBytes(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}

func sumByteArr(packet []byte) (sum uint) {
	for i := range packet {
		if i%2 == 0 {
			sum += uint(byteToUint16(packet[i:]))
		}
	}
	return sum
}

func calcCechksum(packet []byte) []byte {
	sum := sumByteArr(packet)
	sum = (sum & 0xffff) + sum>>16
	return uint16ToBytes(uint16(sum ^ 0xffff))
}
