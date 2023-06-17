package main

import (
	"fmt"
	"log"
	"net"
	"syscall"
)

func runChapter1() {
	var netDeviceList []netDevice

	events := make([]syscall.EpollEvent, 10)

	// create epoll
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatalf("epoll create err: %v", err)
	}

	// fetch the list of the system's network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("failed to fetch interfaces: %v", err)
	}

	for _, netif := range interfaces {
		if isIgnoreInterfaces(netif.Name) {
			continue
		}

		// open socket
		sock, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
		if err != nil {
			log.Fatalf("failed to create socket: %v", err)
		}

		// bind the interface to the socket
		addr := syscall.SockaddrLinklayer{
			Protocol: htons(syscall.ETH_P_ALL),
			Ifindex:  netif.Index,
		}
		if err := syscall.Bind(sock, &addr); err != nil {
			log.Fatalf("failed to bind the interface to the socket: %v", err)
		}

		fmt.Printf("Created device %s socket %d address %s\n",
			netif.Name,
			sock,
			netif.HardwareAddr.String(),
		)

		// monitor the socket by epoll
		if err := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, sock, &syscall.EpollEvent{
			Events: syscall.EPOLLIN,
			Fd:     int32(sock),
		}); err != nil {
			log.Fatalf("failed to epoll ctrl: %v", err)
		}

		// set non block
		// if err := syscall.SetNonblock(sock, true); err != nil {
		// 	log.Fatalf("failed to set non block: %v", err)
		// }

		netDeviceList = append(netDeviceList, netDevice{
			name:     netif.Name,
			macaddr:  setMacAddr(netif.HardwareAddr),
			socket:   sock,
			sockaddr: addr,
		})
	}

	for {
		// wait to receive packet by epoll_wait
		nfds, err := syscall.EpollWait(epfd, events, -1)
		if err != nil {
			log.Fatalf("failed to EpollWait: %v", err)
		}
		for i := 0; i < nfds; i++ {
			for _, netdev := range netDeviceList {
				if events[i].Fd != int32(netdev.socket) {
					continue
				}
				if err := netdev.netDevicePoll("ch1"); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

}
