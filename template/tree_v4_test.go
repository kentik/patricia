package template

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	"github.com/kentik/patricia"
	"github.com/stretchr/testify/assert"
)

func ipv4FromBytes(bytes []byte, length int) patricia.IPv4Address {
	return patricia.IPv4Address{
		Address: binary.BigEndian.Uint32(bytes),
		Length:  uint(length),
	}
}

func BenchmarkFindTags(b *testing.B) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	tree := NewTreeV4()

	tree.Add(patricia.IPv4Address{}, tagZ, nil) // default
	tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), tagA, nil)
	tree.Add(ipv4FromBytes([]byte{160, 0, 0, 0}, 2), tagB, nil) // 160 -> 128
	tree.Add(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), tagC, nil)

	address := patricia.NewIPv4Address(uint32(2156823809), 32)

	var buf []GeneratedType
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		buf = tree.FindTags(address)
	}
	_ = buf
}

func BenchmarkFindTagsAppend(b *testing.B) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	tree := NewTreeV4()

	buf := make([]GeneratedType, 0)
	tree.Add(patricia.IPv4Address{}, tagZ, nil) // default
	tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), tagA, nil)
	tree.Add(ipv4FromBytes([]byte{160, 0, 0, 0}, 2), tagB, nil) // 160 -> 128
	tree.Add(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), tagC, nil)

	address := patricia.NewIPv4Address(uint32(2156823809), 32)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		buf = buf[:0]
		buf = tree.FindTagsAppend(buf, address)
	}
}

func BenchmarkFindDeepestTag(b *testing.B) {
	tree := NewTreeV4()
	for i := 32; i > 0; i-- {
		tree.Add(ipv4FromBytes([]byte{127, 0, 0, 1}, i), fmt.Sprintf("Tag-%d", i), nil)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		address := patricia.NewIPv4Address(uint32(2130706433), 32)
		tree.FindDeepestTag(address)
	}
}

func BenchmarkBuildTreeAndFindDeepestTag(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree := NewTreeV4()

		// populate

		address := patricia.NewIPv4Address(uint32(1653323544), 32)
		tree.Add(address, "tagA", nil)

		address = patricia.NewIPv4Address(uint32(3334127283), 32)
		tree.Add(address, "tagB", nil)

		address = patricia.NewIPv4Address(uint32(2540010580), 32)
		tree.Add(address, "tagC", nil)

		// search

		address = patricia.NewIPv4Address(uint32(1653323544), 32)
		tree.FindDeepestTag(address)

		address = patricia.NewIPv4Address(uint32(3334127283), 32)
		tree.FindDeepestTag(address)

		address = patricia.NewIPv4Address(uint32(2540010580), 32)
		tree.FindDeepestTag(address)
	}
}

func TestTree2(t *testing.T) {
	tags := make([]GeneratedType, 0)
	found := false

	tree := NewTreeV4()
	// insert a bunch of tags
	v4, _, err := patricia.ParseIPFromString("1.2.3.0/24")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "foo", nil)
	tree.Add(*v4, "bar", nil)

	v4, _, err = patricia.ParseIPFromString("188.212.216.242")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "a", nil)

	v4, _, err = patricia.ParseIPFromString("171.233.143.228")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "b", nil)

	v4, _, err = patricia.ParseIPFromString("186.244.183.12")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "c", nil)

	v4, _, err = patricia.ParseIPFromString("171.233.143.222")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "d", nil)

	v4, _, err = patricia.ParseIPFromString("190.207.189.24")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "e", nil)

	v4, _, err = patricia.ParseIPFromString("188.212.216.240")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "f", nil)

	v4, _, err = patricia.ParseIPFromString("185.76.10.148")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "g", nil)

	v4, _, err = patricia.ParseIPFromString("14.208.248.50")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "h", nil)

	v4, _, err = patricia.ParseIPFromString("59.60.75.52")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "i", nil)

	v4, _, err = patricia.ParseIPFromString("185.76.10.146")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "j", nil)
	tree.Add(*v4, "k", nil)

	// --------
	// now assert they're all found
	v4, _, _ = patricia.ParseIPFromString("188.212.216.242")
	found, tag := tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "a", tag)

	v4, _, _ = patricia.ParseIPFromString("171.233.143.228")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "b", tag)

	v4, _, _ = patricia.ParseIPFromString("186.244.183.12")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "c", tag)

	v4, _, _ = patricia.ParseIPFromString("171.233.143.222")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "d", tag)

	v4, _, _ = patricia.ParseIPFromString("190.207.189.24")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "e", tag)

	v4, _, _ = patricia.ParseIPFromString("188.212.216.240")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "f", tag)

	v4, _, _ = patricia.ParseIPFromString("185.76.10.148")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "g", tag)

	v4, _, _ = patricia.ParseIPFromString("14.208.248.50")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "h", tag)

	v4, _, _ = patricia.ParseIPFromString("59.60.75.52")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "i", tag)

	v4, _, _ = patricia.ParseIPFromString("185.76.10.146")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "j", tag)

	tags = tags[:0]
	v4, _, _ = patricia.ParseIPFromString("185.76.10.146")
	found, tags = tree.FindDeepestTagsAppend(tags, *v4)
	assert.True(t, found)
	assert.Equal(t, "j", tags[0])
	assert.Equal(t, "k", tags[1])

	// test searching for addresses with no leaf nodes
	tags = tags[:0]
	v4, _, _ = patricia.ParseIPFromString("1.2.3.4")
	found, tags = tree.FindDeepestTagsAppend(tags, *v4)
	assert.True(t, found)
	assert.Equal(t, "foo", tags[0])
	assert.Equal(t, "bar", tags[1])

	tags = tags[:0]
	v4, _, _ = patricia.ParseIPFromString("1.2.3.5")
	found, tags = tree.FindDeepestTagsAppend(tags, *v4)
	assert.True(t, found)
	assert.Equal(t, "foo", tags[0])
	assert.Equal(t, "bar", tags[1])

	v4, _, _ = patricia.ParseIPFromString("1.2.3.4")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "foo", tag)

	// test searching for an address that has nothing
	tags = tags[:0]
	v4, _, _ = patricia.ParseIPFromString("9.9.9.9")
	found, tags = tree.FindDeepestTagsAppend(tags, *v4)
	assert.False(t, found)
	assert.NotNil(t, tags)
	assert.Equal(t, 0, len(tags))

	// test searching for an empty address
	tags = tags[:0]
	v4, _, _ = patricia.ParseIPFromString("9.9.9.9/0")
	found, tags = tree.FindDeepestTagsAppend(tags, *v4)
	assert.False(t, found)
	assert.NotNil(t, tags)
	assert.Equal(t, 0, len(tags))

	// add a root node tag and try again
	v4, _, err = patricia.ParseIPFromString("1.1.1.1/0")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	tree.Add(*v4, "root_node", nil)

	v4, _, _ = patricia.ParseIPFromString("9.9.9.9/0")
	tags = tags[:0]
	found, tags = tree.FindDeepestTagsAppend(tags, *v4)
	assert.True(t, found)
	assert.NotNil(t, tags)
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "root_node", tags[0])

	v4, _, _ = patricia.ParseIPFromString("9.9.9.9/0")
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "root_node", tag)
}

