package uint_array_tree

import (
	"github.com/kentik/patricia"
)

const _leftmost32Bit = uint32(1 << 31)

// MatchesFunc is called to check if tag data matches the input value
type MatchesFunc func(payload []uint, val []uint) bool

// FilterFunc is called on each result to see if it belongs in the resulting set
type FilterFunc func(payload []uint) bool

// Tree is an IP Address patricia tree
type TreeV4 struct {
	root *treeNodeV4
}

// NewTree returns a new Tree
func NewTreeV4() *TreeV4 {
	return &TreeV4{
		root: &treeNodeV4{},
	}
}

// Add adds a node to the tree
func (t *TreeV4) Add(address *patricia.IPv4Address, tag []uint) error {
	// handle root tags
	if address == nil || address.Length == 0 {
		t.root.AddTag(tag)
		return nil
	}

	// root node doesn't have any prefix, so find the starting point
	var node *treeNodeV4
	parent := t.root
	if address.Address < _leftmost32Bit {
		if t.root.Left == nil {
			t.root.Left = &treeNodeV4{
				prefix:       address.Address,
				prefixLength: address.Length,
			}
			t.root.Left.AddTag(tag)
			return nil
		}
		node = t.root.Left
	} else {
		if t.root.Right == nil {
			t.root.Right = &treeNodeV4{
				prefix:       address.Address,
				prefixLength: address.Length,
			}
			t.root.Right.AddTag(tag)
			return nil
		}
		node = t.root.Right
	}

	for {
		matchCount := uint(node.MatchCount(address))
		if matchCount == address.Length {
			// all the bits in the address matched

			if matchCount == node.prefixLength {
				// the whole prefix matched - we're done!
				node.AddTag(tag)
				return nil
			}

			// the input address is shorter than the match found - need to create a new, intermediate parent
			newNode := &treeNodeV4{
				prefix:       address.Address,
				prefixLength: address.Length,
			}
			newNode.AddTag(tag)

			// the existing node loses those matching bits, and becomes a child of the new node

			// shift
			node.prefix <<= matchCount
			node.prefixLength -= matchCount

			if node.prefix < _leftmost32Bit {
				newNode.Left = node
			} else {
				newNode.Right = node
			}

			// now give this new node a home
			if parent.Left == node {
				parent.Left = newNode
			} else {
				if parent.Right != node {
					panic("node isn't left or right parent - should be impossible! (1)")
				}
				parent.Right = newNode
			}
			return nil
		}

		if matchCount == 0 {
			panic("Should not have traversed to a node with no prefix match")
		}

		if matchCount == node.prefixLength {
			// partial match - we have to keep traversing

			// shift
			address.Address <<= matchCount // chop off what's matched so far
			address.Length -= matchCount

			if address.Address < _leftmost32Bit {
				if node.Left == nil {
					// nowhere else to go - create a new node here
					node.Left = &treeNodeV4{
						prefix:       address.Address,
						prefixLength: address.Length,
					}
					node.Left.AddTag(tag)
					return nil
				}

				// there's a node to the left - traverse it
				parent = node
				node = node.Left
				continue
			}

			// node didn't belong on the left, so it belongs on the right
			if node.Right == nil {
				// nowhere else to go - create a new node here
				node.Right = &treeNodeV4{
					prefix:       address.Address,
					prefixLength: address.Length,
				}
				node.Right.AddTag(tag)
				return nil
			}

			// there's a node to the right - traverse it
			parent = node
			node = node.Right
			continue
		}

		// partial match with this node - need to split this node
		newCommonParentNode := &treeNodeV4{
			prefix:       address.Address,
			prefixLength: matchCount,
		}
		// shift
		address.Address <<= matchCount
		address.Length -= matchCount

		// see where the existing node fits - left or right
		// shift
		node.prefix <<= matchCount
		node.prefixLength -= matchCount
		if node.prefix < _leftmost32Bit {
			newCommonParentNode.Left = node
			newCommonParentNode.Right = &treeNodeV4{
				prefix:       address.Address,
				prefixLength: address.Length,
			}
			newCommonParentNode.Right.AddTag(tag)
		} else {
			newCommonParentNode.Right = node
			newCommonParentNode.Left = &treeNodeV4{
				prefix:       address.Address,
				prefixLength: address.Length,
			}
			newCommonParentNode.Left.AddTag(tag)
		}

		// now determine where the new node belongs
		if parent.Left == node {
			parent.Left = newCommonParentNode
		} else {
			if parent.Right != node {
				panic("node isn't left or right parent - should be impossible! (2)")
			}
			parent.Right = newCommonParentNode
		}
		return nil
	}
}

