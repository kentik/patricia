package template

import (
	"github.com/kentik/patricia"
)

const _leftmost32Bit = uint32(1 << 31)

// MatchesFunc is called to check if tag data matches the input value
type MatchesFunc func(payload GeneratedType, val GeneratedType) bool

// FilterFunc is called on each result to see if it belongs in the resulting set
type FilterFunc func(payload GeneratedType) bool

// Tree is an IP Address patricia tree
type TreeV4 struct {
	nodes            []treeNodeV4 // root is always at [1] - [0] is unused
	availableIndexes []uint       // a place to store node indexes that we deleted, and are available
}

// NewTree returns a new Tree
func NewTreeV4() *TreeV4 {
	return &TreeV4{
		nodes:            make([]treeNodeV4, 2, 1024),
		availableIndexes: make([]uint, 0),
	}
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
func (t *TreeV4) Add(address *patricia.IPv4Address, tag GeneratedType) error {
	root := &t.nodes[1]

	// handle root tags
	if address == nil || address.Length == 0 {
		root.AddTag(tag)
		return nil
	}

	// root node doesn't have any prefix, so find the starting point
	nodeIndex := uint(0)
	parent := root
	if address.Address < _leftmost32Bit {
		if root.Left == 0 {
			newNodeIndex := t.newNode(address.Address, address.Length)
			newNode := &t.nodes[newNodeIndex]
			newNode.AddTag(tag)
			root.Left = newNodeIndex
			return nil
		}
		nodeIndex = root.Left
	} else {
		if root.Right == 0 {
			newNodeIndex := t.newNode(address.Address, address.Length)
			newNode := &t.nodes[newNodeIndex]
			newNode.AddTag(tag)
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
				node.AddTag(tag)
				return nil
			}

			// the input address is shorter than the match found - need to create a new, intermediate parent
			newNodeIndex := t.newNode(address.Address, address.Length)
			newNode := &t.nodes[newNodeIndex]
			newNode.AddTag(tag)

			// the existing node loses those matching bits, and becomes a child of the new node

			// shift
			node.prefix <<= matchCount
			node.prefixLength -= matchCount

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

			// shift
			address.Address <<= matchCount // chop off what's matched so far
			address.Length -= matchCount

			if address.Address < _leftmost32Bit {
				if node.Left == 0 {
					// nowhere else to go - create a new node here
					newNodeIndex := t.newNode(address.Address, address.Length)
					newNode := &t.nodes[newNodeIndex]
					newNode.AddTag(tag)
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
				newNode := &t.nodes[newNodeIndex]
				newNode.AddTag(tag)
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
		address.Address <<= matchCount
		address.Length -= matchCount

		newNodeIndex := t.newNode(address.Address, address.Length)
		newNode := &t.nodes[newNodeIndex]
		newNode.AddTag(tag)

		// see where the existing node fits - left or right
		// shift
		node.prefix <<= matchCount
		node.prefixLength -= matchCount
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
func (t *TreeV4) Delete(address *patricia.IPv4Address, matchFunc MatchesFunc, matchVal GeneratedType) (int, error) {
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
			address.Address <<= matchCount
			address.Length -= matchCount
			if address.Address < _leftmost32Bit {
				nodeIndex = node.Left
			} else {
				nodeIndex = node.Right
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
		newTagList := make([]GeneratedType, 0, newTagListLength)
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
func (t *TreeV4) FindTagsWithFilter(address *patricia.IPv4Address, filterFunc FilterFunc) ([]GeneratedType, error) {
	root := &t.nodes[1]
	if filterFunc == nil {
		return t.FindTags(address)
	}

	var matchCount uint
	ret := make([]GeneratedType, 0)

	if root.HasTags {
		for _, tag := range root.Tags {
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
			nodeIndex = node.Left
		} else {
			nodeIndex = node.Right
		}
	}
}

// FindTags finds all matching tags that passes the filter function
func (t *TreeV4) FindTags(address *patricia.IPv4Address) ([]GeneratedType, error) {
	var matchCount uint
	root := &t.nodes[1]
	ret := make([]GeneratedType, 0)

	if root.HasTags {
		ret = append(ret, root.Tags...)
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
			nodeIndex = node.Left
		} else {
			nodeIndex = node.Right
		}
	}
}

// FindDeepestTag finds a tag at the deepest level in the tree, representing the closest match
func (t *TreeV4) FindDeepestTag(address *patricia.IPv4Address) (bool, GeneratedType, error) {
	root := &t.nodes[1]
	var found bool
	var ret GeneratedType

	if root.HasTags {
		ret = root.Tags[0]
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

func (t *TreeV4) countTags(nodeIndex uint) int {
	node := &t.nodes[nodeIndex]

	tagCount := len(node.Tags)
	if node.Left != 0 {
		tagCount += t.countTags(node.Left)
	}
	if node.Right != 0 {
		tagCount += t.countTags(node.Right)
	}
	return tagCount
}
