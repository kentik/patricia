package template

import (
	"testing"

	"github.com/kentik/patricia"
	"github.com/stretchr/testify/assert"
)

func TestV6MatchCount(t *testing.T) {
	node := &treeNodeV6{
		prefixLeft:   uint64(0xFFFFFFFFFFFFFFFF),
		prefixRight:  uint64(0xFFFFFFFFFFFFFFFF),
		prefixLength: 84,
	}

	// test moving the non-matching bit forward, making sure we never return more than the length of the prefix
	left := uint64(0xFFFFFFFFFFFFFFFF)
	for i := 1; i < 128; i++ {
		addressLeft := left
		if i <= 64 {
			addressLeft = clearBit64(addressLeft, uint64(i))
		}

		expected := i - 1
		if i > 64 {
			expected = 84
		}

		address := &patricia.IPv6Address{
			Left:   addressLeft,
			Right:  uint64(0xFFFFFFFFFFFFFFFF),
			Length: 128,
		}
		assert.Equal(t, uint(expected), node.MatchCount(address))
	}

	// test moving the non-matching bit forward, making sure we never return more than the length of the address
	node.prefixLength = 128
	left = uint64(0xFFFFFFFFFFFFFFFF)
	for i := 1; i < 128; i++ {
		addressLeft := left
		if i <= 64 {
			addressLeft = clearBit64(addressLeft, uint64(i))
		}

		expected := i - 1
		if i > 64 {
			expected = 84
		}
		address := &patricia.IPv6Address{
			Left:   addressLeft,
			Right:  uint64(0xFFFFFFFFFFFFFFFF),
			Length: 84,
		}
		assert.Equal(t, uint(expected), node.MatchCount(address))
	}
}

func TestV4MatchCount(t *testing.T) {
	node := &treeNodeV4{
		prefix:       uint32(0xFFFFFFFF),
		prefixLength: 32,
	}

	// test moving the non-matching bit forward, making sure we never return more than the length of the prefix
	addr := uint32(0xFFFFFFFF)
	for i := 1; i < 32; i++ {
		address := &patricia.IPv4Address{
			Address: clearBit32(addr, uint32(i)),
			Length:  32,
		}
		assert.Equal(t, uint(i-1), node.MatchCount(address))
	}

	// test moving the non-matching bit forward, with max of node address's 16 length
	node.prefixLength = 16
	for i := 1; i < 32; i++ {
		expected := uint(i - 1)
		if expected > 16 {
			expected = 16
		}

		address := &patricia.IPv4Address{
			Address: clearBit32(addr, uint32(i)),
			Length:  32,
		}
		assert.Equal(t, expected, node.MatchCount(address))
	}

	// test moving the non-matching bit forward, with max of comparison address's 16 length
	node.prefixLength = 32
	for i := 1; i < 32; i++ {
		expected := uint(i - 1)
		if expected > 16 {
			expected = 16
		}

		address := &patricia.IPv4Address{
			Address: clearBit32(addr, uint32(i)),
			Length:  16,
		}
		assert.Equal(t, expected, node.MatchCount(address))
	}
}

// Clears the bit at pos in n.
func clearBit32(n uint32, pos uint32) uint32 {
	pos = 32 - pos
	mask := uint32(^(1 << pos))
	n &= mask
	return n
}

// Clears the bit at pos in n.
func clearBit64(n uint64, pos uint64) uint64 {
	pos = 64 - pos
	mask := uint64(^(1 << pos))
	n &= mask
	return n
}
