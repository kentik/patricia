package patricia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIPv4Address(t *testing.T) {
	sut := NewIPv4Address(uint32(0x01234567), 7)

	assert.Equal(t, uint32(0x01234567), sut.Address)
	assert.Equal(t, uint(7), sut.Length)
}
