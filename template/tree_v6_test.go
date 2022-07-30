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

	tree.Add(patricia.IPv6Address{}, tagZ, nil) // default
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagA, nil)
	tree.Add(ipv6FromString("2001:db8:0:0:0:5:2:1/128", 16), tagB, nil) // 160 -> 128
	tree.Add(ipv6FromString("2001:db7:0:0:0:0:2:1/128", 77), tagC, nil)

	buf := make([]GeneratedType, 0)
	address := ipv6FromString("2001:db7:0:0:0:0:2:1/128", 32)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree.FindTagsAppend(buf, address)
		buf = buf[:0]
	}
}

func BenchmarkFindDeepestTagV6(b *testing.B) {
	tree := NewTreeV6()
	for i := 128; i > 0; i-- {
		tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", i), fmt.Sprintf("Tag-%d", i), nil)
	}
	address := ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree.FindDeepestTag(address)
	}
}

func TestSimpleTreeV6(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tree := NewTreeV6()

	for i := 128; i > 0; i-- {
		countIncreased, count := tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", i), fmt.Sprintf("Tag-%d", i), nil)
		assert.True(t, countIncreased)
		assert.Equal(t, 1, count)
	}

	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	if assert.Equal(t, 128, len(tags)) {
		assert.Equal(t, "Tag-128", tags[127].(string))
		assert.Equal(t, "Tag-32", tags[31].(string))
	}

	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("4001:db8:0:0:0:0:2:1/128", 128))
	if assert.Equal(t, 1, len(tags)) {
		assert.Equal(t, "Tag-1", tags[0])
	}

	// find deepest tag: match at lowest level
	found, tag := tree.FindDeepestTag(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-128", tag)
	}

	// find deepest tag: match at top level
	found, tag = tree.FindDeepestTag(ipv6FromString("7001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-1", tag.(string))
	}
	found, tag = tree.FindDeepestTag(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 1))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-1", tag.(string))
	}

	// find deepest tag: match at mid level
	found, tag = tree.FindDeepestTag(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 32))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-32", tag.(string))
	}
	found, tag = tree.FindDeepestTag(ipv6FromString("2001:db8:FFFF:0:0:0:2:1/128", 128))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-32", tag.(string))
	}

	// find deepest tag: no match
	found, tag = tree.FindDeepestTag(ipv6FromString("F001:db8:1:0:0:0:2:1/128", 32))
	assert.False(t, found)
	assert.Zero(t, tag)

	// Add a couple root tags
	countIncreased, count := tree.Add(ipv6FromString("2001:db8:1:0:0:0:2:1/128", 0), "root1", nil)
	assert.True(t, countIncreased)
	assert.Equal(t, 1, count)
	countIncreased, count = tree.Add(patricia.IPv6Address{}, "root2", nil)
	assert.True(t, countIncreased)
	assert.Equal(t, 2, count)

	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, patricia.IPv6Address{})
	if assert.Equal(t, 2, len(tags)) {
		assert.Equal(t, "root1", tags[0].(string))
		assert.Equal(t, "root2", tags[1].(string))
	}
}

func TestTree1V6(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	tree := NewTreeV6()
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 0), tagZ, nil) // default
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 100), tagA, nil)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 67), tagB, nil)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagC, nil)

	// three tags in a hierarchy - ask for all but the most specific
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:0/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagZ}))
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 127))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagZ}))

	// three tags in a hierarchy - ask for an exact match, receive all 3
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))

	// three tags in a hierarchy - get just the first
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:1:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagB, tagZ}))

	// three tags in hierarchy - get none
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("8001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagZ}))
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 66))
	assert.True(t, tagArraysEqual(tags, []string{tagZ}))
}

