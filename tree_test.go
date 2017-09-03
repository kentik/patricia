package patricia

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleTree(t *testing.T) {
	tree, err := NewTree(32)
	assert.NoError(t, err)

	for i := 32; i > 0; i-- {
		err = tree.Add([]byte{127, 0, 0, 1}, uint8(i), fmt.Sprintf("Tag-%d", i))
		assert.NoError(t, err)
	}

	tags, err := tree.FindTags([]byte{127, 0, 0, 1}, 32, nil)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(tags))
	assert.Equal(t, "Tag-32", tags[31].(string))
	assert.Equal(t, "Tag-31", tags[30].(string))
	assert.Equal(t, "Tag-1", tags[0].(string))
}

func TestUnpackBits(t *testing.T) {
	// test #1 - full two bytes
	// 10110101 00110101
	data := []byte{
		byte(181),
		byte(53),
	}

	bits := make([]bool, 32)

	unpackBits(&bits, data, 16)
	assert.Equal(t, 16, len(bits), "expected 16 bools")

	assert.True(t, bits[0])
	assert.False(t, bits[1])
	assert.True(t, bits[2])
	assert.True(t, bits[3])
	assert.False(t, bits[4])
	assert.True(t, bits[5])
	assert.False(t, bits[6])
	assert.True(t, bits[7])

	assert.False(t, bits[8])
	assert.False(t, bits[9])
	assert.True(t, bits[10])
	assert.True(t, bits[11])
	assert.False(t, bits[12])
	assert.True(t, bits[13])
	assert.False(t, bits[14])
	assert.True(t, bits[15])

	// test #2 - partial second byte
	unpackBits(&bits, data, 15)
	assert.Equal(t, 15, len(bits), "expected 15 bools")
	assert.True(t, bits[0])
	assert.False(t, bits[1])
	assert.True(t, bits[2])
	assert.True(t, bits[3])
	assert.False(t, bits[4])
	assert.True(t, bits[5])
	assert.False(t, bits[6])
	assert.True(t, bits[7])

	assert.False(t, bits[8])
	assert.False(t, bits[9])
	assert.True(t, bits[10])
	assert.True(t, bits[11])
	assert.False(t, bits[12])
	assert.True(t, bits[13])
	assert.False(t, bits[14])

	// test #3 - more bits than we need are passed in
	unpackBits(&bits, data, 3)
	assert.Equal(t, 3, len(bits))
}

func BenchmarkUnpackBits(b *testing.B) {
	test := []byte{byte(5), byte(6), byte(7), byte(8)} // 32-bit v4 IP address
	bits := make([]bool, 32)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		unpackBits(&bits, test, 1)
		unpackBits(&bits, test, 2)
		unpackBits(&bits, test, 3)
		unpackBits(&bits, test, 4)
		unpackBits(&bits, test, 5)
		unpackBits(&bits, test, 6)
		unpackBits(&bits, test, 7)
		unpackBits(&bits, test, 8)
	}
}

func TestPackBits(t *testing.T) {
	// test #1 - two full bytes worth of data
	bits := []bool{
		true, false, true, true, false, true, false, true,
		false, false, true, true, false, true, false, true,
	}

	packed := packBits(bits)
	assert.Equal(t, 2, len(packed))
	assert.Equal(t, byte(181), packed[0])
	assert.Equal(t, byte(53), packed[1])

	// test #2 - partial second byte of data
	bits = bits[0:15]
	packed = packBits(bits)
	assert.Equal(t, 2, len(packed))
	assert.Equal(t, byte(181), packed[0])
	assert.Equal(t, byte(52), packed[1])
}

func TestCountMatches(t *testing.T) {
	a := []bool{true, true, true, false, true, true}
	b := []bool{true, true, false, false, true, true}
	assert.Equal(t, byte(2), countMatches(a, b))
	assert.Equal(t, byte(2), countMatches(b, a))

	a = []bool{true, true, true, false, true, true}
	b = []bool{true, true, true, false, true, true}
	assert.Equal(t, byte(6), countMatches(a, b))
	assert.Equal(t, byte(6), countMatches(b, a))
}

