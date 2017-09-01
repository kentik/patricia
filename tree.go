package patricia

import (
	"fmt"
)

// list of bitmasks to mask all but the indexed bit number, starting from least significant
var _bitMasks = []uint8{128, 64, 32, 16, 8, 4, 2, 1}

// MatchesFunc is called to check if tag data matches the input value
type MatchesFunc func(payload interface{}, val interface{}) bool

// Node represents a node in the Patricia tree
type Node struct {
	Left         *Node   // left node
	Right        *Node   // right node
	Prefix       []uint8 // prefix that this node represents, as a series of bits, stored starting with most significant bit
	PrefixBools  []bool  // prefix, as unpacked bits
	PrefixLength uint8   // how many bits are in the prefix, since it might not line up on a byte boundary
	Tags         []interface{}
}

// Searcher is a thread-UNSAFE interface for searching a tree for tags, reusing cache across calls
type Searcher struct {
	tree  *Tree
	buff1 []bool
}

// FindTags looks through the tree for all tags that match the input address, which includes
// nodes that are an exact or partial match.
// - filterFunc can be nil to return all records. If set, then each entry is run through it - include records that return true
func (s *Searcher) FindTags(address []byte, bitCount uint8, filterFunc func(interface{}) bool) ([]interface{}, error) {
	s.buff1 = s.buff1[:0]

	return s.tree.findTags(s, address, bitCount, filterFunc)
}

// NewNode returns a new node
func NewNode(prefix []uint8, prefixLength uint8) *Node {
	prefixBools := make([]bool, prefixLength)
	unpackBits(&prefixBools, prefix, prefixLength)
	return &Node{
		Prefix:       prefix,
		PrefixBools:  prefixBools,
		PrefixLength: prefixLength,
	}
}

func (n *Node) updateBools() {
	n.PrefixBools = make([]bool, 0, n.PrefixLength)
	unpackBits(&n.PrefixBools, n.Prefix, n.PrefixLength)
}

// Tree is an IP Address patricia tree
type Tree struct {
	addressSize uint8 // number of bytes per address
	root        *Node
}

// NewTree returns a new Tree
func NewTree(addressSize uint8) (*Tree, error) {
	if addressSize > 255 {
		return nil, fmt.Errorf("Max address size is 255") // can't go over 255 because we rely on uint8
	}

	return &Tree{
		addressSize: addressSize,
		root:        &Node{},
	}, nil
}

// GetSearcher returns a TreeSearcher for performant sequential searches.
// - each thread/goroutine needs its own TreeSearcher
func (t *Tree) GetSearcher() *Searcher {
	return &Searcher{
		tree:  t,
		buff1: make([]bool, 0, t.addressSize),
	}
}

func countNodes(node *Node) int {
	if node == nil {
		return 0
	}
	return 1 + countNodes(node.Left) + countNodes(node.Right)
}

func (t *Tree) countTags(node *Node) int {
	if node == nil {
		return 0
	}
	return len(node.Tags) + t.countTags(node.Left) + t.countTags(node.Right)
}

// NodeCount returns the number of nodes in the tree
func (t *Tree) NodeCount() int {
	return countNodes(t.root)
}

// TagCount returns the number of tags in the tree
func (t *Tree) TagCount() int {
	return t.countTags(t.root)
}

// FindTags looks through the tree for all tags that match the input address, which includes
// nodes that are an exact or partial match.
// - filterFunc can be nil to return all records. If set, then each entry is run through it - include records that return true
func (t *Tree) FindTags(address []byte, bitCount uint8, filterFunc func(interface{}) bool) ([]interface{}, error) {
	searcher := t.GetSearcher()
	return t.findTags(searcher, address, bitCount, filterFunc)
}

