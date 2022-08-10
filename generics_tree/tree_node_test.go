package generics_tree

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
