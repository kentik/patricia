package template

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/kentik/patricia"
	"github.com/stretchr/testify/assert"
)

func ipv4FromBytes(bytes []byte, length int) *patricia.IPv4Address {
	return &patricia.IPv4Address{
		Address: binary.BigEndian.Uint32(bytes),
		Length:  uint(length),
	}
}

func BenchmarkFindTags(b *testing.B) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	tree := NewTreeV4(2) // intentionally picking a low number to make sure we resize the underlying collection well

	tree.Add(nil, tagZ) // default
	tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), tagA)
	tree.Add(ipv4FromBytes([]byte{160, 0, 0, 0}, 2), tagB) // 160 -> 128
	tree.Add(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), tagC)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		address := patricia.NewIPv4Address(uint32(2156823809), 32)
		tree.FindTags(&address)
	}
}

func BenchmarkFindDeepestTag(b *testing.B) {
	tree := NewTreeV4(2) // intentionally picking a low number to make sure we resize the underlying collection well
	for i := 32; i > 0; i-- {
		tree.Add(ipv4FromBytes([]byte{127, 0, 0, 1}, i), fmt.Sprintf("Tag-%d", i))
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		address := patricia.NewIPv4Address(uint32(2130706433), 32)
		tree.FindDeepestTag(&address)
	}
}

func BenchmarkBuildTreeAndFindDeepestTag(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree := NewTreeV4(2) // intentionally picking a low number to make sure we resize the underlying collection well

		// populate

		address := patricia.NewIPv4Address(uint32(1653323544), 32)
		tree.Add(&address, "tagA")

		address = patricia.NewIPv4Address(uint32(3334127283), 32)
		tree.Add(&address, "tagB")

		address = patricia.NewIPv4Address(uint32(2540010580), 32)
		tree.Add(&address, "tagC")

		// search

		address = patricia.NewIPv4Address(uint32(1653323544), 32)
		tree.FindDeepestTag(&address)

		address = patricia.NewIPv4Address(uint32(3334127283), 32)
		tree.FindDeepestTag(&address)

		address = patricia.NewIPv4Address(uint32(2540010580), 32)
		tree.FindDeepestTag(&address)
	}
}

func TestSimpleTree1(t *testing.T) {
	tree := NewTreeV4(2) // intentionally picking a low number to make sure we resize the underlying collection well

	ipv4a := ipv4FromBytes([]byte{98, 139, 183, 24}, 32)
	ipv4b := ipv4FromBytes([]byte{198, 186, 190, 179}, 32)
	ipv4c := ipv4FromBytes([]byte{151, 101, 124, 84}, 32)

	tree.Add(ipv4a, "tagA")
	tree.Add(ipv4b, "tagB")
	tree.Add(ipv4c, "tagC")

	found, tag, err := tree.FindDeepestTag(ipv4FromBytes([]byte{98, 139, 183, 24}, 32))
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "tagA", tag)

	found, tag, err = tree.FindDeepestTag(ipv4FromBytes([]byte{198, 186, 190, 179}, 32))
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "tagB", tag)

	found, tag, err = tree.FindDeepestTag(ipv4FromBytes([]byte{151, 101, 124, 84}, 32))
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "tagC", tag)
}

func TestSimpleTree(t *testing.T) {
	tree := NewTreeV4(2) // intentionally picking a low number to make sure we resize the underlying collection well

	for i := 32; i > 0; i-- {
		err := tree.Add(ipv4FromBytes([]byte{127, 0, 0, 1}, i), fmt.Sprintf("Tag-%d", i))
		assert.NoError(t, err)
	}

	tags, err := tree.FindTags(ipv4FromBytes([]byte{127, 0, 0, 1}, 32))
	assert.NoError(t, err)
	if assert.Equal(t, 32, len(tags)) {
		assert.Equal(t, "Tag-32", tags[31].(string))
		assert.Equal(t, "Tag-31", tags[30].(string))
	}

	tags, err = tree.FindTags(ipv4FromBytes([]byte{63, 3, 0, 1}, 32))
	assert.NoError(t, err)
	if assert.Equal(t, 1, len(tags)) {
		assert.Equal(t, "Tag-1", tags[0].(string))
	}

	// find deepest tag: match at lowest level
	found, tag, err := tree.FindDeepestTag(ipv4FromBytes([]byte{127, 0, 0, 1}, 32))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-32", tag.(string))
	}

	// find deepest tag: match at top level
	found, tag, err = tree.FindDeepestTag(ipv4FromBytes([]byte{63, 5, 4, 3}, 32))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-1", tag.(string))
	}

	// find deepest tag: match at mid level
	found, tag, err = tree.FindDeepestTag(ipv4FromBytes([]byte{119, 5, 4, 3}, 32))
	assert.True(t, found)
	if assert.NotNil(t, tag) {
		assert.Equal(t, "Tag-4", tag.(string))
	}

	// find deepest tag: no match
	found, tag, err = tree.FindDeepestTag(ipv4FromBytes([]byte{128, 4, 3, 2}, 32))
	assert.False(t, found)
	assert.Nil(t, tag)

	// Add a couple root tags
	err = tree.Add(ipv4FromBytes([]byte{127, 0, 0, 1}, 0), "root1")
	assert.NoError(t, err)
	err = tree.Add(nil, "root2")
	assert.NoError(t, err)

	tags, err = tree.FindTags(nil)
	assert.NoError(t, err)
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

