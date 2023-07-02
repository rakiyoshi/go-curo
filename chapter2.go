package main

import (
	"log"
	"net"
	"syscall"
)

var netDeviceList []*netDevice
var iproute radixTreeNode

func runChapter2() {
	// register route to host2
	routeEntryToHost2 := ipRouteEntry{
		iptype:  IpRouteTypeNetwork,
		nexthop: 0xc0a80002,
	}
	// register entry route to 192.168.2.0/24
	iproute.radixTreeAdd(0xc0a80202&0xffffff00, 24, routeEntryToHost2)

	// create epoll
	events := make([]syscall.EpollEvent, 10)
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

		log.Printf("Created device %s socket %d address %s",
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

		netaddrs, err := netif.Addrs()
		if err != nil {
			log.Fatalf("failed to get IP address from NIC interface: %v", err)
		}

		ipdev, err := getIPDevice(netaddrs)
		if err != nil {
			log.Fatalf("failed to get IP address from NIC interface: %v", err)
		}
		netdev := netDevice{
			name:     netif.Name,
			macaddr:  setMacAddr(netif.HardwareAddr),
			socket:   sock,
			sockaddr: addr,
			ipdev:    *ipdev,
		}

		routeEntry := ipRouteEntry{
			iptype: IpRouteTypeConnected,
			netdev: &netdev,
		}
		prefixLen := subnetToPrefixLen(netdev.ipdev.netmask)
		prefixIpAddr := uint32(netdev.ipdev.address) & netdev.ipdev.netmask
		iproute.radixTreeAdd(prefixIpAddr, prefixLen, routeEntry)
		log.Printf("Set directly connected route %s (%d via %s)",
			printIPAddr(prefixIpAddr), prefixLen, netdev.name,
		)

		netDeviceList = append(netDeviceList, &netdev)
	}

	for {
		nfds, err := syscall.EpollWait(epfd, events, -1)
		if err != nil {
			log.Fatalf("failed to EpollWait: %v", err)
		}
		for i := 0; i < nfds; i++ {

			for _, netdev := range netDeviceList {
				if events[i].Fd != int32(netdev.socket) {
					continue
				}
				if err := netdev.netDevicePoll("ch2"); err != nil {
					log.Fatalf("failed to net device poll: %v", err)
				}
			}
		}
	}
}
