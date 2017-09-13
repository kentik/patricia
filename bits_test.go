package patricia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeftMasks(t *testing.T) {
	// 32-bit
	assert.Equal(t, uint32(0), _leftMasks32[0])
	assert.Equal(t, uint32(0x80000000), _leftMasks32[1])
	assert.Equal(t, uint32(0xC0000000), _leftMasks32[2])
	assert.Equal(t, uint32(0xE0000000), _leftMasks32[3])
	assert.Equal(t, uint32(0xF0000000), _leftMasks32[4])
	assert.Equal(t, uint32(0xFFFFFFFE), _leftMasks32[31])
	assert.Equal(t, uint32(0xFFFFFFFF), _leftMasks32[32])

	// 64-bit
	assert.Equal(t, uint64(0), _leftMasks64[0])
	assert.Equal(t, uint64(0x8000000000000000), _leftMasks64[1])
	assert.Equal(t, uint64(0xC000000000000000), _leftMasks64[2])
	assert.Equal(t, uint64(0xE000000000000000), _leftMasks64[3])
	assert.Equal(t, uint64(0xF000000000000000), _leftMasks64[4])
	assert.Equal(t, uint64(0xFFFFFFFE00000000), _leftMasks64[31])
	assert.Equal(t, uint64(0xFFFFFFFF00000000), _leftMasks64[32])
	assert.Equal(t, uint64(0xFFFFFFFF80000000), _leftMasks64[33])
	assert.Equal(t, uint64(0xFFFFFFFFFFFFFFFE), _leftMasks64[63])
	assert.Equal(t, uint64(0xFFFFFFFFFFFFFFFF), _leftMasks64[64])
}

func TestMergePrefixes32(t *testing.T) {
	newPrefix, newLength := MergePrefixes32(uint32(0x88803000), uint(4), uint32(0x8FE30000), uint(4))
	assert.Equal(t, uint32(0x88000000), newPrefix)
	assert.Equal(t, uint(8), newLength)

	newPrefix, newLength = MergePrefixes32(uint32(0x80000000), 4, uint32(0xFFFFFFFF), 0)
	assert.Equal(t, uint32(0x80000000), newPrefix)
	assert.Equal(t, uint(4), newLength)
}
