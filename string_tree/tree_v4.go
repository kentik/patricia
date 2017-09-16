package string_tree

import (
	"github.com/kentik/patricia"
)

const _leftmost32Bit = uint32(1 << 31)

// MatchesFunc is called to check if tag data matches the input value
type MatchesFunc func(payload string, val string) bool

// FilterFunc is called on each result to see if it belongs in the resulting set
type FilterFunc func(payload string) bool

// Tree is an IP Address patricia tree
type TreeV4 struct {
	nodes            []treeNodeV4 // root is always at [1] - [0] is unused
	availableIndexes []uint       // a place to store node indexes that we deleted, and are available
	tags             map[uint64]string
}

// NewTree returns a new Tree
func NewTreeV4(startingCapacity uint) *TreeV4 {
	return &TreeV4{
		nodes:            make([]treeNodeV4, 2, startingCapacity), // index 0 is skipped, 1 is root
		availableIndexes: make([]uint, 0),
		tags:             make(map[uint64]string),
	}
}

func (t *TreeV4) addTag(tag string, nodeIndex uint) {
	tagCount := t.nodes[nodeIndex].TagCount
	t.tags[(uint64(nodeIndex)<<32)+(uint64(tagCount))] = tag
	t.nodes[nodeIndex].TagCount++
}

func (t *TreeV4) tagsForNode(nodeIndex uint) []string {
	// TODO: clean up the typing in here, between uint, uint64
	tagCount := t.nodes[nodeIndex].TagCount
	ret := make([]string, tagCount)
	key := uint64(nodeIndex) << 32
	for i := uint(0); i < tagCount; i++ {
		ret[i] = t.tags[key+uint64(i)]
	}
	return ret
}

func (t *TreeV4) firstTagForNode(nodeIndex uint) string {
	return t.tags[(uint64(nodeIndex) << 32)]
}

// delete tags at the input node, returning how many were deleted, and how many are left
func (t *TreeV4) deleteTag(nodeIndex uint, matchTag string, matchFunc MatchesFunc) (int, int) {
	// WBCTODO: write this properly - this is a hacky get-this-done implementation

	// get tags
	tags := t.tagsForNode(nodeIndex)

	// delete tags
	for i := uint(0); i < t.nodes[nodeIndex].TagCount; i++ {
		delete(t.tags, (uint64(nodeIndex)<<32)+uint64(i))
	}
	t.nodes[nodeIndex].TagCount = 0

	// put them back
	deleteCount := 0
	keepCount := 0
	for _, tag := range tags {
		if matchFunc(tag, matchTag) {
			deleteCount++
		} else {
			// doesn't match - get to keep it
			t.addTag(tag, nodeIndex)
			keepCount++
		}
	}
	return deleteCount, keepCount
}

// create a new node in the tree, return its index
func (t *TreeV4) newNode(prefix uint32, prefixLength uint) uint {
	availCount := len(t.availableIndexes)
	if availCount > 0 {
		index := t.availableIndexes[availCount-1]
		t.availableIndexes = t.availableIndexes[:availCount-1]
		t.nodes[index] = treeNodeV4{prefix: prefix, prefixLength: prefixLength}
		return index
	}

	t.nodes = append(t.nodes, treeNodeV4{prefix: prefix, prefixLength: prefixLength})
	return uint(len(t.nodes) - 1)
}

