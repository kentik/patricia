package uint64_tree

import (
	"fmt"

	"github.com/kentik/patricia"
)

// this is IPv4 tree code that's not very copy/paste friendly for when we transfer IPv4 code to IPv6

// create a new node in the tree, return its index
func (t *TreeV4) newNode(address patricia.IPv4Address, prefixLength uint) uint {
	availCount := len(t.availableIndexes)
	if availCount > 0 {
		index := t.availableIndexes[availCount-1]
		t.availableIndexes = t.availableIndexes[:availCount-1]
		t.nodes[index] = treeNodeV4{prefix: address.Address, prefixLength: prefixLength}
		return index
	}

	t.nodes = append(t.nodes, treeNodeV4{prefix: address.Address, prefixLength: prefixLength})
	return uint(len(t.nodes) - 1)
}

func (t *TreeV4) print() {
	for i := range t.nodes {
		fmt.Printf("%d: \tleft: %d, right: %d, prefix: %032b (%d), tags: (%d): %v\n", i, int(t.nodes[i].Left), int(t.nodes[i].Right), int(t.nodes[i].prefix), int(t.nodes[i].prefixLength), t.nodes[i].TagCount, t.tagsForNode(uint(i)))
	}
}