// assert two byte arrays are equal
func tagsEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// assert two collections of arrays have the same tags - don't worry about performance
func tagArraysEqual(a []interface{}, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		found := false
		for j := 0; j < len(b); j++ {
			if tagsEqual(a[i].([]byte), b[j]) {
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

func TestTree1(t *testing.T) {
	tagA := []byte{1, 2, 3}
	tagB := []byte{2, 3, 4}
	tagC := []byte{3, 4, 5}
	tagZ := []byte{4, 5, 6}

	tree, err := NewTree(32)
	assert.NoError(t, err)
	tree.Add([]byte{}, 0, tagZ) // default
	tree.Add([]byte{129, 0, 0, 1}, 7, tagA)
	tree.Add([]byte{160, 0, 0, 0}, 2, tagB) // 160 -> 128
	tree.Add([]byte{128, 3, 6, 240}, 32, tagC)

	// three tags in a hierarchy - ask for all but the most specific
	tags, err := tree.FindTags([]byte{128, 142, 133, 1}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB, tagZ}))

	// three tags in a hierarchy - ask for an exact match, receive all 3
	tags, err = tree.FindTags([]byte{128, 3, 6, 240}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB, tagC, tagZ}))

	// three tags in a hierarchy - get just the first
	tags, err = tree.FindTags([]byte{162, 1, 0, 5}, 30, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagB, tagZ}))

	// three tags in hierarchy - get none
	tags, err = tree.FindTags([]byte{1, 0, 0, 0}, 1, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagZ}))
}

// BenchmarkTestTree tests searching the tree directly, without a searcher
func BenchmarkTestTree(b *testing.B) {
	tagA := []byte{1, 2, 3}
	tagB := []byte{2, 3, 4}
	tagC := []byte{3, 4, 5}
	tagZ := []byte{4, 5, 6}

	tree, _ := NewTree(32)

	tree.Add([]byte{}, 0, tagZ) // default
	tree.Add([]byte{129, 0, 0, 1}, 7, tagA)
	tree.Add([]byte{160, 0, 0, 0}, 2, tagB) // 160 -> 128
	tree.Add([]byte{128, 3, 6, 240}, 32, tagC)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree.FindTags([]byte{128, 142, 133, 1}, 32, nil)
	}
}

// BenchmarkTestTreeSearcher tests searching the tree through a Searcher
func BenchmarkTestTreeSearcher(b *testing.B) {
	tagA := []byte{1, 2, 3}
	tagB := []byte{2, 3, 4}
	tagC := []byte{3, 4, 5}
	tagZ := []byte{4, 5, 6}

	tree, _ := NewTree(32)
	searcher := tree.GetSearcher()

	tree.Add([]byte{}, 0, tagZ) // default
	tree.Add([]byte{129, 0, 0, 1}, 7, tagA)
	tree.Add([]byte{160, 0, 0, 0}, 2, tagB) // 160 -> 128
	tree.Add([]byte{128, 3, 6, 240}, 32, tagC)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		searcher.FindTags([]byte{128, 142, 133, 1}, 32, nil)
	}
}

// Test that all queries get the root nodes
func TestRootNode(t *testing.T) {
	tagA := []byte{1, 2, 3}
	tagB := []byte{2, 3, 4}
	tagC := []byte{3, 4, 5}
	tagD := []byte{3, 4, 5}
	tagZ := []byte{4, 5, 6}

	tree, err := NewTree(32)
	assert.NoError(t, err)

	// root node gets tags A & B
	tree.Add(nil, 0, tagA)
	tree.Add(nil, 0, tagB)

	// query the root node with no address
	tags, err := tree.FindTags(nil, 0, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB}))

	// query a node that doesn't exist
	tags, err = tree.FindTags([]byte{1, 2, 3, 4}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB}))

	// create a new /16 node with C & D
	tree.Add([]byte{1, 2}, 16, tagC)
	tree.Add([]byte{1, 2}, 16, tagD)
	tags, err = tree.FindTags([]byte{1, 2}, 16, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB, tagC, tagD}))

	// create a node under the /16 node
	tree.Add([]byte{1, 2, 3, 4}, 32, tagZ)
	tags, err = tree.FindTags([]byte{1, 2, 3, 4}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB, tagC, tagD, tagZ}))

	// check the /24 and make sure we still get the /16 and root
	tags, err = tree.FindTags([]byte{1, 2, 3}, 24, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB, tagC, tagD}))
}