func TestTree1FindTags(t *testing.T) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	tree := NewTreeV4(2)                                 // intentionally picking a low number to make sure we resize the underlying collection well
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 0), tagZ) // default
	tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), tagA)
	tree.Add(ipv4FromBytes([]byte{160, 0, 0, 0}, 2), tagB) // 160 -> 128
	tree.Add(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), tagC)

	// three tags in a hierarchy - ask for all but the most specific
	tags, err := tree.FindTags(ipv4FromBytes([]byte{128, 142, 133, 1}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagZ}))

	// three tags in a hierarchy - ask for an exact match, receive all 3
	tags, err = tree.FindTags(ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))

	// three tags in a hierarchy - get just the first
	tags, err = tree.FindTags(ipv4FromBytes([]byte{162, 1, 0, 5}, 30))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagB, tagZ}))

	// three tags in hierarchy - get none
	tags, err = tree.FindTags(ipv4FromBytes([]byte{1, 0, 0, 0}, 1))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagZ}))
}

func TestTree1FindTagsWithFilter(t *testing.T) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagD"

	filterFunc := func(val GeneratedType) bool {
		return val == tagA || val == tagB
	}

	tree := NewTreeV4(2)                                 // intentionally picking a low number to make sure we resize the underlying collection well
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 0), tagZ) // default
	tree.Add(ipv4FromBytes([]byte{129, 0, 0, 1}, 7), tagA)
	tree.Add(ipv4FromBytes([]byte{160, 0, 0, 0}, 2), tagB) // 160 -> 128
	tree.Add(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), tagC)

	// three tags in a hierarchy - ask for all but the most specific
	tags, err := tree.FindTagsWithFilter(ipv4FromBytes([]byte{128, 142, 133, 1}, 32), filterFunc)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// three tags in a hierarchy - ask for an exact match, receive all 3
	tags, err = tree.FindTagsWithFilter(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), filterFunc)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// three tags in a hierarchy - get just the first
	tags, err = tree.FindTagsWithFilter(ipv4FromBytes([]byte{162, 1, 0, 5}, 30), filterFunc)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagB}))

	// three tags in hierarchy - get none
	tags, err = tree.FindTagsWithFilter(ipv4FromBytes([]byte{1, 0, 0, 0}, 1), filterFunc)
	assert.NoError(t, err)
	assert.Zero(t, len(tags))
}

// Test that all queries get the root nodes
func TestRootNode(t *testing.T) {
	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagD := "tagD"
	tagZ := "tagE"

	tree := NewTreeV4(2) // intentionally picking a low number to make sure we resize the underlying collection well

	// root node gets tags A & B
	tree.Add(nil, tagA)
	tree.Add(nil, tagB)

	// query the root node with no address
	tags, err := tree.FindTags(nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// query a node that doesn't exist
	tags, err = tree.FindTags(ipv4FromBytes([]byte{1, 2, 3, 4}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB}))

	// create a new /16 node with C & D
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 16), tagC)
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 16), tagD)
	tags, err = tree.FindTags(ipv4FromBytes([]byte{1, 2, 3, 4}, 16))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD}))

	// create a node under the /16 node
	tree.Add(ipv4FromBytes([]byte{1, 2, 3, 4}, 32), tagZ)
	tags, err = tree.FindTags(ipv4FromBytes([]byte{1, 2, 3, 4}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD, tagZ}))

	// check the /24 and make sure we still get the /16 and root
	tags, err = tree.FindTags(ipv4FromBytes([]byte{1, 2, 3, 4}, 24))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagD}))
}

