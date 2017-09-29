package template

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"

	"github.com/kentik/patricia"
	"github.com/stretchr/testify/assert"
)

func ipv6FromString(address string, length int) patricia.IPv6Address {
	ip, _, err := net.ParseCIDR(address)
	if err != nil {
		panic(fmt.Sprintf("Invalid IP address: %s: %s", address, err))
	}
	return patricia.IPv6Address{
		Left:   binary.BigEndian.Uint64([]byte(ip[:8])),
		Right:  binary.BigEndian.Uint64([]byte(ip[8:])),
		Length: uint(length),
	}
}

func BenchmarkFindTagsV6(b *testing.B) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	tree := NewTreeV6()

	tree.Add(patricia.IPv6Address{}, tagZ) // default
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagA)
	tree.Add(ipv6FromString("2001:db8:0:0:0:5:2:1/128", 16), tagB) // 160 -> 128
	tree.Add(ipv6FromString("2001:db7:0:0:0:0:2:1/128", 77), tagC)

	address := ipv6FromString("2001:db7:0:0:0:0:2:1/128", 32)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree.FindTags(address)
	}
}

func BenchmarkFindDeepestTagV6(b *testing.B) {
	tree := NewTreeV6()
	for i := 128; i > 0; i-- {
		tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", i), fmt.Sprintf("Tag-%d", i))
	}
	address := ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree.FindDeepestTag(address)
	}
}

func TestSimpleTreeV6(t *testing.T) {
	tree := NewTreeV6()

	for i := 128; i > 0; i-- {
		countIncreased, count, err := tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", i), fmt.Sprintf("Tag-%d", i))
		assert.NoError(t, err)
		assert.True(t, countIncreased)
		assert.Equal(t, 1, count)
	}

	tags, err := tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	if assert.Equal(t, 128, len(tags)) {
		assert.Equal(t, "Tag-128", tags[127].(string))
		assert.Equal(t, "Tag-32", tags[31].(string))
	}

	tags, err = tree.FindTags(ipv6FromString("4001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	if assert.Equal(t, 1, len(tags)) {
		assert.Equal(t, "Tag-1", tags[0])
	}

	// find deepest tag: match at lowest level
	found, tag, err := tree.FindDeepestTag(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-128", tag)
	}

	// find deepest tag: match at top level
	found, tag, err = tree.FindDeepestTag(ipv6FromString("7001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-1", tag.(string))
	}
	found, tag, err = tree.FindDeepestTag(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 1))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-1", tag.(string))
	}

	// find deepest tag: match at mid level
	found, tag, err = tree.FindDeepestTag(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 32))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-32", tag.(string))
	}
	found, tag, err = tree.FindDeepestTag(ipv6FromString("2001:db8:FFFF:0:0:0:2:1/128", 128))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-32", tag.(string))
	}

	// find deepest tag: no match
	found, tag, err = tree.FindDeepestTag(ipv6FromString("F001:db8:1:0:0:0:2:1/128", 32))
	assert.False(t, found)
	assert.Nil(t, tag)

	// Add a couple root tags
	countIncreased, count, err := tree.Add(ipv6FromString("2001:db8:1:0:0:0:2:1/128", 0), "root1")
	assert.NoError(t, err)
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)
	countIncreased, count, err = tree.Add(patricia.IPv6Address{}, "root2")
	assert.NoError(t, err)
	assert.True(t, countIncreased)
	assert.Equal(t, 2, count)

	tags, err = tree.FindTags(patricia.IPv6Address{})
	assert.NoError(t, err)
	if assert.Equal(t, 2, len(tags)) {
		assert.Equal(t, "root1", tags[0].(string))
		assert.Equal(t, "root2", tags[1].(string))
	}

}

func TestTree1V6(t *testing.T) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	tree := NewTreeV6()
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 0), tagZ) // default
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 100), tagA)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 67), tagB)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagC)

	// three tags in a hierarchy - ask for all but the most specific
	tags, err := tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:0/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagZ}))
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 127))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagZ}))

	// three tags in a hierarchy - ask for an exact match, receive all 3
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))

	// three tags in a hierarchy - get just the first
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:1:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagB, tagZ}))

	// three tags in hierarchy - get none
	tags, err = tree.FindTags(ipv6FromString("8001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagZ}))
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 66))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagZ}))
}