func (t *Tree) findTags(searcher *Searcher, address []byte, bitCount uint8, filterFunc func(interface{}) bool) ([]interface{}, error) {
	if bitCount > t.addressSize {
		return nil, fmt.Errorf("bitcount of %d is too long for the tree, which has an address size of %d bits", bitCount, t.addressSize)
	}
	if bitCount > uint8(len(address)*8) {
		return nil, fmt.Errorf("not enough bits (%d) for the input bitcount %d", len(address)*8, bitCount)
	}

	// add common tags
	ret := make([]interface{}, 0)
	filterAndAppend := func(tags []interface{}) {
		if filterFunc == nil {
			ret = append(ret, tags...)
		} else {
			for _, tag := range tags {
				if filterFunc(tag) {
					ret = append(ret, tag)
				}
			}
		}
	}

	if t.root.Tags != nil && len(t.root.Tags) > 0 {
		filterAndAppend(t.root.Tags)
	}

	if address == nil {
		// caller just looking for root tags
		return ret, nil
	}

	// unpack the binary address into an array of booleans, reusing searcher.buff1 slice
	unpackBits(&searcher.buff1, address, bitCount)

	// find the starting point
	var node *Node
	if searcher.buff1[0] == false {
		node = t.root.Left
	} else {
		node = t.root.Right
	}

	// restore buff1 when done, since we chop off the left side of it while searching
	buff1Original := searcher.buff1
	defer func() {
		searcher.buff1 = buff1Original
	}()

	// go down the tree
	for {
		if node == nil {
			return ret, nil
		}

		matchCount := countMatches(searcher.buff1, node.PrefixBools)
		if matchCount < node.PrefixLength {
			// didn't match all of the node's prefixes - stop traversing
			return ret, nil
		}

		// matched the full node - get its tags, and continue on
		if node.Tags != nil && len(node.Tags) > 0 {
			filterAndAppend(node.Tags)
		}

		if matchCount == uint8(len(searcher.buff1)) {
			// exact match - we're done
			return ret, nil
		}

		searcher.buff1 = searcher.buff1[matchCount:] // chop off the bits we've already matched
		if searcher.buff1[0] == false {
			node = node.Left
		} else {
			node = node.Right
		}
	}
}

// Delete a tag from the tree if it matches matchVal as determined by matchFunc. Return how many tags are removed.
func (t *Tree) Delete(address []byte, bitCount uint8, matchFunc MatchesFunc, matchVal interface{}) (int, error) {
	deleteCount := 0

	if bitCount > t.addressSize {
		return 0, fmt.Errorf("bitcount of %d is too long for the tree, which has an address size of %d bits", bitCount, t.addressSize)
	}
	if bitCount > uint8(len(address)*8) {
		return 0, fmt.Errorf("not enough bits (%d) for the input bitcount %d", len(address)*8, bitCount)
	}

	parents := []*Node{t.root}

	// Find the node that would have this address
	addressBits := make([]bool, 0, t.addressSize)

	var targetNode *Node
	if bitCount == 0 {
		targetNode = t.root
	} else {
		unpackBits(&addressBits, address, bitCount)

		// find the starting point
		var node *Node
		if addressBits[0] == false {
			node = t.root.Left
		} else {
			node = t.root.Right
		}

		// go down the tree
		for {
			if node == nil {
				return 0, nil
			}

			matchCount := countMatches(addressBits, node.PrefixBools)
			if matchCount < node.PrefixLength {
				// didn't match all of the node's prefixes - stop traversing
				return 0, nil
			}

			if matchCount == uint8(len(addressBits)) {
				// exact match - we're done
				targetNode = node
				break
			}

			addressBits = addressBits[matchCount:] // chop off the bits we've already matched
			parents = append(parents, node)
			if addressBits[0] == false {
				node = node.Left
			} else {
				node = node.Right
			}
		}
	}

	if targetNode == nil || len(targetNode.Tags) == 0 {
		// couldn't find the node or no tags at this node - nothing to do
		return 0, nil
	}

	// found the node, it has tags - look for the tag we're searching for
	matchIndices := make(map[int]bool) // just in case, delete multiple instances of the tag
	for index, tagData := range targetNode.Tags {
		if matchFunc(tagData, matchVal) {
			matchIndices[index] = true
			deleteCount++
		}
	}

	if len(matchIndices) == 0 {
		// node exists, but tag does not
		return 0, nil
	}

	newListLength := len(targetNode.Tags) - len(matchIndices)
	newTagList := make([]interface{}, 0, newListLength)
	for index, tagData := range targetNode.Tags {
		if _, ok := matchIndices[index]; !ok {
			newTagList = append(newTagList, tagData)
		}
	}
	targetNode.Tags = newTagList

	// See if we now have an empty node that's not the root (root needs to have no prefix)
	if newListLength == 0 && targetNode != t.root {
		// see if this empty node is a leaf that we can delete
		if targetNode.Left == nil && targetNode.Right == nil {
			// empty leaf - unhook this node from the tree
			parent := parents[len(parents)-1]
			if parent.Left == targetNode {
				parent.Left = nil
			} else if parent.Right == targetNode {
				parent.Right = nil
			} else {
				panic("Trying to delete an empty leaf, and its parent isn't wired up properly.")
			}

			// walk back up the tree, removing any other parents that are also empty leaves
			for i := len(parents) - 1; i > 0; i-- {
				if parents[i].Left == nil && parents[i].Right == nil && len(parents[i].Tags) == 0 {
					// this is also an empty leaf - unhook it
					if parents[i-1].Left == parents[i] {
						parents[i-1].Left = nil
					} else if parents[i-1].Right == parents[i] {
						parents[i-1].Right = nil
					} else {
						panic("Trying to delete an empty leaf while walking up the parent list, and its parent isn't wired up properly.")
					}
				}
			}

			return deleteCount, nil
		}

		// not a leaf - if we only have one child, then move the child up
		if targetNode.Left != nil && targetNode.Right != nil {
			// both children are set - can't do anything - this node will stay empty
			return deleteCount, nil
		}

		// we have only one child - move it up

		// can get rid of a node - get rid of the child, since we have its parent now
		childNode := targetNode.Left
		if childNode == nil {
			childNode = targetNode.Right
		}

		// create the new, combined prefix
		targetNode.PrefixBools = append(targetNode.PrefixBools, childNode.PrefixBools...)
		targetNode.Prefix = packBits(targetNode.PrefixBools)
		targetNode.PrefixLength = uint8(len(targetNode.PrefixBools))

		targetNode.Left = childNode.Left
		targetNode.Right = childNode.Right
		targetNode.Tags = childNode.Tags
	}

	return deleteCount, nil
}