// TestFindDeepestTags tests self-allocating FindDeepestTags
func TestFindDeepestTags(t *testing.T) {
	assert := assert.New(t)

	var tags []GeneratedType
	found := false

	tree := NewTreeV4()
	// insert a bunch of tags
	v4, _, err := patricia.ParseIPFromString("1.2.3.0/24")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "foo", nil)
	tree.Add(*v4, "bar", nil)

	v4, _, err = patricia.ParseIPFromString("188.212.216.242")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "a", nil)

	v4, _, err = patricia.ParseIPFromString("171.233.143.228")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "b", nil)

	v4, _, err = patricia.ParseIPFromString("186.244.183.12")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "c", nil)

	v4, _, err = patricia.ParseIPFromString("171.233.143.222")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "d", nil)

	v4, _, err = patricia.ParseIPFromString("190.207.189.24")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "e", nil)

	v4, _, err = patricia.ParseIPFromString("188.212.216.240")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "f", nil)

	v4, _, err = patricia.ParseIPFromString("185.76.10.148")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "g", nil)

	v4, _, err = patricia.ParseIPFromString("14.208.248.50")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "h", nil)

	v4, _, err = patricia.ParseIPFromString("59.60.75.52")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "i", nil)

	v4, _, err = patricia.ParseIPFromString("185.76.10.146")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "j", nil)
	tree.Add(*v4, "k", nil)

	// now test
	v4, _, _ = patricia.ParseIPFromString("185.76.10.146")
	found, tags = tree.FindDeepestTags(*v4)
	assert.True(found)
	assert.Equal("j", tags[0])
	assert.Equal("k", tags[1])
	assert.Equal(2, len(tags))

	// test searching for addresses with no leaf nodes
	v4, _, _ = patricia.ParseIPFromString("1.2.3.4")
	found, tags = tree.FindDeepestTags(*v4)
	assert.True(found)
	assert.Equal("foo", tags[0])
	assert.Equal("bar", tags[1])
	assert.Equal(2, len(tags))

	v4, _, _ = patricia.ParseIPFromString("1.2.3.5")
	found, tags = tree.FindDeepestTags(*v4)
	assert.True(found)
	assert.Equal("foo", tags[0])
	assert.Equal("bar", tags[1])
	assert.Equal(2, len(tags))

	v4, _, _ = patricia.ParseIPFromString("1.2.3.5")
	found, tags = tree.FindDeepestTagsWithFilter(*v4, func(t GeneratedType) bool { return t.(string) == "foo" })
	assert.True(found)
	assert.Equal("foo", tags[0])
	assert.Equal(1, len(tags))

	v4, _, _ = patricia.ParseIPFromString("1.2.3.5")
	found, tags = tree.FindDeepestTagsWithFilter(*v4, func(t GeneratedType) bool { return t.(string) == "nothing" })
	assert.True(found)
	assert.Equal(0, len(tags))

	// test searching for an address that has nothing
	v4, _, _ = patricia.ParseIPFromString("9.9.9.9")
	found, tags = tree.FindDeepestTags(*v4)
	assert.False(found)
	assert.NotNil(tags)
	assert.Equal(0, len(tags))

	// test searching for an empty address
	v4, _, _ = patricia.ParseIPFromString("9.9.9.9/0")
	found, tags = tree.FindDeepestTags(*v4)
	assert.False(found)
	assert.NotNil(tags)
	assert.Equal(0, len(tags))

	// add a root node tag and try again
	v4, _, err = patricia.ParseIPFromString("1.1.1.1/0")
	assert.NoError(err)
	assert.NotNil(v4)
	tree.Add(*v4, "root_node", nil)

	v4, _, _ = patricia.ParseIPFromString("9.9.9.9/0")
	found, tags = tree.FindDeepestTags(*v4)
	assert.True(found)
	assert.NotNil(tags)
	assert.Equal(1, len(tags))
	assert.Equal("root_node", tags[0])
}

// test that the find functions don't destroy an address - too brittle and confusing for caller for what gains?
func TestAddressReusable(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tree := NewTreeV4()

	pv4, pv6, err := patricia.ParseIPFromString("59.60.75.53") // needs to share same second-level node with the address we're going to work with
	tree.Add(*pv4, "Don't panic!", nil)
	assert.NoError(t, err)
	assert.NotNil(t, pv4)
	assert.Nil(t, pv6)

	v4, v6, err := patricia.ParseIPFromString("59.60.75.52")
	assert.NoError(t, err)
	assert.NotNil(t, v4)
	assert.Nil(t, v6)

	tree.Add(*v4, "Hello", nil)
	found, tag := tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "Hello", tag)

	// search again with same address
	found, tag = tree.FindDeepestTag(*v4)
	assert.True(t, found)
	assert.Equal(t, "Hello", tag)

	// search again with same address
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, *v4)
	if assert.Equal(t, 1, len(tags)) {
		assert.Equal(t, "Hello", tags[0])
	}

	// search again with same address
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, *v4)
	if assert.Equal(t, 1, len(tags)) {
		assert.Equal(t, "Hello", tags[0])
	}
}

