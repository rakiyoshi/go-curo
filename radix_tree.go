package main

// The binary tree node for longest prefix matching of IP address
type radixTreeNode struct {
	depth  int
	parent *radixTreeNode
	node0  *radixTreeNode
	node1  *radixTreeNode
	data   ipRouteEntry
	value  int
}

func (n *radixTreeNode) radixTreeAdd(prefixIpAddr, prefixLen uint32, entryData ipRouteEntry) {
	current := n

	for d := 1; d < int(prefixLen); d++ {
		// check the first `d` bit
		switch prefixIpAddr >> (32 - d) & 0x01 {
		case 0:
			if current.node0 == nil {
				current.node0 = &radixTreeNode{
					parent: current,
					depth:  d,
					value:  0,
				}
			}
			current = current.node0
		case 1:
			if current.node1 == nil {
				current.node1 = &radixTreeNode{
					parent: current,
					depth:  d,
					value:  0,
				}
			}
			current = current.node1
		}
	}
	current.data = entryData
}

// nolint:unused
func (n *radixTreeNode) radixTreeSearch(prefixIpAddr uint32) (result ipRouteEntry) {
	current := n

	for i := 1; i < 32; i++ {
		if current.data != (ipRouteEntry{}) {
			result = current.data
		}
		switch (prefixIpAddr >> uint32(32-i)) & 0x01 {
		case 0:
			if current.node0 == nil {
				return
			}
			current = current.node1
		case 1:
			if current.node1 == nil {
				return
			}
			current = current.node1
		}
	}

	return
}