// Add adds a node to the tree
func (t *TreeV4) Add(address *patricia.IPv4Address, tag string) error {
	// make sure we have more than enough capacity before we start adding to the tree, which invalidates pointers into the array
	if cap(t.nodes) < (len(t.nodes) + 10) {
		temp := make([]treeNodeV4, len(t.nodes), (cap(t.nodes)+1)*2)
		copy(temp, t.nodes)
		t.nodes = temp
	}

	root := &t.nodes[1]

	// handle root tags
	if address == nil || address.Length == 0 {
		t.addTag(tag, 1)
		return nil
	}

	// root node doesn't have any prefix, so find the starting point
	nodeIndex := uint(0)
	parent := root
	if address.Address < _leftmost32Bit {
		if root.Left == 0 {
			newNodeIndex := t.newNode(address.Address, address.Length)
			t.addTag(tag, newNodeIndex)
			root.Left = newNodeIndex
			return nil
		}
		nodeIndex = root.Left
	} else {
		if root.Right == 0 {
			newNodeIndex := t.newNode(address.Address, address.Length)
			t.addTag(tag, newNodeIndex)
			root.Right = newNodeIndex
			return nil
		}
		nodeIndex = root.Right
	}

	for {
		node := &t.nodes[nodeIndex]
		matchCount := uint(node.MatchCount(address))
		if matchCount == address.Length {
			// all the bits in the address matched

			if matchCount == node.prefixLength {
				// the whole prefix matched - we're done!
				t.addTag(tag, nodeIndex)
				return nil
			}

			// the input address is shorter than the match found - need to create a new, intermediate parent
			newNodeIndex := t.newNode(address.Address, address.Length)
			newNode := &t.nodes[newNodeIndex]
			t.addTag(tag, newNodeIndex)

			// the existing node loses those matching bits, and becomes a child of the new node

			// shift
			node.ShiftPrefix(matchCount)

			if node.prefix < _leftmost32Bit {
				newNode.Left = nodeIndex
			} else {
				newNode.Right = nodeIndex
			}

			// now give this new node a home
			if parent.Left == nodeIndex {
				parent.Left = newNodeIndex
			} else {
				if parent.Right != nodeIndex {
					panic("node isn't left or right parent - should be impossible! (1)")
				}
				parent.Right = newNodeIndex
			}
			return nil
		}

		if matchCount == 0 {
			panic("Should not have traversed to a node with no prefix match")
		}

		if matchCount == node.prefixLength {
			// partial match - we have to keep traversing

			// chop off what's matched so far
			address.ShiftLeft(matchCount)

			if address.Address < _leftmost32Bit {
				if node.Left == 0 {
					// nowhere else to go - create a new node here
					newNodeIndex := t.newNode(address.Address, address.Length)
					t.addTag(tag, newNodeIndex)
					node.Left = newNodeIndex
					return nil
				}

				// there's a node to the left - traverse it
				parent = node
				nodeIndex = node.Left
				continue
			}

			// node didn't belong on the left, so it belongs on the right
			if node.Right == 0 {
				// nowhere else to go - create a new node here
				newNodeIndex := t.newNode(address.Address, address.Length)
				t.addTag(tag, newNodeIndex)
				node.Right = newNodeIndex
				return nil
			}

			// there's a node to the right - traverse it
			parent = node
			nodeIndex = node.Right
			continue
		}

		// partial match with this node - need to split this node
		newCommonParentNodeIndex := t.newNode(address.Address, matchCount)
		newCommonParentNode := &t.nodes[newCommonParentNodeIndex]

		// shift
		address.ShiftLeft(matchCount)

		newNodeIndex := t.newNode(address.Address, address.Length)
		t.addTag(tag, newNodeIndex)

		// see where the existing node fits - left or right
		node.ShiftPrefix(matchCount)
		if node.prefix < _leftmost32Bit {
			newCommonParentNode.Left = nodeIndex
			newCommonParentNode.Right = newNodeIndex
		} else {
			newCommonParentNode.Right = nodeIndex
			newCommonParentNode.Left = newNodeIndex
		}

		// now determine where the new node belongs
		if parent.Left == nodeIndex {
			parent.Left = newCommonParentNodeIndex
		} else {
			if parent.Right != nodeIndex {
				panic("node isn't left or right parent - should be impossible! (2)")
			}
			parent.Right = newCommonParentNodeIndex
		}
		return nil
	}
}