func TestSimpleTree1(t *testing.T) {
	tree := NewTreeV4()

	ipv4a := ipv4FromBytes([]byte{98, 139, 183, 24}, 32)
	ipv4b := ipv4FromBytes([]byte{198, 186, 190, 179}, 32)
	ipv4c := ipv4FromBytes([]byte{151, 101, 124, 84}, 32)

	tree.Add(ipv4a, "tagA", nil)
	tree.Add(ipv4b, "tagB", nil)
	tree.Add(ipv4c, "tagC", nil)

	found, tag := tree.FindDeepestTag(ipv4FromBytes([]byte{98, 139, 183, 24}, 32))
	assert.True(t, found)
	assert.Equal(t, "tagA", tag)

	found, tag = tree.FindDeepestTag(ipv4FromBytes([]byte{198, 186, 190, 179}, 32))
	assert.True(t, found)
	assert.Equal(t, "tagB", tag)

	found, tag = tree.FindDeepestTag(ipv4FromBytes([]byte{151, 101, 124, 84}, 32))
	assert.True(t, found)
	assert.Equal(t, "tagC", tag)
}

// TestSimpleTree1Append tests that FindTagsAppend appends
func TestSimpleTree1Append(t *testing.T) {
	tree := NewTreeV4()

	ipv4a := ipv4FromBytes([]byte{98, 139, 183, 24}, 32)
	ipv4b := ipv4FromBytes([]byte{198, 186, 190, 179}, 32)
	ipv4c := ipv4FromBytes([]byte{151, 101, 124, 84}, 32)

	tree.Add(ipv4a, "tagA", nil)
	tree.Add(ipv4b, "tagB", nil)
	tree.Add(ipv4c, "tagC", nil)

	tags := make([]GeneratedType, 0)
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{98, 139, 183, 24}, 32))
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "tagA", tags[0])

	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{198, 186, 190, 179}, 32))
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, "tagA", tags[0])
	assert.Equal(t, "tagB", tags[1])

	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{151, 101, 124, 84}, 32))
	assert.Equal(t, 3, len(tags))
	assert.Equal(t, "tagA", tags[0])
	assert.Equal(t, "tagB", tags[1])
	assert.Equal(t, "tagC", tags[2])
}

// TestSimpleTree1Append tests the self-allocating FindTags
func TestSimpleTree1FindTags(t *testing.T) {
	tree := NewTreeV4()

	ipv4a := ipv4FromBytes([]byte{98, 139, 183, 24}, 32)
	ipv4b := ipv4FromBytes([]byte{198, 186, 190, 179}, 32)
	ipv4c := ipv4FromBytes([]byte{151, 101, 124, 84}, 32)

	tree.Add(ipv4a, "tagA", nil)
	tree.Add(ipv4b, "tagB", nil)
	tree.Add(ipv4c, "tagC", nil)

	tags := tree.FindTags(ipv4FromBytes([]byte{98, 139, 183, 24}, 32))
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "tagA", tags[0])

	tags = tree.FindTags(ipv4FromBytes([]byte{198, 186, 190, 179}, 32))
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "tagB", tags[0])

	tags = tree.FindTags(ipv4FromBytes([]byte{151, 101, 124, 84}, 32))
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "tagC", tags[0])
}

// TestSimpleTree1FilterAppend tests that FindTagsWithFilterAppend appends
func TestSimpleTree1FilterAppend(t *testing.T) {
	assert := assert.New(t)

	include := true
	filterFunc := func(val GeneratedType) bool {
		return include
	}

	tree := NewTreeV4()

	ipv4a := ipv4FromBytes([]byte{98, 139, 183, 24}, 32)
	ipv4b := ipv4FromBytes([]byte{198, 186, 190, 179}, 32)
	ipv4c := ipv4FromBytes([]byte{151, 101, 124, 84}, 32)

	tree.Add(ipv4a, "tagA", nil)
	tree.Add(ipv4b, "tagB", nil)
	tree.Add(ipv4c, "tagC", nil)

	include = false
	tags := make([]GeneratedType, 0)
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{98, 139, 183, 24}, 32), filterFunc)
	assert.Equal(0, len(tags))

	include = true
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{98, 139, 183, 24}, 32), filterFunc)
	assert.Equal(1, len(tags))
	assert.Equal("tagA", tags[0])

	include = false
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{198, 186, 190, 179}, 32), filterFunc)
	assert.Equal(1, len(tags))

	include = true
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{198, 186, 190, 179}, 32), filterFunc)
	assert.Equal(2, len(tags))
	assert.Equal("tagA", tags[0])
	assert.Equal("tagB", tags[1])

	include = false
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{151, 101, 124, 84}, 32), filterFunc)
	assert.Equal(2, len(tags))

	include = true
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{151, 101, 124, 84}, 32), filterFunc)
	assert.Equal(3, len(tags))
	assert.Equal("tagA", tags[0])
	assert.Equal("tagB", tags[1])
	assert.Equal("tagC", tags[2])

	include = false
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{151, 101, 124, 84}, 32), filterFunc)
	assert.Equal(3, len(tags))

	include = true
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{151, 101, 124, 84}, 32), filterFunc)
	assert.Equal(4, len(tags))
	assert.Equal("tagA", tags[0])
	assert.Equal("tagB", tags[1])
	assert.Equal("tagC", tags[2])
	assert.Equal("tagC", tags[2])
}

