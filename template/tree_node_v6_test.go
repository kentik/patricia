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

		address := patricia.IPv6Address{
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
		address := patricia.IPv6Address{
			Left:   addressLeft,
			Right:  uint64(0xFFFFFFFFFFFFFFFF),
			Length: 84,
		}
		assert.Equal(t, uint(expected), node.MatchCount(address))
	}
}
