package patricia

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIPv4Address(t *testing.T) {
	sut := NewIPv4Address(uint32(0x01234567), 7)

	assert.Equal(t, uint32(0x01234567), sut.Address)
	assert.Equal(t, uint(7), sut.Length)
	assert.Equal(t, "1.35.69.103/7", sut.String())
	assert.Equal(t, "1.35.69.103/7", (&sut).String())
	assert.Equal(t, "1.35.69.103/7", fmt.Sprintf("%s", sut))
	assert.Equal(t, "1.35.69.103/7", fmt.Sprintf("%s", &sut))
}