// TestSimpleTree1Filter tests that FindTagsWithFilter
func TestSimpleTree1Filter(t *testing.T) {
	assert := assert.New(t)

	include := true
	filterFunc := func(val GeneratedType) bool {
		return include
	}

	tree := NewTreeV4()

	ipv4a := ipv4FromBytes([]byte{98, 139, 183, 24}, 32)
	ipv4b := ipv4FromBytes([]byte{198, 186, 190, 179}, 32)
	ipv4c := ipv4FromBytes([]byte{151, 101, 124, 84}, 32)

	tree.Add(ipv4a, "tagA", nil)
	tree.Add(ipv4b, "tagB", nil)
	tree.Add(ipv4c, "tagC", nil)

	include = false
	tags := tree.FindTagsWithFilter(ipv4FromBytes([]byte{98, 139, 183, 24}, 32), filterFunc)
	assert.Equal(0, len(tags))

	include = true
	tags = tree.FindTagsWithFilter(ipv4FromBytes([]byte{98, 139, 183, 24}, 32), filterFunc)
	assert.Equal(1, len(tags))
	assert.Equal("tagA", tags[0])

	include = false
	tags = tree.FindTagsWithFilter(ipv4FromBytes([]byte{198, 186, 190, 179}, 32), filterFunc)
	assert.Equal(0, len(tags))

	include = true
	tags = tree.FindTagsWithFilter(ipv4FromBytes([]byte{198, 186, 190, 179}, 32), filterFunc)
	assert.Equal(1, len(tags))
	assert.Equal("tagB", tags[0])

	include = false
	tags = tree.FindTagsWithFilter(ipv4FromBytes([]byte{151, 101, 124, 84}, 32), filterFunc)
	assert.Equal(0, len(tags))

	include = true
	tags = tree.FindTagsWithFilter(ipv4FromBytes([]byte{151, 101, 124, 84}, 32), filterFunc)
	assert.Equal(1, len(tags))
	assert.Equal("tagC", tags[0])

	include = false
	tags = tree.FindTagsWithFilter(ipv4FromBytes([]byte{151, 101, 124, 84}, 32), filterFunc)
	assert.Equal(0, len(tags))

	include = true
	tags = tree.FindTagsWithFilter(ipv4FromBytes([]byte{151, 101, 124, 84}, 32), filterFunc)
	assert.Equal(1, len(tags))
	assert.Equal("tagC", tags[0])
}

// Test having a couple of inner nodes
func TestSimpleTree2(t *testing.T) {
	buf := make([]GeneratedType, 0)

	ipA, _, _ := patricia.ParseIPFromString("203.143.220.0/23")
	ipB, _, _ := patricia.ParseIPFromString("203.143.220.198/32")
	ipC, _, _ := patricia.ParseIPFromString("203.143.0.0/16")
	ipD, _, _ := patricia.ParseIPFromString("203.143.221.75/32")

	// add the 4 addresses
	tree := NewTreeV4()
	tree.Add(*ipA, "A", nil)
	tree.Add(*ipB, "B", nil)
	tree.Add(*ipC, "C", nil)
	tree.Add(*ipD, "D", nil)

	// find the 4 addresses
	found, _ := tree.FindDeepestTag(*ipA)
	assert.True(t, found)
	found, _ = tree.FindDeepestTag(*ipB)
	assert.True(t, found)
	found, _ = tree.FindDeepestTag(*ipC)
	assert.True(t, found)
	found, _ = tree.FindDeepestTag(*ipD)
	assert.True(t, found)

	// delete each one
	matchFunc := func(a GeneratedType, b GeneratedType) bool {
		return a == b
	}
	deleteCount := tree.DeleteWithBuffer(buf, *ipA, matchFunc, "A")
	assert.Equal(t, 1, deleteCount)
	deleteCount = tree.DeleteWithBuffer(buf, *ipB, matchFunc, "B")
	assert.Equal(t, 1, deleteCount)
	deleteCount = tree.DeleteWithBuffer(buf, *ipC, matchFunc, "C")
	assert.Equal(t, 1, deleteCount)
	deleteCount = tree.DeleteWithBuffer(buf, *ipD, matchFunc, "D")
	assert.Equal(t, 1, deleteCount)

	// should have zero logical nodes except for root
	assert.Equal(t, 1, tree.countNodes(1))
}

// Test having a couple of inner nodes - with self-allocating Delete method
func TestSimpleTree2WithDelete(t *testing.T) {
	ipA, _, _ := patricia.ParseIPFromString("203.143.220.0/23")
	ipB, _, _ := patricia.ParseIPFromString("203.143.220.198/32")
	ipC, _, _ := patricia.ParseIPFromString("203.143.0.0/16")
	ipD, _, _ := patricia.ParseIPFromString("203.143.221.75/32")

	// add the 4 addresses
	tree := NewTreeV4()
	tree.Add(*ipA, "A", nil)
	tree.Add(*ipB, "B", nil)
	tree.Add(*ipC, "C", nil)
	tree.Add(*ipD, "D", nil)

	// find the 4 addresses
	found, _ := tree.FindDeepestTag(*ipA)
	assert.True(t, found)
	found, _ = tree.FindDeepestTag(*ipB)
	assert.True(t, found)
	found, _ = tree.FindDeepestTag(*ipC)
	assert.True(t, found)
	found, _ = tree.FindDeepestTag(*ipD)
	assert.True(t, found)

	// delete each one
	matchFunc := func(a GeneratedType, b GeneratedType) bool {
		return a == b
	}
	deleteCount := tree.Delete(*ipA, matchFunc, "A")
	assert.Equal(t, 1, deleteCount)
	deleteCount = tree.Delete(*ipB, matchFunc, "B")
	assert.Equal(t, 1, deleteCount)
	deleteCount = tree.Delete(*ipC, matchFunc, "C")
	assert.Equal(t, 1, deleteCount)
	deleteCount = tree.Delete(*ipD, matchFunc, "D")
	assert.Equal(t, 1, deleteCount)

	// should have zero logical nodes except for root
	assert.Equal(t, 1, tree.countNodes(1))
}

func TestSimpleTree(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tree := NewTreeV4()

	for i := 32; i > 0; i-- {
		countIncreased, count := tree.Add(ipv4FromBytes([]byte{127, 0, 0, 1}, i), fmt.Sprintf("Tag-%d", i), nil)
		assert.True(t, countIncreased)
		assert.Equal(t, 1, count)
	}

	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{127, 0, 0, 1}, 32))
	if assert.Equal(t, 32, len(tags)) {
		assert.Equal(t, "Tag-32", tags[31].(string))
		assert.Equal(t, "Tag-31", tags[30].(string))
	}

	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{63, 3, 0, 1}, 32))
	if assert.Equal(t, 1, len(tags)) {
		assert.Equal(t, "Tag-1", tags[0].(string))
	}

	// find deepest tag: match at lowest level
	found, tag := tree.FindDeepestTag(ipv4FromBytes([]byte{127, 0, 0, 1}, 32))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-32", tag.(string))
	}

	// find deepest tag: match at top level
	found, tag = tree.FindDeepestTag(ipv4FromBytes([]byte{63, 5, 4, 3}, 32))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-1", tag.(string))
	}

	// find deepest tag: match at mid level
	found, tag = tree.FindDeepestTag(ipv4FromBytes([]byte{119, 5, 4, 3}, 32))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-4", tag.(string))
	}

	// find deepest tag: no match
	found, tag = tree.FindDeepestTag(ipv4FromBytes([]byte{128, 4, 3, 2}, 32))
	assert.False(t, found)
	assert.Zero(t, tag)

	// Add a couple root tags
	countIncreased, count := tree.Add(ipv4FromBytes([]byte{127, 0, 0, 1}, 0), "root1", nil)
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)
	countIncreased, count = tree.Add(patricia.IPv4Address{}, "root2", nil)
	assert.True(t, countIncreased)
	assert.Equal(t, 2, count)

	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, patricia.IPv4Address{})
	if assert.Equal(t, 2, len(tags)) {
		assert.Equal(t, "root1", tags[0].(string))
		assert.Equal(t, "root2", tags[1].(string))
	}

}