// Add adds an address to the tree with the given bit count
func (t *Tree) Add(address []byte, bitCount uint8, tag interface{}) error {
	if bitCount > t.addressSize {
		return fmt.Errorf("bitcount of %d is too long for the tree, which has an address size of %d bits", bitCount, t.addressSize)
	}
	if bitCount > uint8(len(address)*8) {
		return fmt.Errorf("not enough bits (%d) for the input bitcount %d", len(address)*8, bitCount)
	}

	// handle the root tags
	if bitCount == 0 {
		// long-lived memory: don't let append resize by doubling
		if cap(t.root.Tags) < len(t.root.Tags)+1 {
			old := t.root.Tags
			t.root.Tags = make([]interface{}, len(t.root.Tags), len(t.root.Tags)+1)
			copy(t.root.Tags, old)
		}
		t.root.Tags = append(t.root.Tags, tag)
		return nil
	}

	addressBits := make([]bool, 0, t.addressSize)
	thisNodeBits := make([]bool, 0, t.addressSize)

	unpackBits(&addressBits, address, bitCount)

	// root node doesn't have any prefix, so find the starting point
	var node *Node
	parent := t.root
	if addressBits[0] == false {
		if t.root.Left == nil {
			// one of the base cases
			t.root.Left = NewNode(address, bitCount)
			t.root.Left.Tags = []interface{}{tag}

			return nil
		}
		node = t.root.Left
	} else {
		if t.root.Right == nil {
			// one of the base cases
			t.root.Right = NewNode(address, bitCount)
			t.root.Right.Tags = []interface{}{tag}

			return nil
		}
		node = t.root.Right
	}

	for {
		// this node has a prefix - should share some of if with the remaining bits of this with this node

		// see how many bits match this node's prefix
		unpackBits(&thisNodeBits, node.Prefix, node.PrefixLength)
		if len(thisNodeBits) == 0 {
			panic("Navigated to a node with no prefix")
		}

		matchCount := countMatches(addressBits, thisNodeBits)
		if matchCount == uint8(len(addressBits)) {
			// all bits in the input address matched

			if matchCount == uint8(len(thisNodeBits)) {
				// the whole prefix matched - we're done!

				// long-lived memory: don't let append resize by doubling
				if cap(node.Tags) < len(node.Tags)+1 {
					old := node.Tags
					node.Tags = make([]interface{}, len(node.Tags), len(node.Tags)+1)
					copy(node.Tags, old)
				}
				node.Tags = append(node.Tags, tag)

				return nil
			}

			// the input address is shorter than the match found - need to create a new parent
			newNode := NewNode(packBits(addressBits), matchCount)
			newNode.Tags = []interface{}{tag}

			node.Prefix = packBits(thisNodeBits[matchCount:]) // remove the prefix that matched
			node.PrefixLength -= matchCount
			node.updateBools()
			if thisNodeBits[matchCount] == false {
				newNode.Left = node
			} else {
				newNode.Right = node
			}

			// now determine where the new node belongs
			if parent.Left == node {
				parent.Left = newNode
			} else {
				if parent.Right != node {
					panic("node isn't left or right parent - should be impossible")
				}
				parent.Right = newNode
			}
			return nil
		}

		if matchCount == 0 {
			// no match at all - should not happen
			panic("Should not have traversed to a node with no prefix match")
		}

		if matchCount == node.PrefixLength {
			// throw away the matched bits
			addressBits = addressBits[matchCount:] // chop off the bits we've matched so far

			// match all of this node, but not all of the input address - continue on
			if !addressBits[0] {
				if node.Left == nil {
					// nothing to the left - create a node there
					node.Left = NewNode(packBits(addressBits), uint8(len(addressBits)))
					node.Left.Tags = []interface{}{tag}

					return nil
				}

				// there's a node to the left - iterate further
				parent = node
				node = node.Left
				continue
			}

			// node didn't belong on left - so, belongs on right
			if node.Right == nil {
				// nothing on the right - create a node there
				node.Right = NewNode(packBits(addressBits), uint8(len(addressBits)))
				node.Right.Tags = []interface{}{tag}

				return nil
			}

			// there's a node to the right - iterate further
			parent = node
			node = node.Right
			continue
		}

		// partial match with this node - need to create a new parent with the matched prefix
		newNode := NewNode(packBits(addressBits[:matchCount]), matchCount)

		// see where we're moving this node to, based on the next bit after the match
		node.Prefix = packBits(thisNodeBits[matchCount:]) // remove the prefix that matched
		node.PrefixLength -= matchCount
		node.updateBools()

		// create the new node for the match
		matchedNode := NewNode(packBits(addressBits[matchCount:]), uint8(len(addressBits))-matchCount)
		matchedNode.Tags = []interface{}{tag}

		if !thisNodeBits[matchCount] {
			// after the prefix, this node had a zero, so it belongs left
			newNode.Left = node

			// since the input address didn't match past this prefix, must have 1 and belong right
			if addressBits[matchCount] == false {
				panic("after the match, expected a true")
			}
			newNode.Right = matchedNode
		} else {
			// after the prefix, this node had a 1, so it belongs right
			newNode.Right = node

			// since the input address didm't match past this prefix, must have 0 and belong left
			if addressBits[matchCount] == true {
				panic("after the match, expected a false")
			}
			newNode.Left = matchedNode
		}

		// now determine where the new node belongs
		if parent.Left == node {
			parent.Left = newNode
		} else {
			if parent.Right != node {
				panic("node isn't left or right parent - should be impossible")
			}
			parent.Right = newNode
		}
		return nil
	}
}