func TestDelete1(t *testing.T) {
	matchFunc := func(tagData GeneratedType, val GeneratedType) bool {
		return tagData.(string) == val.(string)
	}

	tagA := "tagA"
	tagB := "tagB"
	tagC := "tagC"
	tagZ := "tagZ"

	tree := NewTreeV4(2) // intentionally picking a low number to make sure we resize the underlying collection well
	assert.Equal(t, 1, tree.countNodes(1))
	tree.Add(ipv4FromBytes([]byte{8, 7, 6, 5}, 0), tagZ) // default
	assert.Equal(t, 1, tree.countNodes(1))
	assert.Zero(t, len(tree.availableIndexes))
	assert.Equal(t, 2, len(tree.nodes)) // empty first node plus root

	tree.Add(ipv4FromBytes([]byte{128, 3, 0, 5}, 7), tagA) // 1000000
	assert.Equal(t, 2, tree.countNodes(1))
	assert.Equal(t, 3, len(tree.nodes))

	tree.Add(ipv4FromBytes([]byte{128, 5, 1, 1}, 2), tagB) // 10
	assert.Equal(t, 3, tree.countNodes(1))
	assert.Equal(t, 4, len(tree.nodes))

	tree.Add(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), tagC)
	assert.Equal(t, 4, tree.countNodes(1))
	assert.Equal(t, 5, len(tree.nodes))

	// verify status of internal nodes collections
	assert.Zero(t, len(tree.availableIndexes))
	assert.Equal(t, "tagZ", tree.nodes[1].Tags[0])
	assert.Equal(t, "tagA", tree.nodes[2].Tags[0])
	assert.Equal(t, "tagB", tree.nodes[3].Tags[0])
	assert.Equal(t, "tagC", tree.nodes[4].Tags[0])

	// three tags in a hierarchy - ask for an exact match, receive all 3 and the root
	tags, err := tree.FindTags(ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))

	// 1. delete a tag that doesn't exist
	count := 0
	count, err = tree.Delete(ipv4FromBytes([]byte{9, 9, 9, 9}, 32), matchFunc, "bad tag")
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 4, tree.countNodes(1))
	assert.Equal(t, 4, tree.countTags(1))

	// 2. delete a tag on an address that exists, but doesn't have the tag
	count, err = tree.Delete(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), matchFunc, "bad tag")
	assert.Equal(t, 0, count)
	assert.NoError(t, err)

	// verify
	tags, err = tree.FindTags(ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC, tagZ}))
	assert.Equal(t, 4, tree.countNodes(1))
	assert.Equal(t, 4, tree.countTags(1))

	// 3. delete the default/root tag
	count, err = tree.Delete(ipv4FromBytes([]byte{0, 0, 0, 0}, 0), matchFunc, "tagZ")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 4, tree.countNodes(1)) // doesn't delete anything
	assert.Equal(t, 3, tree.countTags(1))
	assert.Equal(t, 0, len(tree.availableIndexes))

	// three tags in a hierarchy - ask for an exact match, receive all 3, not the root, which we deleted
	tags, err = tree.FindTags(ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagA, tagB, tagC}))

	// 4. delete tagA
	count, err = tree.Delete(ipv4FromBytes([]byte{128, 0, 0, 0}, 7), matchFunc, "tagA")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// verify
	tags, err = tree.FindTags(ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagB, tagC}))
	assert.Equal(t, 3, tree.countNodes(1))
	assert.Equal(t, 2, tree.countTags(1))
	assert.Equal(t, 1, len(tree.availableIndexes))
	assert.Equal(t, uint(2), tree.availableIndexes[0])

	// 5. delete tag B
	count, err = tree.Delete(ipv4FromBytes([]byte{128, 0, 0, 0}, 2), matchFunc, "tagB")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// verify
	tags, err = tree.FindTags(ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{tagC}))
	assert.Equal(t, 2, tree.countNodes(1))
	assert.Equal(t, 1, tree.countTags(1))
	assert.Equal(t, 2, len(tree.availableIndexes))
	assert.Equal(t, uint(3), tree.availableIndexes[1])

	// add tagE & tagF to the same node
	tree.Add(ipv4FromBytes([]byte{1, 3, 6, 240}, 32), "tagE")
	tree.Add(ipv4FromBytes([]byte{1, 3, 6, 240}, 32), "tagF")
	assert.Equal(t, 3, tree.countNodes(1))
	assert.Equal(t, 3, tree.countTags(1))

	// this should be recycling tagB
	assert.Equal(t, 1, len(tree.availableIndexes))
	assert.Equal(t, uint(2), tree.availableIndexes[0])
	assert.Equal(t, "tagE", tree.nodes[3].Tags[0])

	// 6. delete tag C
	count, err = tree.Delete(ipv4FromBytes([]byte{128, 3, 6, 240}, 32), matchFunc, "tagC")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// verify
	tags, err = tree.FindTags(ipv4FromBytes([]byte{128, 3, 6, 240}, 32))
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, []string{}))
	assert.Equal(t, 2, tree.countNodes(1))
	assert.Equal(t, 2, tree.countTags(1))
}

func payloadToByteArrays(tags []GeneratedType) [][]byte {
	ret := make([][]byte, 0, len(tags))
	for _, tag := range tags {
		ret = append(ret, tag.([]byte))
	}
	return ret
}