// assert two collections of arrays have the same tags - don't worry about performance
func tagArraysEqual(a []GeneratedType, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		found := false
		for j := 0; j < len(b); j++ {
			if a[i].(string) == b[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestTree1FindTagsAppend(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	tree := NewTreeV4()
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 0), tagZ, nil) // default
	tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), tagA, nil)
	tree.Add(ipv4FromBytes([]byte{160, 0, 0, 0}, 2), tagB, nil) // 160 -> 128
	tree.Add(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), tagC, nil)

	// three tags in a hierarchy - ask for all but the most specific
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{128, 142, 133, 1}, 32))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagZ}))

	// three tags in a hierarchy - ask for an exact match, receive all 3
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))

	// three tags in a hierarchy - get just the first
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{162, 1, 0, 5}, 30))
	assert.True(t, tagArraysEqual(tags, []string{tagB, tagZ}))

	// three tags in hierarchy - get none
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{1, 0, 0, 0}, 1))
	assert.True(t, tagArraysEqual(tags, []string{tagZ}))
}

func TestTree1FindTagsWithFilterAppend(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	filterFunc := func(val GeneratedType) bool {
		return val == tagA || val == tagB
	}

	tree := NewTreeV4()
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 0), tagZ, nil) // default
	tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), tagA, nil)
	tree.Add(ipv4FromBytes([]byte{160, 0, 0, 0}, 2), tagB, nil) // 160 -> 128
	tree.Add(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), tagC, nil)

	// three tags in a hierarchy - ask for all but the most specific
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{128, 142, 133, 1}, 32), filterFunc)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// three tags in a hierarchy - ask for an exact match, receive all 3
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32), filterFunc)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// three tags in a hierarchy - get just the first
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{162, 1, 0, 5}, 30), filterFunc)
	assert.True(t, tagArraysEqual(tags, []string{tagB}))

	// three tags in hierarchy - get none
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv4FromBytes([]byte{1, 0, 0, 0}, 1), filterFunc)
	assert.Zero(t, len(tags))
}

// Test that all queries get the root nodes
func TestRootNode(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagD := "tagD"
	tagZ := "tagE"

	tree := NewTreeV4()

	// root node gets tags A & B
	tree.Add(patricia.IPv4Address{}, tagA, nil)
	tree.Add(patricia.IPv4Address{}, tagB, nil)

	// query the root node with no address
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, patricia.IPv4Address{})
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// query a node that doesn't exist
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{1, 2, 3, 4}, 32))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// create a new /16 node with C & D
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 16), tagC, nil)
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 16), tagD, nil)
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{1, 2, 3, 4}, 16))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD}))

	// create a node under the /16 node
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 32), tagZ, nil)
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{1, 2, 3, 4}, 32))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD, tagZ}))

	// check the /24 and make sure we still get the /16 and root
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{1, 2, 3, 4}, 24))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD}))
}

// TestAdd returns the right counts
func TestAdd(t *testing.T) {
	address := ipv4FromBytes([]byte{1, 2, 3, 4}, 32)

	tree := NewTreeV4()
	countIncreased, count := tree.Add(address, "hi", nil)
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)

	countIncreased, count = tree.Add(address, "hi", nil)
	assert.True(t, countIncreased)
	assert.Equal(t, 2, count)

	countIncreased, count = tree.Add(address, "hi", nil)
	assert.True(t, countIncreased)
	assert.Equal(t, 3, count)
}

// Test setting a value to a node, rather than adding to a list
func TestSet(t *testing.T) {
	buf := make([]GeneratedType, 0)

	address := ipv4FromBytes([]byte{1, 2, 3, 4}, 32)

	tree := NewTreeV4()

	// add a parent node, just to mix things up
	countIncreased, count := tree.Set(ipv4FromBytes([]byte{1, 2, 3, 0}, 24), "parent")
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)

	countIncreased, count = tree.Set(address, "tagA")
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)
	found, tag := tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagA", tag)

	countIncreased, count = tree.Set(address, "tagB")
	assert.Equal(t, 1, count)
	assert.False(t, countIncreased)
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagB", tag)

	countIncreased, count = tree.Set(address, "tagC")
	assert.Equal(t, 1, count)
	assert.False(t, countIncreased)
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagC", tag)

	countIncreased, count = tree.Set(address, "tagD")
	assert.Equal(t, 1, count)
	assert.False(t, countIncreased)
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagD", tag)

	// now delete the tag
	delCount := tree.DeleteWithBuffer(buf, address, func(a GeneratedType, b GeneratedType) bool { return true }, "")
	assert.Equal(t, 1, delCount)

	// verify it's gone - should get the parent
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "parent", tag)
}

