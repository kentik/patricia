package uint16_tree

import (
	"math/bits"

	"github.com/kentik/patricia"
)

const _leftmost32Bit = uint32(1 << 31)

type treeNodeV4 struct {
	Left         uint // left node index: 0 for not set
	Right        uint // right node index: 0 for not set
	prefix       uint32
	prefixLength uint
	TagCount     int
}

// See how many bits match the input address
func (n *treeNodeV4) MatchCount(address patricia.IPv4Address) uint {
	var length uint
	if address.Length > n.prefixLength {
		length = n.prefixLength
	} else {
		length = address.Length
	}

	matches := uint(bits.LeadingZeros32(n.prefix ^ address.Address))
	if matches > length {
		return length
	}
	return matches
}

// ShiftPrefix shifts the prefix by the input shiftCount
func (n *treeNodeV4) ShiftPrefix(shiftCount uint) {
	n.prefix <<= shiftCount
	n.prefixLength -= shiftCount
}

// IsLeftBitSet returns whether the leftmost bit is set
func (n *treeNodeV4) IsLeftBitSet() bool {
	return n.prefix >= _leftmost32Bit
}

// MergeFromNodes updates the prefix and prefix length from the two input nodes
func (n *treeNodeV4) MergeFromNodes(left *treeNodeV4, right *treeNodeV4) {
	n.prefix, n.prefixLength = patricia.MergePrefixes32(left.prefix, left.prefixLength, right.prefix, right.prefixLength)
}
