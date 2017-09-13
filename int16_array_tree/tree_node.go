// Package template is the base of code generation for type-specific trees
package int16_array_tree

import (
	"github.com/kentik/patricia"
)

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    [][]int16
}

func (n *treeNode) AddTag(tag []int16) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = [][]int16{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}

type treeNodeV4 struct {
	treeNode
	Left         *treeNodeV4 // left node
	Right        *treeNodeV4 // right node
	prefix       uint32
	prefixLength uint
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

func (n *treeNodeV6) ShiftLeft(bitCount uint) {
	n.prefixLeft, n.prefixRight, n.prefixLength = patricia.ShiftLeftIPv6(n.prefixLeft, n.prefixRight, n.prefixLength, bitCount)
}
