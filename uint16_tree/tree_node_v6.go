package uint16_tree

import (
	"github.com/kentik/patricia"
)

type treeNodeV6 struct {
	treeNode
	Left         *treeNodeV6 // left node
	Right        *treeNodeV6 // right node
	prefixLeft   uint64
	prefixRight  uint64
	prefixLength uint
}

func (n *treeNodeV6) MatchCount(address *patricia.IPv6Address) uint {
	length := address.Length
	if length > n.prefixLength {
		length = n.prefixLength
	}

	matches := uint(patricia.LeadingZeros64(n.prefixLeft ^ address.Left))
	if matches == 64 && length > 64 {
		matches += uint(patricia.LeadingZeros64(n.prefixRight ^ address.Right))
	}
	if matches > length {
		return length
	}
	return matches
}

// ShiftPrefix shifts the prefix by the input shiftCount
func (n *treeNodeV6) ShiftPrefix(shiftCount uint) {
	n.prefixLeft, n.prefixRight, n.prefixLength = patricia.ShiftLeftIPv6(n.prefixLeft, n.prefixRight, n.prefixLength, shiftCount)
}
