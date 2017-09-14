package patricia

import (
	"encoding/binary"
)

// IPv4Address is a representation of an IPv4 address and CIDR
type IPv4Address struct {
	Address uint32
	Length  uint
}

// NewIPv4Address creates an address from the input IPv4 address
func NewIPv4Address(address uint32, length uint) IPv4Address {
	return IPv4Address{
		Address: address,
		Length:  length,
	}
}

// NewIPv4AddressFromBytes creates an address from the input IPv4 address bytes
// - address must be 4 bytes
func NewIPv4AddressFromBytes(address []byte, length uint) IPv4Address {
	return IPv4Address{
		Address: binary.BigEndian.Uint32(address),
		Length:  length,
	}
}