func countMatches(first []bool, second []bool) uint8 {
	count := len(first)
	lenSecond := len(second)
	if lenSecond < count {
		count = lenSecond
	}

	matchCount := uint8(0)
	for i := 0; i < count; i++ {
		if first[i] != second[i] {
			return matchCount
		}
		matchCount++
	}
	return matchCount
}

func unpackBits(bools *[]bool, compressedBits []byte, length uint8) {
	*bools = (*bools)[:length]

	// loop each byte
	next := uint8(0)
	if next == length {
		return
	}

	for _, thisByte := range compressedBits {
		(*bools)[next] = (thisByte & 128) != 0
		next++
		if next == length {
			return
		}

		(*bools)[next] = (thisByte & 64) != 0
		next++
		if next == length {
			return
		}

		(*bools)[next] = (thisByte & 32) != 0
		next++
		if next == length {
			return
		}

		(*bools)[next] = (thisByte & 16) != 0
		next++
		if next == length {
			return
		}

		(*bools)[next] = (thisByte & 8) != 0
		next++
		if next == length {
			return
		}

		(*bools)[next] = (thisByte & 4) != 0
		next++
		if next == length {
			return
		}

		(*bools)[next] = (thisByte & 2) != 0
		next++
		if next == length {
			return
		}

		(*bools)[next] = (thisByte & 1) != 0
		next++
		if next == length {
			return
		}
	}
}

// Convert an array of bools into an array of bytes
func packBits(bits []bool) []byte {
	bitCount := len(bits)
	byteCount := bitCount / 8
	lastByteBitCount := bitCount % 8
	if lastByteBitCount != 0 {
		byteCount++
	} else {
		lastByteBitCount = 8
	}
	ret := make([]byte, byteCount)

	// loop each output byte
	for thisByte := 0; thisByte < byteCount; thisByte++ {
		bitsThisByte := 8
		if thisByte == byteCount-1 {
			bitsThisByte = lastByteBitCount
		}

		startBit := thisByte * 8
		bitmaskBit := 0
		for thisBit := startBit; thisBit < startBit+bitsThisByte; thisBit++ {
			if bits[thisBit] {
				ret[thisByte] |= _bitMasks[bitmaskBit]
			}
			bitmaskBit++
		}
	}
	return ret
}