// Delete a tag from the tree if it matches matchVal, as determined by matchFunc. Returns how many tags are removed
func (t *TreeV4) Delete(address *patricia.IPv4Address, matchFunc MatchesFunc, matchVal string) (int, error) {
	// traverse the tree, finding the node and its parent
	root := &t.nodes[1]
	var parent *treeNodeV4
	var targetNode *treeNodeV4
	var targetNodeIndex uint

	if address == nil || address.Length == 0 {
		// caller just looking for root tags
		targetNode = root
		targetNodeIndex = 1
	} else {
		nodeIndex := uint(0)

		parent = root
		if address.Address < _leftmost32Bit {
			nodeIndex = root.Left
		} else {
			nodeIndex = root.Right
		}

		// traverse the tree
		for {
			if nodeIndex == 0 {
				return 0, nil
			}

			node := &t.nodes[nodeIndex]
			matchCount := node.MatchCount(address)
			if matchCount < node.prefixLength {
				// didn't match the entire node - we're done
				return 0, nil
			}

			if matchCount == address.Length {
				// exact match - we're done
				targetNode = node
				targetNodeIndex = nodeIndex
				break
			}

			// there's still more address - keep traversing
			parent = node
			address.ShiftLeft(matchCount)
			if address.Address < _leftmost32Bit {
				nodeIndex = node.Left
			} else {
				nodeIndex = node.Right
			}
		}
	}

	if targetNode == nil || targetNode.TagCount == 0 {
		// no tags found
		return 0, nil
	}

	// we have tags to delete - build up a new list with the exact size needed
	deleteCount, remainingTagCount := t.deleteTag(targetNodeIndex, matchVal, matchFunc)
	if remainingTagCount > 0 {
		// target node still has tags - we're not deleting it
		return deleteCount, nil
	}

	if targetNodeIndex == 1 {
		// can't delete the root node
		return deleteCount, nil
	}

	// see if we can just move the children up
	if targetNode.Left != 0 && targetNode.Right != 0 {
		if parent.Left == 0 || parent.Right == 0 {
			// target node has two children, parent has just the target node - move target node's children up
			parent.Left = targetNode.Left
			parent.Right = targetNode.Right

			// need to update the parent prefix to include target node's
			parent.prefix, parent.prefixLength = patricia.MergePrefixes32(parent.prefix, parent.prefixLength, targetNode.prefix, targetNode.prefixLength)
		}
	} else if targetNode.Left != 0 {
		// target node only has only left child
		// move target's left node up
		if parent.Left == targetNodeIndex {
			parent.Left = targetNode.Left
		} else {
			parent.Right = targetNode.Left
		}

		// need to update the child node prefix to include target node's
		tmpNode := &t.nodes[targetNode.Left]
		tmpNode.prefix, tmpNode.prefixLength = patricia.MergePrefixes32(targetNode.prefix, targetNode.prefixLength, tmpNode.prefix, tmpNode.prefixLength)
	} else if targetNode.Right != 0 {
		// target node has only right child

		// only has right child - see where it goes
		if parent.Left == targetNodeIndex {
			parent.Left = targetNode.Right
		} else {
			parent.Right = targetNode.Right
		}

		// need to update the child node prefix to include target node's
		tmpNode := &t.nodes[targetNode.Right]
		tmpNode.prefix, tmpNode.prefixLength = patricia.MergePrefixes32(targetNode.prefix, targetNode.prefixLength, tmpNode.prefix, tmpNode.prefixLength)
	} else {
		// target node has no children
		if parent.Left == targetNodeIndex {
			parent.Left = 0
		} else {
			parent.Right = 0
		}
	}

	t.availableIndexes = append(t.availableIndexes, targetNodeIndex)
	return deleteCount, nil
}

// FindTagsWithFilter finds all matching tags that passes the filter function
func (t *TreeV4) FindTagsWithFilter(address *patricia.IPv4Address, filterFunc FilterFunc) ([]string, error) {
	root := &t.nodes[1]
	if filterFunc == nil {
		return t.FindTags(address)
	}

	var matchCount uint
	ret := make([]string, 0)

	if root.TagCount > 0 {
		for _, tag := range t.tagsForNode(1) {
			if filterFunc(tag) {
				ret = append(ret, tag)
			}
		}
	}

	if address == nil || address.Length == 0 {
		// caller just looking for root tags
		return ret, nil
	}

	var nodeIndex uint
	if address.Address < _leftmost32Bit {
		nodeIndex = root.Left
	} else {
		nodeIndex = root.Right
	}

	// traverse the tree
	count := 0
	for {
		count++
		if nodeIndex == 0 {
			return ret, nil
		}
		node := &t.nodes[nodeIndex]

		matchCount = node.MatchCount(address)
		if matchCount < node.prefixLength {
			// didn't match the entire node - we're done
			return ret, nil
		}

		// matched the full node - get its tags, then chop off the bits we've already matched and continue
		if node.TagCount > 0 {
			for _, tag := range t.tagsForNode(nodeIndex) {
				if filterFunc(tag) {
					ret = append(ret, tag)
				}
			}
		}

		if matchCount == address.Length {
			// exact match - we're done
			return ret, nil
		}

		// there's still more address - keep traversing
		address.ShiftLeft(matchCount)
		if address.Address < _leftmost32Bit {
			nodeIndex = node.Left
		} else {
			nodeIndex = node.Right
		}
	}
}

