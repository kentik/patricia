package patricia

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIPv4Address(t *testing.T) {
	// 4 byte format
	sut := NewIPv4Address(uint32(0x01234567), 7)

	assert.Equal(t, uint32(0x01234567), sut.Address)
	assert.Equal(t, uint(7), sut.Length)
	assert.Equal(t, "1.35.69.103/7", sut.String())
	assert.Equal(t, "1.35.69.103/7", (&sut).String())
	//nolint
	assert.Equal(t, "1.35.69.103/7", fmt.Sprintf("%s", sut))
	//nolint
	assert.Equal(t, "1.35.69.103/7", fmt.Sprintf("%s", &sut))
}

func TestNewIPv4AddressFromBytes(t *testing.T) {
	// 4-byte format - from bytes
	sut := NewIPv4AddressFromBytes([]byte{3, 4, 5, 6}, 31)
	assert.Equal(t, uint32(0x03040506), sut.Address)
	assert.Equal(t, uint(31), sut.Length)

	// 16-byte format - from net.IPV4
	sut = NewIPv4AddressFromBytes(net.IPv4(3, 4, 5, 6), 31)
	assert.Equal(t, uint32(0x03040506), sut.Address)
	assert.Equal(t, uint(31), sut.Length)

	// 16-byte format - directly from bytes
	sut = NewIPv4AddressFromBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 3, 4, 5, 6}, 31)
	assert.Equal(t, uint32(0x03040506), sut.Address)
	assert.Equal(t, uint(31), sut.Length)

	// 3-byte invalid format - directly from bytes
	sut = NewIPv4AddressFromBytes([]byte{4, 5, 6}, 31)
	assert.Equal(t, uint32(0), sut.Address)
	assert.Equal(t, uint(0), sut.Length)

	// 17-byte invalid format - directly from bytes
	sut = NewIPv4AddressFromBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 3, 4, 5, 6}, 31)
	assert.Equal(t, uint32(0), sut.Address)
	assert.Equal(t, uint(0), sut.Length)
}