// Delete a tag from the tree if it matches matchVal, as determined by matchFunc. Returns how many tags are removed
func (t *TreeV4) Delete(address *patricia.IPv4Address, matchFunc MatchesFunc, matchVal []uint) (int, error) {
	// traverse the tree, finding the node and its parent
	var parent *treeNodeV4
	var targetNode *treeNodeV4

	if address.Length == 0 {
		// caller just looking for root tags
		targetNode = t.root
	} else {
		var node *treeNodeV4
		parent = t.root
		if address.Address < _leftmost32Bit {
			node = t.root.Left
		} else {
			node = t.root.Right
		}

		// traverse the tree
		for {
			if node == nil {
				return 0, nil
			}

			matchCount := node.MatchCount(address)
			if matchCount < node.prefixLength {
				// didn't match the entire node - we're done
				return 0, nil
			}

			if matchCount == address.Length {
				// exact match - we're done
				targetNode = node
				break
			}

			// there's still more address - keep traversing
			parent = node
			address.Address <<= matchCount
			address.Length -= matchCount
			if address.Address < _leftmost32Bit {
				node = node.Left
			} else {
				node = node.Right
			}
		}
	}

	if targetNode == nil || !targetNode.HasTags {
		// no tags found
		return 0, nil
	}

	// we have tags - see if any need to be deleted
	deleteCount := 0
	matchIndices := make(map[int]bool)
	for index, tagData := range targetNode.Tags {
		if matchFunc(tagData, matchVal) {
			matchIndices[index] = true
			deleteCount++
		}
	}
	if len(matchIndices) == 0 {
		// node exists, but doesn't have our tag
		return 0, nil
	}

	// we have tags to delete - build up a new list with the exact size needed
	newTagListLength := len(targetNode.Tags) - len(matchIndices)
	if newTagListLength > 0 {
		// node will still have tags when we're done with it
		newTagList := make([][]uint, 0, newTagListLength)
		for index, tagData := range targetNode.Tags {
			if _, ok := matchIndices[index]; !ok {
				newTagList = append(newTagList, tagData)
			}
		}
		targetNode.Tags = newTagList

		// target node still has tags - we're not deleting it
		return deleteCount, nil
	}

	// this node no longer has tags
	targetNode.Tags = nil
	targetNode.HasTags = false

	if targetNode == t.root {
		// can't delete the root node
		return deleteCount, nil
	}

	// see if we can just move the children up
	if targetNode.Left != nil && targetNode.Right != nil {
		if parent.Left == nil || parent.Right == nil {
			// target node has two children, parent has just the target node - move target node's children up
			parent.Left = targetNode.Left
			parent.Right = targetNode.Right

			// need to update the parent prefix to include target node's
			parent.prefix, parent.prefixLength = mergePrefixes32(parent.prefix, parent.prefixLength, targetNode.prefix, targetNode.prefixLength)
		}
	} else if targetNode.Left != nil {
		// target node only has only left child
		// move target's left node up
		if parent.Left == targetNode {
			parent.Left = targetNode.Left
		} else {
			parent.Right = targetNode.Left
		}

		// need to update the child node prefix to include target node's
		targetNode.Left.prefix, targetNode.Left.prefixLength = mergePrefixes32(targetNode.prefix, targetNode.prefixLength, targetNode.Left.prefix, targetNode.Left.prefixLength)
	} else if targetNode.Right != nil {
		// target node has only right child

		// only has right child - see where it goes
		if parent.Left == targetNode {
			parent.Left = targetNode.Right
		} else {
			parent.Right = targetNode.Right
		}

		// need to update the child node prefix to include target node's
		targetNode.Right.prefix, targetNode.Right.prefixLength = mergePrefixes32(targetNode.prefix, targetNode.prefixLength, targetNode.Right.prefix, targetNode.Right.prefixLength)
	} else {
		// target node has no children
		if parent.Left == targetNode {
			parent.Left = nil
		} else {
			parent.Right = nil
		}
	}

	return deleteCount, nil
}