// FindTags finds all matching tags that passes the filter function
func (t *TreeV4) FindTags(address *patricia.IPv4Address) ([]string, error) {
	var matchCount uint
	root := &t.nodes[1]
	ret := make([]string, 0)

	if root.TagCount > 0 {
		ret = append(ret, t.tagsForNode(1)...)
	}

	if address == nil || address.Length == 0 {
		// caller just looking for root tags
		return ret, nil
	}

	var nodeIndex uint
	if address.Address < _leftmost32Bit {
		nodeIndex = root.Left
	} else {
		nodeIndex = root.Right
	}

	// traverse the tree
	count := 0
	for {
		count++
		if nodeIndex == 0 {
			return ret, nil
		}
		node := &t.nodes[nodeIndex]

		matchCount = node.MatchCount(address)
		if matchCount < node.prefixLength {
			// didn't match the entire node - we're done
			return ret, nil
		}

		// matched the full node - get its tags, then chop off the bits we've already matched and continue
		if node.TagCount > 0 {
			ret = append(ret, t.tagsForNode(nodeIndex)...)
		}

		if matchCount == address.Length {
			// exact match - we're done
			return ret, nil
		}

		// there's still more address - keep traversing
		address.ShiftLeft(matchCount)
		if address.Address < _leftmost32Bit {
			nodeIndex = node.Left
		} else {
			nodeIndex = node.Right
		}
	}
}

// FindDeepestTag finds a tag at the deepest level in the tree, representing the closest match
func (t *TreeV4) FindDeepestTag(address *patricia.IPv4Address) (bool, string, error) {
	root := &t.nodes[1]
	var found bool
	var ret string

	if root.TagCount > 0 {
		ret = t.firstTagForNode(1)
		found = true
	}

	if address.Length == 0 {
		// caller just looking for root tags
		return found, ret, nil
	}

	var nodeIndex uint
	if address.Address < _leftmost32Bit {
		nodeIndex = root.Left
	} else {
		nodeIndex = root.Right
	}

	// traverse the tree
	for {
		if nodeIndex == 0 {
			return found, ret, nil
		}
		node := &t.nodes[nodeIndex]

		matchCount := node.MatchCount(address)
		if matchCount < node.prefixLength {
			// didn't match the entire node - we're done
			return found, ret, nil
		}

		// matched the full node - get its tags, then chop off the bits we've already matched and continue
		if node.TagCount > 0 {
			ret = t.firstTagForNode(nodeIndex)
			found = true
		}

		if matchCount == address.Length {
			// exact match - we're done
			return found, ret, nil
		}

		// there's still more address - keep traversing
		address.ShiftLeft(matchCount)
		if address.Address < _leftmost32Bit {
			nodeIndex = node.Left
		} else {
			nodeIndex = node.Right
		}
	}
}

func (t *TreeV4) countNodes(nodeIndex uint) int {
	nodeCount := 1

	node := &t.nodes[nodeIndex]
	if node.Left != 0 {
		nodeCount += t.countNodes(node.Left)
	}
	if node.Right != 0 {
		nodeCount += t.countNodes(node.Right)
	}
	return nodeCount
}

func (t *TreeV4) countTags(nodeIndex uint) uint {
	node := &t.nodes[nodeIndex]

	tagCount := node.TagCount
	if node.Left != 0 {
		tagCount += t.countTags(node.Left)
	}
	if node.Right != 0 {
		tagCount += t.countTags(node.Right)
	}
	return tagCount
}
