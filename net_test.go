package patricia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIPFromString(t *testing.T) {
	var v4IP *IPv4Address
	var v6IP *IPv6Address
	var err error

	v4IP, v6IP, err = ParseIPFromString("sdfkjsdfdsfd")
	assert.Error(t, err)

	v4IP, v6IP, err = ParseIPFromString("")
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

	v4IP, v6IP, err = ParseIPFromString("127.0.0.1/999")
	assert.Error(t, err)

	v4IP, v6IP, err = ParseIPFromString("127.0.0.1/33")
	assert.Error(t, err)

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

	v4IP, v6IP, err = ParseIPFromString("2001:0db8:85a3:0000:0000:8a2e:0370:7334/129")
	assert.Error(t, err)

	v4IP, v6IP, err = ParseIPFromString("2001:0db8:85a3:0000:0000:8a2e:0370:7334/16")
	assert.NoError(t, err)
	assert.Nil(t, v4IP)
	assert.NotNil(t, v6IP)
	assert.Equal(t, uint(16), v6IP.Length)
	assert.Equal(t, uint64(0x2001000000000000), v6IP.Left)
	assert.Equal(t, uint64(0x0), v6IP.Right)
}
