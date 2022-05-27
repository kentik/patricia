package patricia

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ParseIPFromString parses a string address, returning a v4 or v6 IP address
// TODO: make this more performant:
//       - is the fmt.Sprintf necessary?
func ParseIPFromString(address string) (*IPv4Address, *IPv6Address, error) {
	var err error

	// see if there's a CIDR
	parts := strings.Split(address, "/")
	cidr := -1 // default needs to be -1 to handle /0
	if len(parts) == 2 {
		c, err := strconv.ParseUint(parts[1], 10, 8)
		if err != nil {
			return nil, nil, fmt.Errorf("couldn't parse CIDR to int: %s", err)
		}
		if c > 128 {
			return nil, nil, fmt.Errorf("Invalid CIDR: %d", c)
		}
		cidr = int(c)
	}

	// try parsing as IPv4 - force CIDR at the end
	v4AddrStr := address
	if cidr == -1 {
		// no CIDR specified - tack on /32
		v4AddrStr = fmt.Sprintf("%s/32", address)
	}
	_, ipNet, err := net.ParseCIDR(v4AddrStr)
	if err == nil {
		cidr, mask := ipNet.Mask.Size()
		if v4Addr := ipNet.IP.To4(); v4Addr != nil && mask == 32 { // nil error here
			ret := NewIPv4AddressFromBytes(v4Addr, uint(cidr))
			return &ret, nil, nil
		}
	}

	// try parsing as IPv6
	v6AddrStr := address
	if cidr == -1 {
		// no CIDR specified - tack on /128
		v6AddrStr = fmt.Sprintf("%s/128", address)
	}
	_, ipNet, err = net.ParseCIDR(v6AddrStr)
	if err == nil {
		cidr, mask := ipNet.Mask.Size()
		if v6Addr := ipNet.IP.To16(); v6Addr != nil && mask == 128 {
			ret := NewIPv6Address(v6Addr, uint(cidr))
			return nil, &ret, nil
		}
	}

	return nil, nil, fmt.Errorf("couldn't parse either v4 or v6 address")
}

// ParseFromIP builds an IPv4Address or IPv6Address from a net.IP
func ParseFromIP(ip *net.IP) (*IPv4Address, *IPv6Address, error) {
	if ip == nil {
		return nil, nil, fmt.Errorf("Nil address: %v", ip)
	}

	if v4Addr := ip.To4(); v4Addr != nil {
		ret := NewIPv4AddressFromBytes(v4Addr, 32)
		return &ret, nil, nil
	}
	if v6Addr := ip.To16(); v6Addr != nil {
		ret := NewIPv6Address(v6Addr, 128)
		return nil, &ret, nil
	}

	return nil, nil, fmt.Errorf("couldn't parse either v4 or v6 address: %v", ip)
}

// ParseFromIPAddr builds an IPv4Address or IPv6Address from a net.IPNet
func ParseFromIPAddr(ipNet *net.IPNet) (*IPv4Address, *IPv6Address, error) {
	if ipNet == nil {
		return nil, nil, fmt.Errorf("Nil address: %v", ipNet)
	}

	if v4Addr := ipNet.IP.To4(); v4Addr != nil {
		cidr, _ := ipNet.Mask.Size()
		ret := NewIPv4AddressFromBytes(v4Addr, uint(cidr))
		return &ret, nil, nil
	}
	if v6Addr := ipNet.IP.To16(); v6Addr != nil {
		cidr, _ := ipNet.Mask.Size()
		ret := NewIPv6Address(v6Addr, uint(cidr))
		return nil, &ret, nil
	}

	return nil, nil, fmt.Errorf("couldn't parse either v4 or v6 address: %v", ipNet)
}