func TestTree1V6WithFilter(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	filterFunc := func(val GeneratedType) bool {
		return val == "tagA" || val == "tagB"
	}

	tree := NewTreeV6()
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 0), tagZ, nil) // default
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 100), tagA, nil)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 67), tagB, nil)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagC, nil)

	// three tags in a hierarchy - ask for all but the most specific
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:0/128", 128), filterFunc)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 127), filterFunc)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// three tags in a hierarchy - ask for an exact match, receive all 3
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), filterFunc)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// three tags in a hierarchy - get just the first
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv6FromString("2001:db8:0:0:0:1:2:1/128", 128), filterFunc)
	assert.True(t, tagArraysEqual(tags, []string{tagB}))

	// three tags in hierarchy - get none
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv6FromString("8001:db8:0:0:0:0:2:1/128", 128), filterFunc)
	assert.Zero(t, len(tags))
	tags = tags[:0]
	tags = tree.FindTagsWithFilterAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 66), filterFunc)
	assert.Zero(t, len(tags))
}

// Test that all queries get the root nodes
func TestRootNodeV6(t *testing.T) {
	tags := make([]GeneratedType, 0)

	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagD := "tagD"
	tagZ := "tagE"

	tree := NewTreeV6()

	// root node gets tags A & B
	tree.Add(patricia.IPv6Address{}, tagA, nil)
	tree.Add(patricia.IPv6Address{}, tagB, nil)

	// query the root node with no address
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, patricia.IPv6Address{})
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// query a node that doesn't exist
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("FFFF:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// create a new /65 node with C & D
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 65), tagC, nil)
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 65), tagD, nil)
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD}))

	// create a node under the /65 node
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagZ, nil)
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD, tagZ}))

	// check the /77 and make sure we still get the /65 and root
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 77))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD}))
}

func TestDelete1V6(t *testing.T) {
	tags := make([]GeneratedType, 0)

	matchFunc := func(tagData GeneratedType, val GeneratedType) bool {
		return tagData.(string) == val.(string)
	}

	buf := make([]GeneratedType, 0, 10)
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagZ"

	tree := NewTreeV6()
	assert.Equal(t, 1, tree.countNodes(1))
	tree.Add(patricia.IPv6Address{}, tagZ, nil) // default
	assert.Equal(t, 1, tree.countNodes(1))
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/67", 67), tagA, nil) // 1000000
	assert.Equal(t, 2, tree.countNodes(1))
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/2", 2), tagB, nil) // 10
	assert.Equal(t, 3, tree.countNodes(1))
	tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), tagC, nil)
	assert.Equal(t, 4, tree.countNodes(1))

	// three tags in a hierarchy - ask for an exact match, receive all 3 and the root
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))

	// 1. delete a tag that doesn't exist
	count := 0
	count = tree.DeleteWithBuffer(buf, ipv6FromString("F001:db8:0:0:0:0:2:1/128", 128), matchFunc, "bad tag")
	assert.Equal(t, 0, count)
	assert.Equal(t, 4, tree.countTags(1))
	assert.Equal(t, 4, tree.countNodes(1))

	// 2. delete a tag on an address that exists, but doesn't have the tag
	count = tree.DeleteWithBuffer(buf, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 67), matchFunc, "bad tag")
	assert.Equal(t, 0, count)

	// verify
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))
	assert.Equal(t, 4, tree.countNodes(1))
	assert.Equal(t, 4, tree.countTags(1))

	// 3. delete the default/root tag
	count = tree.DeleteWithBuffer(buf, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 0), matchFunc, "tagZ")
	assert.Equal(t, 1, count)
	assert.Equal(t, 4, tree.countNodes(1)) // doesn't delete anything
	assert.Equal(t, 3, tree.countTags(1))

	// three tags in a hierarchy - ask for an exact match, receive all 3, not the root, which we deleted
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC}))

	// 4. delete tagA
	count = tree.DeleteWithBuffer(buf, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 67), matchFunc, "tagA")
	assert.Equal(t, 1, count)

	// verify
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagB, tagC}))
	assert.Equal(t, 3, tree.countNodes(1))
	assert.Equal(t, 2, tree.countTags(1))

	// 5. delete tag B
	count = tree.DeleteWithBuffer(buf, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 2), matchFunc, "tagB")
	assert.Equal(t, 1, count)

	// verify
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{tagC}))
	assert.Equal(t, 2, tree.countNodes(1))
	assert.Equal(t, 1, tree.countTags(1))

	// 6. delete tag C
	count = tree.DeleteWithBuffer(buf, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), matchFunc, "tagC")
	assert.Equal(t, 1, count)

	// verify
	tags = tags[:0]
	tags = tree.FindTagsAppend(tags, ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128))
	assert.True(t, tagArraysEqual(tags, []string{}))
	assert.Equal(t, 1, tree.countNodes(1))
	assert.Equal(t, 0, tree.countTags(1))
}

