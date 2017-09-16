package byte_array_tree

import (
	"github.com/kentik/patricia"
)

type treeNodeV4 struct {
	Left         uint // left node index: -1 for not set
	Right        uint // right node index: -1 for not set
	prefix       uint32
	prefixLength uint
	TagCount     uint
}

// See how many bits match the input address
func (n *treeNodeV4) MatchCount(address *patricia.IPv4Address) uint {
	var length uint
	if address.Length > n.prefixLength {
		length = n.prefixLength
	} else {
		length = address.Length
	}

	matches := uint(patricia.LeadingZeros32(n.prefix ^ address.Address))
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