// FindTagsWithFilter finds all matching tags that passes the filter function
func (t *TreeV4) FindTagsWithFilter(address *patricia.IPv4Address, filterFunc FilterFunc) ([][]uint, error) {
	if filterFunc == nil {
		return t.FindTags(address)
	}

	var matchCount uint
	ret := make([][]uint, 0)

	if t.root.HasTags {
		for _, tag := range t.root.Tags {
			if filterFunc(tag) {
				ret = append(ret, tag)
			}
		}
	}

	if address == nil || address.Length == 0 {
		// caller just looking for root tags
		return ret, nil
	}

	var node *treeNodeV4
	if address.Address < _leftmost32Bit {
		node = t.root.Left
	} else {
		node = t.root.Right
	}

	// traverse the tree
	count := 0
	for {
		count++
		if node == nil {
			return ret, nil
		}

		matchCount = node.MatchCount(address)
		if matchCount < node.prefixLength {
			// didn't match the entire node - we're done
			return ret, nil
		}

		// matched the full node - get its tags, then chop off the bits we've already matched and continue
		if node.HasTags {
			for _, tag := range node.Tags {
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
		// shift
		address.Address <<= matchCount
		address.Length -= matchCount
		if address.Address < _leftmost32Bit {
			node = node.Left
		} else {
			node = node.Right
		}
	}
}

// FindTags finds all matching tags that passes the filter function
func (t *TreeV4) FindTags(address *patricia.IPv4Address) ([][]uint, error) {
	var matchCount uint
	ret := make([][]uint, 0)

	if t.root.HasTags {
		ret = append(ret, t.root.Tags...)
	}

	if address == nil || address.Length == 0 {
		// caller just looking for root tags
		return ret, nil
	}

	var node *treeNodeV4
	if address.Address < _leftmost32Bit {
		node = t.root.Left
	} else {
		node = t.root.Right
	}

	// traverse the tree
	count := 0
	for {
		count++
		if node == nil {
			return ret, nil
		}

		matchCount = node.MatchCount(address)
		if matchCount < node.prefixLength {
			// didn't match the entire node - we're done
			return ret, nil
		}

		// matched the full node - get its tags, then chop off the bits we've already matched and continue
		if node.HasTags {
			ret = append(ret, node.Tags...)
		}

		if matchCount == address.Length {
			// exact match - we're done
			return ret, nil
		}

		// there's still more address - keep traversing
		// shift
		address.Address <<= matchCount
		address.Length -= matchCount
		if address.Address < _leftmost32Bit {
			node = node.Left
		} else {
			node = node.Right
		}
	}
}

// FindDeepestTag finds a tag at the deepest level in the tree, representing the closest match
func (t *TreeV4) FindDeepestTag(address *patricia.IPv4Address) (bool, []uint, error) {
	var found bool
	var ret []uint

	if t.root.HasTags {
		ret = t.root.Tags[0]
		found = true
	}

	if address.Length == 0 {
		// caller just looking for root tags
		return found, ret, nil
	}

	var node *treeNodeV4
	if address.Address < _leftmost32Bit {
		node = t.root.Left
	} else {
		node = t.root.Right
	}

	// traverse the tree
	for {
		if node == nil {
			return found, ret, nil
		}

		matchCount := node.MatchCount(address)
		if matchCount < node.prefixLength {
			// didn't match the entire node - we're done
			return found, ret, nil
		}

		// matched the full node - get its tags, then chop off the bits we've already matched and continue
		if node.HasTags {
			ret = node.Tags[0]
			found = true
		}

		if matchCount == address.Length {
			// exact match - we're done
			return found, ret, nil
		}

		// there's still more address - keep traversing
		address.Address <<= matchCount
		address.Length -= matchCount
		if address.Address < _leftmost32Bit {
			node = node.Left
		} else {
			node = node.Right
		}
	}
}

func countNodes(node *treeNodeV4) int {
	nodeCount := 1
	if node.Left != nil {
		nodeCount += countNodes(node.Left)
	}
	if node.Right != nil {
		nodeCount += countNodes(node.Right)
	}
	return nodeCount
}

func countTags(node *treeNodeV4) int {
	tagCount := len(node.Tags)
	if node.Left != nil {
		tagCount += countTags(node.Left)
	}
	if node.Right != nil {
		tagCount += countTags(node.Right)
	}
	return tagCount
}