func TestTree1V6WithFilter(t *testing.T) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	filterFunc := func(val GeneratedType) bool {
		return val == "tagA" || val == "tagB"
	}

	tree := NewTreeV6()
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 0), tagZ) // default
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 100), tagA)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 67), tagB)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagC)

	// three tags in a hierarchy - ask for all but the most specific
	tags, err := tree.FindTagsWithFilter(ipv6FromString("2001:db8:0:0:0:0:2:0/128", 128), filterFunc)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))
	tags, err = tree.FindTagsWithFilter(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 127), filterFunc)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// three tags in a hierarchy - ask for an exact match, receive all 3
	tags, err = tree.FindTagsWithFilter(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), filterFunc)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// three tags in a hierarchy - get just the first
	tags, err = tree.FindTagsWithFilter(ipv6FromString("2001:db8:0:0:0:1:2:1/128", 128), filterFunc)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagB}))

	// three tags in hierarchy - get none
	tags, err = tree.FindTagsWithFilter(ipv6FromString("8001:db8:0:0:0:0:2:1/128", 128), filterFunc)
	assert.NoError(t, err)
	assert.Zero(t, len(tags))
	tags, err = tree.FindTagsWithFilter(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 66), filterFunc)
	assert.NoError(t, err)
	assert.Zero(t, len(tags))
}

// Test that all queries get the root nodes
func TestRootNodeV6(t *testing.T) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagD := "tagD"
	tagZ := "tagE"

	tree := NewTreeV6()

	// root node gets tags A & B
	tree.Add(patricia.IPv6Address{}, tagA)
	tree.Add(patricia.IPv6Address{}, tagB)

	// query the root node with no address
	tags, err := tree.FindTags(patricia.IPv6Address{})
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// query a node that doesn't exist
	tags, err = tree.FindTags(ipv6FromString("FFFF:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// create a new /65 node with C & D
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 65), tagC)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 65), tagD)
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD}))

	// create a node under the /65 node
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagZ)
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD, tagZ}))

	// check the /77 and make sure we still get the /65 and root
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 77))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD}))
}

func TestDelete1V6(t *testing.T) {
	matchFunc := func(tagData GeneratedType, val GeneratedType) bool {
		return tagData.(string) == val.(string)
	}

	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagZ"

	tree := NewTreeV6()
	assert.Equal(t, 1, tree.countNodes(1))
	tree.Add(patricia.IPv6Address{}, tagZ) // default
	assert.Equal(t, 1, tree.countNodes(1))
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/67", 67), tagA) // 1000000
	assert.Equal(t, 2, tree.countNodes(1))
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/2", 2), tagB) // 10
	assert.Equal(t, 3, tree.countNodes(1))
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagC)
	assert.Equal(t, 4, tree.countNodes(1))

	// three tags in a hierarchy - ask for an exact match, receive all 3 and the root
	tags, err := tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))

	// 1. delete a tag that doesn't exist
	count := 0
	count, err = tree.Delete(ipv6FromString("F001:db8:0:0:0:0:2:1/128", 128), matchFunc, "bad tag")
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 4, tree.countTags(1))
	assert.Equal(t, 4, tree.countNodes(1))

	// 2. delete a tag on an address that exists, but doesn't have the tag
	count, err = tree.Delete(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 67), matchFunc, "bad tag")
	assert.Equal(t, 0, count)
	assert.NoError(t, err)

	// verify
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))
	assert.Equal(t, 4, tree.countNodes(1))
	assert.Equal(t, 4, tree.countTags(1))

	// 3. delete the default/root tag
	count, err = tree.Delete(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 0), matchFunc, "tagZ")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 4, tree.countNodes(1)) // doesn't delete anything
	assert.Equal(t, 3, tree.countTags(1))

	// three tags in a hierarchy - ask for an exact match, receive all 3, not the root, which we deleted
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC}))

	// 4. delete tagA
	count, err = tree.Delete(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 67), matchFunc, "tagA")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// verify
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagB, tagC}))
	assert.Equal(t, 3, tree.countNodes(1))
	assert.Equal(t, 2, tree.countTags(1))

	// 5. delete tag B
	count, err = tree.Delete(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 2), matchFunc, "tagB")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// verify
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagC}))
	assert.Equal(t, 2, tree.countNodes(1))
	assert.Equal(t, 1, tree.countTags(1))

	// 6. delete tag C
	count, err = tree.Delete(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), matchFunc, "tagC")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// verify
	tags, err = tree.FindTags(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{}))
	assert.Equal(t, 1, tree.countNodes(1))
	assert.Equal(t, 0, tree.countTags(1))
}
