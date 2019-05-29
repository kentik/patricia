package patricia

import (
	"encoding/binary"
	"net"
)

const _leftmost64Bit = uint64(1 << 63)

// IPv6Address is a representation of an IPv6 address and CIDR
type IPv6Address struct {
	Left   uint64
	Right  uint64
	Length uint
}

// NewIPv6Address creates an address from the input IPv6 bytes (must be length 16)
func NewIPv6Address(address []byte, length uint) IPv6Address {
	if len(address) < 16 {
		return IPv6Address{
			Left:   0,
			Right:  0,
			Length: 0,
		}
	}
	return IPv6Address{
		Left:   binary.BigEndian.Uint64(address),
		Right:  binary.BigEndian.Uint64(address[8:]),
		Length: length,
	}
}

// ShiftLeft shifts the bits |bitCount| bits left
func (ip *IPv6Address) ShiftLeft(bitCount uint) {
	ip.Left, ip.Right, ip.Length = ShiftLeftIPv6(ip.Left, ip.Right, ip.Length, bitCount)
}

// String returns a string version of this IP address.
// - not optimized for performance, alloates a byte slice
func (ip IPv6Address) String() string {
	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data, ip.Left)
	binary.BigEndian.PutUint64(data[8:], ip.Right)

	ipNet := net.IPNet{
		IP:   data,
		Mask: net.CIDRMask(int(ip.Length), 128),
	}
	return ipNet.String()
}

// ShiftLeftIPv6 shifts IPv6 (as two uint64's) to the left
func ShiftLeftIPv6(left uint64, right uint64, length uint, bitCount uint) (uint64, uint64, uint) {
	length = length - bitCount
	if bitCount >= 64 {
		// we don't care about the right bit - move it left, then shift shift-64
		return right << (bitCount - 64), 0, length
	}

	// shifting less than 64 - need to shift right over to left
	left = (left << bitCount) | (right >> (64 - bitCount))
	right <<= bitCount

	return left, right, length
}

// ShiftRightIPv6 shifts IPv6 (as two uint64's) to the right
// - assumes left and rights have already been masked clean, so there's no extra bits
func ShiftRightIPv6(left uint64, right uint64, bitCount uint) (uint64, uint64) {
	if bitCount >= 64 {
		// shifting by at least 64 - just move left to right, and shift the remaining amount
		return 0, left >> (bitCount - 64)
	}

	// shifting less than 64 - need to shift left over to right
	right = (right >> bitCount) | (left << (64 - bitCount))
	left >>= bitCount
	return left, right
}

// IsLeftBitSet returns whether the leftmost bit is set
func (ip *IPv6Address) IsLeftBitSet() bool {
	return ip.Left >= _leftmost64Bit
}