func TestSetOrUpdate(t *testing.T) {
	buf := make([]GeneratedType, 0)

	address := ipv4FromBytes([]byte{1, 2, 3, 4}, 32)

	tree := NewTreeV4()

	// add a parent node, just to mix things up
	countIncreased, count := tree.Set(ipv4FromBytes([]byte{1, 2, 3, 0}, 24), "parent")
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)

	countIncreased, count = tree.SetOrUpdate(address, "tagA",
		func(GeneratedType) GeneratedType { return "tagZ" })
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)
	found, tag := tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagA", tag)

	countIncreased, count = tree.SetOrUpdate(address, "tagZ",
		func(old GeneratedType) GeneratedType { return old.(string) + "B" })
	assert.Equal(t, 1, count)
	assert.False(t, countIncreased)
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagAB", tag)

	countIncreased, count = tree.SetOrUpdate(address, "tagZ",
		func(old GeneratedType) GeneratedType { return old.(string) + "C" })
	assert.Equal(t, 1, count)
	assert.False(t, countIncreased)
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagABC", tag)

	countIncreased, count = tree.SetOrUpdate(address, "tagZ",
		func(old GeneratedType) GeneratedType { return old.(string) + "D" })
	assert.Equal(t, 1, count)
	assert.False(t, countIncreased)
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagABCD", tag)

	// now delete the tag
	delCount := tree.DeleteWithBuffer(buf, address, func(a GeneratedType, b GeneratedType) bool { return true }, "")
	assert.Equal(t, 1, delCount)

	// verify it's gone - should get the parent
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "parent", tag)
}

func TestAddOrUpdate(t *testing.T) {
	buf := make([]GeneratedType, 0)

	address := ipv4FromBytes([]byte{1, 2, 3, 4}, 32)

	tree := NewTreeV4()

	// add a parent node, just to mix things up
	countIncreased, count := tree.Set(ipv4FromBytes([]byte{1, 2, 3, 0}, 24), "parent")
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)

	countIncreased, count = tree.AddOrUpdate(address, "tagA",
		nil,
		func(GeneratedType) GeneratedType { return "tagZ" })
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)
	found, tag := tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagA", tag)

	countIncreased, count = tree.AddOrUpdate(address, "tagZ",
		func(GeneratedType, GeneratedType) bool { return true },
		func(old GeneratedType) GeneratedType { return old.(string) + "B" })
	assert.Equal(t, 1, count)
	assert.False(t, countIncreased)
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagAB", tag)

	countIncreased, count = tree.AddOrUpdate(address, "tagAB",
		func(val1 GeneratedType, val2 GeneratedType) bool { return val1 == val2 },
		func(old GeneratedType) GeneratedType { return old.(string) + "C" })
	assert.Equal(t, 1, count)
	assert.False(t, countIncreased)
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "tagABC", tag)

	countIncreased, count = tree.AddOrUpdate(address, "tagABCD",
		func(val1 GeneratedType, val2 GeneratedType) bool { return val1 == val2 },
		func(old GeneratedType) GeneratedType { return old.(string) + "Z" })
	assert.Equal(t, 2, count)
	assert.True(t, countIncreased)
	found, tags := tree.FindDeepestTags(address)
	assert.True(t, found)
	assert.True(t, tagArraysEqual(tags, []string{"tagABC", "tagABCD"}))

	countIncreased, count = tree.AddOrUpdate(address, "tagABC",
		func(val1 GeneratedType, val2 GeneratedType) bool { return val1 == val2 },
		func(old GeneratedType) GeneratedType { return old.(string) + "DE" })
	assert.Equal(t, 2, count)
	assert.False(t, countIncreased)
	found, tags = tree.FindDeepestTags(address)
	assert.True(t, found)
	assert.True(t, tagArraysEqual(tags, []string{"tagABCDE", "tagABCD"}))

	// now delete the tag
	delCount := tree.DeleteWithBuffer(buf, address, func(a GeneratedType, b GeneratedType) bool { return true }, "")
	assert.Equal(t, 2, delCount)

	// verify it's gone - should get the parent
	found, tag = tree.FindDeepestTag(address)
	assert.True(t, found)
	assert.Equal(t, "parent", tag)
}

func TestDelete1(t *testing.T) {
	tags := make([]GeneratedType, 0)

	matchFunc := func(tagData GeneratedType, val GeneratedType) bool {
		return tagData.(string) == val.(string)
	}

	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagZ"

	tree := NewTreeV4()
	assert.Equal(t, 1, tree.countNodes(1))
	tree.Add(ipv4FromBytes([]byte{8, 7, 6, 5}, 0), tagZ, nil) // default
	assert.Equal(t, 1, tree.countNodes(1))
	assert.Zero(t, len(tree.availableIndexes))
	assert.Equal(t, 2, len(tree.nodes)) // empty first node plus root

	tree.Add(ipv4FromBytes([]byte{128, 3, 0, 5}, 7), tagA, nil) // 1000000
	assert.Equal(t, 2, tree.countNodes(1))
	assert.Equal(t, 3, len(tree.nodes))

	tree.Add(ipv4FromBytes([]byte{128, 5, 1, 1}, 2), tagB, nil) // 10
	assert.Equal(t, 3, tree.countNodes(1))
	assert.Equal(t, 4, len(tree.nodes))

	tree.Add(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), tagC, nil)
	assert.Equal(t, 4, tree.countNodes(1))
	assert.Equal(t, 5, len(tree.nodes))

	// verify status of internal nodes collections
	assert.Zero(t, len(tree.availableIndexes))
	assert.Equal(t, "tagZ", tree.tagsForNode(tags, 1, nil)[0], nil)
	tags = tags[:0]
	assert.Equal(t, "tagA", tree.tagsForNode(tags, 2, nil)[0], nil)
	tags = tags[:0]
	assert.Equal(t, "tagB", tree.tagsForNode(tags, 3, nil)[0], nil)
	tags = tags[:0]
	assert.Equal(t, "tagC", tree.tagsForNode(tags, 4, nil)[0], nil)
	tags = tags[:0]

	// three tags in a hierarchy - ask for an exact match, receive all 3 and the root
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))

	// 1. delete a tag that doesn't exist
	count := 0
	count = tree.DeleteWithBuffer(tags, ipv4FromBytes([]byte{9, 9, 9, 9}, 32), matchFunc, "bad tag")
	assert.Equal(t, 0, count)
	assert.Equal(t, 4, tree.countNodes(1))
	assert.Equal(t, 4, tree.countTags(1))

	// 2. delete a tag on an address that exists, but doesn't have the tag
	count = tree.DeleteWithBuffer(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32), matchFunc, "bad tag")
	assert.Equal(t, 0, count)

	// verify
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))
	assert.Equal(t, 4, tree.countNodes(1))
	assert.Equal(t, 4, tree.countTags(1))

	// 3. delete the default/root tag
	count = tree.DeleteWithBuffer(tags, ipv4FromBytes([]byte{0, 0, 0, 0}, 0), matchFunc, "tagZ")
	assert.Equal(t, 1, count)
	assert.Equal(t, 4, tree.countNodes(1)) // doesn't delete anything
	assert.Equal(t, 3, tree.countTags(1))
	assert.Equal(t, 0, len(tree.availableIndexes))

	// three tags in a hierarchy - ask for an exact match, receive all 3, not the root, which we deleted
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC}))

	// 4. delete tagA
	count = tree.DeleteWithBuffer(tags, ipv4FromBytes([]byte{128, 0, 0, 0}, 7), matchFunc, "tagA")
	assert.Equal(t, 1, count)

	// verify
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.True(t, tagArraysEqual(tags, []string{tagB, tagC}))
	assert.Equal(t, 3, tree.countNodes(1))
	assert.Equal(t, 2, tree.countTags(1))
	assert.Equal(t, 1, len(tree.availableIndexes))
	assert.Equal(t, uint(2), tree.availableIndexes[0])

	// 5. delete tag B
	count = tree.DeleteWithBuffer(tags, ipv4FromBytes([]byte{128, 0, 0, 0}, 2), matchFunc, "tagB")
	assert.Equal(t, 1, count)

	// verify
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.True(t, tagArraysEqual(tags, []string{tagC}))
	assert.Equal(t, 2, tree.countNodes(1))
	assert.Equal(t, 1, tree.countTags(1))
	assert.Equal(t, 2, len(tree.availableIndexes))
	assert.Equal(t, uint(3), tree.availableIndexes[1])

	// add tagE & tagF to the same node
	tree.Add(ipv4FromBytes([]byte{1, 3, 6, 240}, 32), "tagE", nil)
	tree.Add(ipv4FromBytes([]byte{1, 3, 6, 240}, 32), "tagF", nil)
	assert.Equal(t, 3, tree.countNodes(1))
	assert.Equal(t, 3, tree.countTags(1))

	// this should be recycling tagB
	assert.Equal(t, 1, len(tree.availableIndexes))
	assert.Equal(t, uint(2), tree.availableIndexes[0])

	tags = tags[:0]
	assert.Equal(t, "tagE", tree.tagsForNode(tags, 3, nil)[0])

	// 6. delete tag C
	count = tree.DeleteWithBuffer(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32), matchFunc, "tagC")
	assert.Equal(t, 1, count)

	// verify
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.True(t, tagArraysEqual(tags, []string{}))
	assert.Equal(t, 2, tree.countNodes(1))
	assert.Equal(t, 2, tree.countTags(1))
}

