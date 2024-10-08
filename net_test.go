package patricia

import (
	"net"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIPFromString(t *testing.T) {
	var v4IP *IPv4Address
	var v6IP *IPv6Address
	var err error

	_, _, err = ParseIPFromString("sdfkjsdfdsfd")
	assert.Error(t, err)

	_, _, err = ParseIPFromString("")
	assert.Error(t, err)

	v4IP, v6IP, err = ParseIPFromString("0.0.0.0/0")
	assert.NoError(t, err)
	assert.NotNil(t, v4IP)
	assert.Equal(t, uint(0), v4IP.Length)
	assert.Nil(t, v6IP)

	v4IP, v6IP, err = ParseIPFromString("127.0.0.1")
	assert.NoError(t, err)
	assert.Equal(t, uint(32), v4IP.Length)
	assert.NotNil(t, v4IP)
	assert.Nil(t, v6IP)

	_, _, err = ParseIPFromString("127.0.0.1/999")
	assert.Error(t, err)

	_, _, err = ParseIPFromString("127.0.0.1/33")
	assert.Error(t, err)

	v4IP, v6IP, err = ParseIPFromString("1.0.16.0/0")
	assert.NoError(t, err)
	assert.NotNil(t, v4IP)
	assert.Equal(t, uint(0), v4IP.Length)
	assert.Nil(t, v6IP)

	v4IP, v6IP, err = ParseIPFromString("127.0.0.1/10")
	assert.NoError(t, err)
	assert.NotNil(t, v4IP)
	assert.Equal(t, uint(10), v4IP.Length)
	assert.Nil(t, v6IP)

	v4IP, v6IP, err = ParseIPFromString("127.0.0.1/32")
	assert.NoError(t, err)
	assert.NotNil(t, v4IP)
	assert.Equal(t, uint(32), v4IP.Length)
	assert.Nil(t, v6IP)

	v4IP, v6IP, err = ParseIPFromString("2001:0db8:85a3:0000:0000:8a2e:0370:7334/128")
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(128), v6IP.Length)
	assert.Equal(t, uint64(0x20010db885a30000), v6IP.Left)
	assert.Equal(t, uint64(0x00008a2e03707334), v6IP.Right)

	v4IP, v6IP, err = ParseIPFromString("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(128), v6IP.Length)
	assert.Equal(t, uint64(0x20010db885a30000), v6IP.Left)
	assert.Equal(t, uint64(0x00008a2e03707334), v6IP.Right)

	_, _, err = ParseIPFromString("2001:0db8:85a3:0000:0000:8a2e:0370:7334/129")
	assert.Error(t, err)

	v4IP, v6IP, err = ParseIPFromString("2001:0db8:85a3:0000:0000:8a2e:0370:7334/16")
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(16), v6IP.Length)
	assert.Equal(t, uint64(0x2001000000000000), v6IP.Left)
	assert.Equal(t, uint64(0x0), v6IP.Right)

	ip := net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	assert.NotNil(t, ip)
	v4IP, v6IP, err = ParseFromIP(&ip)
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(128), v6IP.Length)
	assert.Equal(t, uint64(0x20010db885a30000), v6IP.Left)
	assert.Equal(t, uint64(0x00008a2e03707334), v6IP.Right)

	ip = net.ParseIP("127.0.0.1")
	assert.NotNil(t, ip)
	v4IP, v6IP, err = ParseFromIP(&ip)
	assert.NoError(t, err)
	assert.NotNil(t, v4IP)
	assert.Equal(t, uint(32), v4IP.Length)
	assert.Equal(t, uint32(0x7f000001), v4IP.Address)
	assert.Nil(t, v6IP)

	_, ipr, _ := net.ParseCIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334/16")
	v4IP, v6IP, err = ParseFromIPAddr(ipr)
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(16), v6IP.Length)
	assert.Equal(t, uint64(0x2001000000000000), v6IP.Left)
	assert.Equal(t, uint64(0x0), v6IP.Right)

	_, ipr, _ = net.ParseCIDR("127.0.0.1/10")
	v4IP, v6IP, err = ParseFromIPAddr(ipr)
	assert.NoError(t, err)
	assert.NotNil(t, v4IP)
	assert.Equal(t, uint(10), v4IP.Length)
	assert.Nil(t, v6IP)

	v4IP, v6IP, err = ParseIPFromString("::ffff:10.10.10.10/128")
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(128), v6IP.Length)
	assert.Equal(t, uint64(0x0000000000000000), v6IP.Left)
	assert.Equal(t, uint64(0x0000ffff0a0a0a0a), v6IP.Right)

	v4IP, v6IP, err = ParseIPFromString("::ffff:10.10.10.10/120")
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(120), v6IP.Length)
	assert.Equal(t, uint64(0x0000000000000000), v6IP.Left)
	assert.Equal(t, uint64(0x0000ffff0a0a0a00), v6IP.Right)

	v4IP, v6IP, err = ParseIPFromString("::ffff:10.10.10.10/118")
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(118), v6IP.Length)
	assert.Equal(t, uint64(0x0000000000000000), v6IP.Left)
	assert.Equal(t, uint64(0x0000ffff0a0a0800), v6IP.Right)

	v4IP, v6IP, err = ParseIPFromString("::ffff:10.10.10.10")
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(128), v6IP.Length)
	assert.Equal(t, uint64(0x0000000000000000), v6IP.Left)
	assert.Equal(t, uint64(0x0000ffff0a0a0a0a), v6IP.Right)

}

func TestParseIPFromNetIP(t *testing.T) {
	addr := netip.MustParseAddr("::ffff:10.10.10.10")
	v4IP, v6IP, err := ParseFromNetIPAddr(addr)
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(128), v6IP.Length)
	assert.Equal(t, uint64(0x0000000000000000), v6IP.Left)
	assert.Equal(t, uint64(0x0000ffff0a0a0a0a), v6IP.Right)

	addr = netip.MustParseAddr("127.0.0.1")
	assert.NotNil(t, addr)
	v4IP, v6IP, err = ParseFromNetIPAddr(addr)
	assert.NoError(t, err)
	assert.NotNil(t, v4IP)
	assert.Equal(t, uint(32), v4IP.Length)
	assert.Equal(t, uint32(0x7f000001), v4IP.Address)
	assert.Nil(t, v6IP)

	prefix := netip.MustParsePrefix("127.0.0.1/10")
	v4IP, v6IP, err = ParseFromNetIPPrefix(prefix)
	assert.NoError(t, err)
	assert.NotNil(t, v4IP)
	assert.Equal(t, uint(10), v4IP.Length)
	assert.Nil(t, v6IP)

	prefix = netip.MustParsePrefix("::ffff:10.10.10.10/128")
	v4IP, v6IP, err = ParseFromNetIPPrefix(prefix)
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(128), v6IP.Length)
	assert.Equal(t, uint64(0x0000000000000000), v6IP.Left)
	assert.Equal(t, uint64(0x0000ffff0a0a0a0a), v6IP.Right)

	_, _, err = ParseFromNetIPAddr(netip.Addr{})
	assert.Error(t, err)

	_, _, err = ParseFromNetIPPrefix(netip.Prefix{})
	assert.Error(t, err)

	_, _, err = ParseFromNetIPAddr(netip.MustParseAddr("0.0.0.0"))
	assert.Error(t, err)

}
