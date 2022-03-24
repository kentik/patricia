package patricia

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIPv6Address(t *testing.T) {
	sut := NewIPv6Address([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xB0}, 117)

	assert.Equal(t, uint64(0x0123456789A0A1A2), sut.Left)
	assert.Equal(t, uint64(0xA3A4A5A6A7A8A9B0), sut.Right)
	assert.Equal(t, uint(117), sut.Length)
	//nolint
	assert.Equal(t, "123:4567:89a0:a1a2:a3a4:a5a6:a7a8:a9b0/117", fmt.Sprintf("%s", sut))
	//nolint
	assert.Equal(t, "123:4567:89a0:a1a2:a3a4:a5a6:a7a8:a9b0/117", fmt.Sprintf("%s", &sut))
	assert.Equal(t, "123:4567:89a0:a1a2:a3a4:a5a6:a7a8:a9b0/117", (&sut).String())
	assert.Equal(t, "123:4567:89a0:a1a2:a3a4:a5a6:a7a8:a9b0/117", sut.String())

	sut = NewIPv6Address([]byte{0x01, 0x02}, 100)
	assert.Equal(t, uint64(0), sut.Left)
	assert.Equal(t, uint64(0), sut.Right)
	assert.Equal(t, uint(0), sut.Length)
	assert.Equal(t, "::/0", sut.String())
}

func TestShiftLeftOneBit(t *testing.T) {
	sut := NewIPv6Address([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xB0}, 117)

	sut.ShiftLeft(uint(1))
	assert.Equal(t, uint64(0x2468ACF13414345), sut.Left)
	assert.Equal(t, uint64(0X47494B4D4F515360), sut.Right)
	assert.Equal(t, uint(116), sut.Length)
}

func TestShiftLeft65Bits(t *testing.T) {
	sut := NewIPv6Address([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xB0}, 117)

	sut.ShiftLeft(uint(65))
	assert.Equal(t, uint64(0X47494B4D4F515360), sut.Left)
	assert.Equal(t, uint(52), sut.Length)
}

func TestShiftLeftIPv6(t *testing.T) {
	var newLeft uint64
	var newRight uint64
	var newLength uint

	// shift < 64 bits
	newLeft, newRight, newLength = ShiftLeftIPv6(uint64(0x0102030405060701), uint64(0x7A33881100338844), 117, 50)
	assert.Equal(t, uint64(0x1C05E8CE204400CE), newLeft)
	assert.Equal(t, uint64(0x2110000000000000), newRight)
	assert.Equal(t, uint(67), newLength)

	// shift 64 bits
	newLeft, newRight, newLength = ShiftLeftIPv6(uint64(0x0102030405060701), uint64(0x7A33881100338844), 117, 64)
	assert.Equal(t, uint64(0x7A33881100338844), newLeft)
	assert.Equal(t, uint64(0x0), newRight)
	assert.Equal(t, uint(53), newLength)

	// shift > 64 bits
	newLeft, newRight, newLength = ShiftLeftIPv6(uint64(0x0102030405060701), uint64(0x7A33881100338844), 117, 100)
	assert.Equal(t, uint64(0x338844000000000), newLeft)
	assert.Equal(t, uint64(0x0), newRight)
	assert.Equal(t, uint(17), newLength)
}

func TestShiftRightIPv6(t *testing.T) {
	var newLeft uint64
	var newRight uint64

	// shift < 64
	newLeft, newRight = ShiftRightIPv6(uint64(0x0102030405060701), uint64(0x7A33881100338844), 50)
	assert.Equal(t, uint64(0x40), newLeft)
	assert.Equal(t, uint64(0x80C1014181C05E8C), newRight)

	// shift 64
	newLeft, newRight = ShiftRightIPv6(uint64(0x0102030405060701), uint64(0x7A33881100338844), 64)
	assert.Equal(t, uint64(0x0), newLeft)
	assert.Equal(t, uint64(0x0102030405060701), newRight)

	// shift > 64
	newLeft, newRight = ShiftRightIPv6(uint64(0x0102030405060701), uint64(0x7A33881100338844), 77)
	assert.Equal(t, uint64(0x0), newLeft)
	assert.Equal(t, uint64(0x81018202830), newRight)
}