func TestTryToBreak(t *testing.T) {
	tree := NewTreeV4()
	for a := byte(1); a < 10; a++ {
		for b := byte(1); b < 10; b++ {
			for c := byte(1); c < 10; c++ {
				for d := byte(1); d < 10; d++ {
					tree.Add(ipv4FromBytes([]byte{a, b, c, d}, rand.Intn(32)), "tag", nil)
				}
			}
		}
	}
}

func TestTagsMap(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tree := NewTreeV4()

	// insert tags
	tree.addTag("tagA", 1, nil, nil)
	tree.addTag("tagB", 1, nil, nil)
	tree.addTag("tagC", 1, nil, nil)
	tree.addTag("tagD", 0, nil, nil) // there's no node0, but it exists, so use it for this test

	// verify
	assert.Equal(t, 3, tree.nodes[1].TagCount)
	assert.Equal(t, "tagA", tree.firstTagForNode(1))
	assert.Equal(t, 3, len(tree.tagsForNode(tags, 1, nil)))
	tags = tags[:0]

	assert.Equal(t, "tagA", tree.tagsForNode(tags, 1, nil)[0])
	tags = tags[:0]

	assert.Equal(t, "tagB", tree.tagsForNode(tags, 1, nil)[1])
	tags = tags[:0]

	assert.Equal(t, "tagC", tree.tagsForNode(tags, 1, nil)[2])
	tags = tags[:0]

	// delete tagB
	matchesFunc := func(payload GeneratedType, val GeneratedType) bool {
		return payload == val
	}
	deleted, kept := tree.deleteTag(tags, 1, "tagB", matchesFunc)
	assert.Equal(t, 0, len(tags))

	// verify
	assert.Equal(t, 1, deleted)
	assert.Equal(t, 2, kept)
	assert.Equal(t, 2, tree.nodes[1].TagCount)
	assert.Equal(t, "tagA", tree.tagsForNode(tags, 1, nil)[0])
	tags = tags[:0]

	assert.Equal(t, "tagC", tree.tagsForNode(tags, 1, nil)[1])
}

// test duplicate tags with no match func
func TestDuplicateTagsWithNoMatchFunc(t *testing.T) {
	matchFunc := MatchesFunc(nil)

	tree := NewTreeV4()

	wasAdded, count := tree.Add(patricia.IPv4Address{}, "FOO", matchFunc) // default
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	wasAdded, count = tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	// add another at previous node
	wasAdded, count = tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), "FOOBAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 2, count)

	// add a dupe to the previous node - will be fine since match is nil
	wasAdded, count = tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 3, count)
}

// test duplicate tags with match func that always returns false
func TestDuplicateTagsWithFalseMatchFunc(t *testing.T) {
	matchFunc := func(val1 GeneratedType, val2 GeneratedType) bool {
		return false
	}

	tree := NewTreeV4()

	wasAdded, count := tree.Add(patricia.IPv4Address{}, "FOO", matchFunc) // default
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	wasAdded, count = tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	// add another at previous node
	wasAdded, count = tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), "FOOBAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 2, count)

	// add a dupe to the previous node - will be fine since match is nil
	wasAdded, count = tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 3, count)
}

// test duplicate tags with match func that does something
func TestDuplicateTagsWithMatchFunc(t *testing.T) {
	matchFunc := func(val1 GeneratedType, val2 GeneratedType) bool {
		return val1 == val2
	}

	tree := NewTreeV4()

	wasAdded, count := tree.Add(patricia.IPv4Address{}, "FOO", matchFunc) // default
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	wasAdded, count = tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	// add another at previous node
	wasAdded, count = tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), "FOOBAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 2, count)

	// add a dupe to the previous node - will be fine since match is nil
	wasAdded, count = tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), "BAR", matchFunc)
	assert.False(t, wasAdded)
	assert.Equal(t, 2, count)
}