// test duplicate tags with no match func
func TestDuplicateTagsWithNoMatchFuncV6(t *testing.T) {
	matchFunc := MatchesFunc(nil)

	tree := NewTreeV6()

	wasAdded, count := tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), "FOO", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	wasAdded, count = tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:2/128", 128), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	// add another at previous node
	wasAdded, count = tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:2/128", 128), "FOOBAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 2, count)

	// add a dupe to the previous node - will be fine since match is nil
	wasAdded, count = tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:2/128", 128), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 3, count)
}

// test duplicate tags with match func that always returns false
func TestDuplicateTagsWithFalseMatchFuncV6(t *testing.T) {
	matchFunc := func(val1 GeneratedType, val2 GeneratedType) bool {
		return false
	}

	tree := NewTreeV6()

	wasAdded, count := tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), "FOO", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	wasAdded, count = tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:2/128", 128), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	// add another at previous node
	wasAdded, count = tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:2/128", 128), "FOOBAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 2, count)

	// add a dupe to the previous node - will be fine since match is nil
	wasAdded, count = tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:2/128", 128), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 3, count)
}

// test duplicate tags with match func that does something
func TestDuplicateTagsWithMatchFuncV6(t *testing.T) {
	matchFunc := func(val1 GeneratedType, val2 GeneratedType) bool {
		return val1 == val2
	}

	tree := NewTreeV6()

	wasAdded, count := tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:1/128", 128), "FOO", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	wasAdded, count = tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:2/128", 128), "BAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 1, count)

	// add another at previous node
	wasAdded, count = tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:2/128", 128), "FOOBAR", matchFunc)
	assert.True(t, wasAdded)
	assert.Equal(t, 2, count)

	// add a dupe to the previous node - will be fine since match is nil
	wasAdded, count = tree.Add(ipv6FromString("2001:db8:0:0:0:0:2:2/128", 128), "BAR", matchFunc)
	assert.False(t, wasAdded)
	assert.Equal(t, 2, count)
}

// test tree traversal
func TestIterateV6(t *testing.T) {
	tree := NewTreeV6()

	// try an empty tree first
	iter := tree.Iterate()
	for iter.Next() {
		assert.Fail(t, "empty tree should not have a next element")
	}

	ipA := ipv6FromString("2001:db8::cb8f:dc00/128", 119)
	ipB := ipv6FromString("2001:db8::cb8f:dcc6/128", 128)
	ipC := ipv6FromString("2001:db8::cb8f:0/128", 112)
	ipD := ipv6FromString("2001:db8::cb8f:dd4b/128", 128)

	// add the 4 addresses
	tree.Add(ipA, "A", nil)
	tree.Add(ipB, "B", nil)
	tree.Add(ipC, "C", nil)
	tree.Add(ipD, "D1", nil)
	tree.Add(ipD, "D2", nil)

	expected := map[string][]string{
		"2001:db8::cb8f:dc00/119": {"A"},
		"2001:db8::cb8f:dcc6/128": {"B"},
		"2001:db8::cb8f:0/112":    {"C"},
		"2001:db8::cb8f:dd4b/128": {"D1", "D2"},
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
