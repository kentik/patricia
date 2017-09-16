package template

import (
	"testing"

	"github.com/kentik/patricia"
	"github.com/stretchr/testify/assert"
)

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