// test tree traversal
func TestIterateV4(t *testing.T) {
	tree := NewTreeV4()

	// try an empty tree first
	iter := tree.Iterate()
	for iter.Next() {
		assert.Fail(t, "empty tree should not have a next element")
	}

	ipA := ipv4FromBytes([]byte{203, 143, 220, 0}, 23)
	ipB := ipv4FromBytes([]byte{203, 143, 220, 198}, 32)
	ipC := ipv4FromBytes([]byte{203, 143, 0, 0}, 16)
	ipD := ipv4FromBytes([]byte{203, 143, 221, 75}, 32)

	// add the 4 addresses
	tree.Add(ipA, "A", nil)
	tree.Add(ipB, "B", nil)
	tree.Add(ipC, "C", nil)
	tree.Add(ipD, "D1", nil)
	tree.Add(ipD, "D2", nil)

	expected := map[string][]string{
		"203.143.220.0/23":   {"A"},
		"203.143.220.198/32": {"B"},
		"203.143.0.0/16":     {"C"},
		"203.143.221.75/32":  {"D1", "D2"},
	}
	got := map[string][]string{}
	iter = tree.Iterate()
	for iter.Next() {
		tags := []string{}
		for _, s := range iter.Tags() { //nolint:gosimple
			tags = append(tags, s.(string))
		}
		got[iter.Address().String()] = tags
	}
	assert.Equal(t, expected, got)
}

// test deletion during tree traversal
func TestIterateAndDeleteV4(t *testing.T) {
	tree := NewTreeV4()

	compare := func(expected [][]string) {
		t.Helper()
		got := [][]string{}
		iter := tree.Iterate()
		for iter.Next() {
			tags := []string{}
			for _, s := range iter.Tags() { //nolint:gosimple
				tags = append(tags, s.(string))
			}
			got = append(got, append([]string{iter.Address().String()}, tags...))
		}
		assert.Equal(t, expected, got)
	}

	ipA := ipv4FromBytes([]byte{203, 143, 220, 0}, 23)
	ipB := ipv4FromBytes([]byte{203, 143, 220, 198}, 31)
	ipC := ipv4FromBytes([]byte{203, 143, 0, 0}, 16)
	ipD := ipv4FromBytes([]byte{203, 143, 221, 75}, 32)
	ipE := ipv4FromBytes([]byte{203, 143, 220, 198}, 32)

	tree.Add(ipA, "A", nil)
	tree.Add(ipB, "B", nil)
	tree.Add(ipC, "C", nil)
	tree.Add(ipD, "D1", nil)
	tree.Add(ipD, "D2", nil)
	compare([][]string{
		{"203.143.0.0/16", "C"},
		{"203.143.220.0/23", "A"},
		{"203.143.220.198/31", "B"},
		{"203.143.221.75/32", "D1", "D2"},
	})

	// Delete one tag, no node
	iter := tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "D1"
		}, "")
	}
	compare([][]string{
		{"203.143.0.0/16", "C"},
		{"203.143.220.0/23", "A"},
		{"203.143.220.198/31", "B"},
		{"203.143.221.75/32", "D2"},
	})

	// Delete a node with two children
	iter = tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "A"
		}, "")
	}
	compare([][]string{
		{"203.143.0.0/16", "C"},
		{"203.143.220.198/31", "B"},
		{"203.143.221.75/32", "D2"},
	})

	// Delete right children of a node two children
	iter = tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "D2"
		}, "")
	}
	compare([][]string{
		{"203.143.0.0/16", "C"},
		{"203.143.220.198/31", "B"},
	})

	// Delete leaf
	iter = tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "B"
		}, "")
	}
	compare([][]string{
		{"203.143.0.0/16", "C"},
	})

	// Delete root
	iter = tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "C"
		}, "")
	}
	compare([][]string{})

	// Delete a node with a left children only
	tree = NewTreeV4()
	tree.Add(ipA, "A", nil)
	tree.Add(ipB, "B", nil)
	tree.Add(ipC, "C", nil)
	iter = tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "A"
		}, "")
	}
	compare([][]string{
		{"203.143.0.0/16", "C"},
		{"203.143.220.198/31", "B"},
	})

	// Delete a node with a right children only
	tree = NewTreeV4()
	tree.Add(ipA, "A", nil)
	tree.Add(ipC, "C", nil)
	tree.Add(ipD, "D", nil)
	iter = tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "A"
		}, "")
	}
	compare([][]string{
		{"203.143.0.0/16", "C"},
		{"203.143.221.75/32", "D"},
	})

	// Delete a node without children and at the left of its empty parent
	tree = NewTreeV4()
	tree.Add(ipB, "B", nil)
	tree.Add(ipC, "C", nil)
	tree.Add(ipD, "D", nil)
	iter = tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "B"
		}, "")
	}
	compare([][]string{
		{"203.143.0.0/16", "C"},
		{"203.143.221.75/32", "D"},
	})

	// Delete left node while a right node is present
	tree = NewTreeV4()
	tree.Add(ipA, "A", nil)
	tree.Add(ipB, "B", nil)
	tree.Add(ipC, "C", nil)
	tree.Add(ipD, "D", nil)
	iter = tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "B"
		}, "")
	}
	compare([][]string{
		{"203.143.0.0/16", "C"},
		{"203.143.220.0/23", "A"},
		{"203.143.221.75/32", "D"},
	})

	// Delete left node with a child
	tree = NewTreeV4()
	tree.Add(ipA, "A", nil)
	tree.Add(ipB, "B", nil)
	tree.Add(ipC, "C", nil)
	tree.Add(ipD, "D", nil)
	tree.Add(ipE, "E", nil)
	iter = tree.Iterate()
	for iter.Next() {
		iter.Delete(func(payload, val GeneratedType) bool {
			return payload == "B"
		}, "")
	}
	compare([][]string{
		{"203.143.0.0/16", "C"},
		{"203.143.220.0/23", "A"},
		{"203.143.220.198/32", "E"},
		{"203.143.221.75/32", "D"},
	})
}