func TestDelete1(t *testing.T) {
	matchFunc := func(tagData interface{}, val interface{}) bool {
		return uint8(tagData.([]byte)[0]) == val.(uint8)
	}

	tagA := []byte{1, 2, 3}
	tagB := []byte{2, 3, 4}
	tagC := []byte{3, 4, 5}
	tagZ := []byte{4, 5, 6}

	tree, err := NewTree(32)
	assert.Equal(t, 1, countNodes(tree.root))
	assert.NoError(t, err)
	tree.Add([]byte{}, 0, tagZ) // default
	assert.Equal(t, 1, countNodes(tree.root))
	tree.Add([]byte{129, 0, 0, 1}, 7, tagA)
	assert.Equal(t, 2, countNodes(tree.root))
	tree.Add([]byte{160, 0, 0, 0}, 2, tagB) // 160/2 -> 128
	assert.Equal(t, 3, countNodes(tree.root))
	tree.Add([]byte{128, 3, 6, 240}, 32, tagC)
	assert.Equal(t, 4, countNodes(tree.root))
	assert.Equal(t, 4, tree.countTags(tree.root))

	// three tags in a hierarchy - ask for an exact match, receive all 3 and the root
	tags, err := tree.FindTags([]byte{128, 3, 6, 240}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB, tagC, tagZ}))

	// 1. delete a tag that doesn't exist
	count := 0
	count, err = tree.Delete([]byte{9, 9, 9, 9}, 32, matchFunc, uint8(9))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 4, countNodes(tree.root))
	assert.Equal(t, 4, tree.countTags(tree.root))

	// 2. delete a tag on an address that exists, but doesn't have the tag
	count, err = tree.Delete([]byte{128, 3, 6, 240}, 32, matchFunc, uint8(0))
	assert.Equal(t, 0, count)
	assert.NoError(t, err)

	// verify
	tags, err = tree.FindTags([]byte{128, 3, 6, 240}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB, tagC, tagZ}))
	assert.Equal(t, 4, countNodes(tree.root))
	assert.Equal(t, 4, tree.countTags(tree.root))

	// 3. delete the default/root tag
	count, err = tree.Delete([]byte{}, 0, matchFunc, uint8(4))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 4, countNodes(tree.root)) // doesn't delete anything
	assert.Equal(t, 3, tree.countTags(tree.root))

	// three tags in a hierarchy - ask for an exact match, receive all 3, not the root, which we deleted
	tags, err = tree.FindTags([]byte{128, 3, 6, 240}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagA, tagB, tagC}))

	// 4. delete tagA
	count, err = tree.Delete([]byte{128}, 7, matchFunc, uint8(1))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// verify
	tags, err = tree.FindTags([]byte{128, 3, 6, 240}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagB, tagC}))
	assert.Equal(t, 3, countNodes(tree.root))
	assert.Equal(t, 2, tree.countTags(tree.root))

	// 5. delete tag B
	count, err = tree.Delete([]byte{128}, 2, matchFunc, uint8(2))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// verify
	tags, err = tree.FindTags([]byte{128, 3, 6, 240}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{tagC}))
	assert.Equal(t, 2, countNodes(tree.root))
	assert.Equal(t, 1, tree.countTags(tree.root))

	// 6. delete tag C
	count, err = tree.Delete([]byte{128, 3, 6, 240}, 32, matchFunc, uint8(3))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// verify
	tags, err = tree.FindTags([]byte{128, 3, 6, 240}, 32, nil)
	assert.NoError(t, err)
	assert.True(t, tagArraysEqual(tags, [][]byte{}))
	assert.Equal(t, 1, countNodes(tree.root))
	assert.Equal(t, 0, tree.countTags(tree.root))
}

func payloadToByteArrays(tags []interface{}) [][]byte {
	ret := make([][]byte, 0, len(tags))
	for _, tag := range tags {
		ret = append(ret, tag.([]byte))
	}
	return ret
}
