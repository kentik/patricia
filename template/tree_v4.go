// Template file.

package template

import (
	"fmt"

	"github.com/kentik/patricia"
)

// TreeV4 is an IP Address patricia tree
type TreeV4 struct {
	nodes            []treeNodeV4 // root is always at [1] - [0] is unused
	availableIndexes []uint       // a place to store node indexes that we deleted, and are available
	tags             map[uint64]GeneratedType
}

// NewTreeV4 returns a new Tree
func NewTreeV4() *TreeV4 {
	return &TreeV4{
		nodes:            make([]treeNodeV4, 2), // index 0 is skipped, 1 is root
		availableIndexes: make([]uint, 0),
		tags:             make(map[uint64]GeneratedType),
	}
}

// Clone creates an identical copy of the tree
// - Note: the items in the tree are not deep copied
func (t *TreeV4) Clone() *TreeV4 {
	ret := &TreeV4{
		nodes:            make([]treeNodeV4, len(t.nodes), cap(t.nodes)),
		availableIndexes: make([]uint, len(t.availableIndexes), cap(t.availableIndexes)),
		tags:             make(map[uint64]GeneratedType, len(t.tags)),
	}

	copy(ret.nodes, t.nodes)
	copy(ret.availableIndexes, t.availableIndexes)
	for k, v := range t.tags {
		ret.tags[k] = v
	}
	return ret
}

// CountTags iterates through the tree, counting the number of tags
// - note: unused nodes will have TagCount==0
func (t *TreeV4) CountTags() int {
	ret := 0
	for _, node := range t.nodes {
		ret += node.TagCount
	}
	return ret
}

// add a tag to the node at the input index
// - if matchFunc is non-nil, it is used to determine equality (if nil, no existing tag match)
// - if udpateFunc is non-nil, it is used to update the tag if it already exists (if nil, the provided tag is used)
// - returns whether the tag count was increased
func (t *TreeV4) addTag(tag GeneratedType, nodeIndex uint, matchFunc MatchesFunc, updateFunc UpdatesFunc) bool {
	key := (uint64(nodeIndex) << 32)
	tagCount := t.nodes[nodeIndex].TagCount
	if matchFunc != nil {
		// need to check if this value already exists
		for i := 0; i < tagCount; i++ {
			if matchFunc(t.tags[key+uint64(i)], tag) {
				if updateFunc != nil {
					t.tags[key+(uint64(i))] = updateFunc(t.tags[key+(uint64(i))])
				}
				return false
			}
		}
	}
	t.tags[key+(uint64(tagCount))] = tag
	t.nodes[nodeIndex].TagCount++
	return true
}

// return the tags at the input node index - appending to the input slice if they pass the optional filter func
// - ret is only appended to
func (t *TreeV4) tagsForNode(ret []GeneratedType, nodeIndex uint, filterFunc FilterFunc) []GeneratedType {
	if nodeIndex == 0 {
		// useful for base cases where we haven't found anything
		return ret
	}

	// TODO: clean up the typing in here, between uint, uint64
	tagCount := t.nodes[nodeIndex].TagCount
	key := uint64(nodeIndex) << 32
	for i := 0; i < tagCount; i++ {
		tag := t.tags[key+uint64(i)]
		if filterFunc == nil || filterFunc(tag) {
			ret = append(ret, tag)
		}
	}
	return ret
}

func (t *TreeV4) moveTags(fromIndex uint, toIndex uint) {
	tagCount := t.nodes[fromIndex].TagCount
	fromKey := uint64(fromIndex) << 32
	toKey := uint64(toIndex) << 32
	for i := 0; i < tagCount; i++ {
		t.tags[toKey+uint64(i)] = t.tags[fromKey+uint64(i)]
		delete(t.tags, fromKey+uint64(i))
	}
	t.nodes[toIndex].TagCount += t.nodes[fromIndex].TagCount
	t.nodes[fromIndex].TagCount = 0
}

func (t *TreeV4) firstTagForNode(nodeIndex uint) GeneratedType {
	return t.tags[(uint64(nodeIndex) << 32)]
}

// delete tags at the input node, returning how many were deleted, and how many are left
// - uses input slice to reduce allocations
func (t *TreeV4) deleteTag(buf []GeneratedType, nodeIndex uint, matchTag GeneratedType, matchFunc MatchesFunc) (int, int) {
	// get tags
	buf = buf[:0]
	buf = t.tagsForNode(buf, nodeIndex, nil)
	if len(buf) == 0 {
		return 0, 0
	}

	// delete tags
	// TODO: this could be done smarter - delete in place?
	for i := 0; i < t.nodes[nodeIndex].TagCount; i++ {
		delete(t.tags, (uint64(nodeIndex)<<32)+uint64(i))
	}
	t.nodes[nodeIndex].TagCount = 0

	// put them back
	deleteCount := 0
	keepCount := 0
	for _, tag := range buf {
		if matchFunc(tag, matchTag) {
			deleteCount++
		} else {
			// doesn't match - get to keep it
			t.addTag(tag, nodeIndex, matchFunc, nil)
			keepCount++
		}
	}
	return deleteCount, keepCount
}

// Set the single value for a node - overwrites what's there
// Returns whether the tag count at this address was increased, and how many tags at this address
func (t *TreeV4) Set(address patricia.IPv4Address, tag GeneratedType) (bool, int) {
	return t.add(address, tag,
		func(GeneratedType, GeneratedType) bool { return true },
		func(GeneratedType) GeneratedType { return tag })
}

// Add adds a tag to the tree
// - if matchFunc is non-nil, it will be used to ensure uniqueness at this node
// - returns whether the tag count at this address was increased, and how many tags at this address
func (t *TreeV4) Add(address patricia.IPv4Address, tag GeneratedType, matchFunc MatchesFunc) (bool, int) {
	return t.add(address, tag, matchFunc, nil)
}

// SetOrUpdate the single value for a node - overwrites what's there using updateFunc if present
// - returns whether the tag count at this address was increased, and how many tags at this address
func (t *TreeV4) SetOrUpdate(address patricia.IPv4Address, tag GeneratedType, updateFunc UpdatesFunc) (bool, int) {
	return t.add(address, tag,
		func(GeneratedType, GeneratedType) bool { return true },
		updateFunc)
}

// AddOrUpdate adds a tag to the tree or update it if it already exists
// - if matchFunc is non-nil, it will be used to ensure uniqueness at this node
// - returns whether the tag count at this address was increased, and how many tags at this address
func (t *TreeV4) AddOrUpdate(address patricia.IPv4Address, tag GeneratedType, matchFunc MatchesFunc, updateFunc UpdatesFunc) (bool, int) {
	return t.add(address, tag, matchFunc, updateFunc)
}

// add a tag to the tree, optionally updating the existing value
// - overwrites the first value in the list if updateFunc function is provided (tag is ignored in this case)
// - returns whether the tag count was increased, and the number of tags at this address
func (t *TreeV4) add(address patricia.IPv4Address, tag GeneratedType, matchFunc MatchesFunc, updateFunc UpdatesFunc) (bool, int) {
	// make sure we have more than enough capacity before we start adding to the tree, which invalidates pointers into the array
	if (len(t.availableIndexes) + cap(t.nodes)) < (len(t.nodes) + 10) {
		temp := make([]treeNodeV4, len(t.nodes), (cap(t.nodes)+1)*2)
		copy(temp, t.nodes)
		t.nodes = temp
	}

	root := &t.nodes[1]

	// handle root tags
	if address.Length == 0 {
		countIncreased := t.addTag(tag, 1, matchFunc, updateFunc)
		return countIncreased, t.nodes[1].TagCount
	}

	// root node doesn't have any prefix, so find the starting point
	nodeIndex := uint(0)
	parent := root
	if !address.IsLeftBitSet() {
		if root.Left == 0 {
			newNodeIndex := t.newNode(address, address.Length)
			countIncreased := t.addTag(tag, newNodeIndex, matchFunc, updateFunc)
			root.Left = newNodeIndex
			return countIncreased, t.nodes[newNodeIndex].TagCount
		}
		nodeIndex = root.Left
	} else {
		if root.Right == 0 {
			newNodeIndex := t.newNode(address, address.Length)
			countIncreased := t.addTag(tag, newNodeIndex, matchFunc, updateFunc)
			root.Right = newNodeIndex
			return countIncreased, t.nodes[newNodeIndex].TagCount
		}
		nodeIndex = root.Right
	}

	for {
		if nodeIndex == 0 {
			panic("Trying to traverse nodeIndex=0")
		}
		node := &t.nodes[nodeIndex]
		if node.prefixLength == 0 {
			panic("Reached a node with no prefix")
		}

		matchCount := uint(node.MatchCount(address))
		if matchCount == 0 {
			panic(fmt.Sprintf("Should not have traversed to a node with no prefix match - node prefix length: %d; address prefix length: %d", node.prefixLength, address.Length))
		}

		if matchCount == address.Length {
			// all the bits in the address matched

			if matchCount == node.prefixLength {
				// the whole prefix matched - we're done!
				countIncreased := t.addTag(tag, nodeIndex, matchFunc, updateFunc)
				return countIncreased, t.nodes[nodeIndex].TagCount
			}

			// the input address is shorter than the match found - need to create a new, intermediate parent
			newNodeIndex := t.newNode(address, address.Length)
			newNode := &t.nodes[newNodeIndex]
			countIncreased := t.addTag(tag, newNodeIndex, matchFunc, updateFunc)

			// the existing node loses those matching bits, and becomes a child of the new node

			// shift
			node.ShiftPrefix(matchCount)

			if !node.IsLeftBitSet() {
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
			return countIncreased, t.nodes[newNodeIndex].TagCount
		}

		if matchCount == node.prefixLength {
			// partial match - we have to keep traversing

			// chop off what's matched so far
			address.ShiftLeft(matchCount)

			if !address.IsLeftBitSet() {
				if node.Left == 0 {
					// nowhere else to go - create a new node here
					newNodeIndex := t.newNode(address, address.Length)
					countIncreased := t.addTag(tag, newNodeIndex, matchFunc, updateFunc)
					node.Left = newNodeIndex
					return countIncreased, t.nodes[newNodeIndex].TagCount
				}

				// there's a node to the left - traverse it
				parent = node
				nodeIndex = node.Left
				continue
			}

			// node didn't belong on the left, so it belongs on the right
			if node.Right == 0 {
				// nowhere else to go - create a new node here
				newNodeIndex := t.newNode(address, address.Length)
				countIncreased := t.addTag(tag, newNodeIndex, matchFunc, updateFunc)
				node.Right = newNodeIndex
				return countIncreased, t.nodes[newNodeIndex].TagCount
			}

			// there's a node to the right - traverse it
			parent = node
			nodeIndex = node.Right
			continue
		}

		// partial match with this node - need to split this node
		newCommonParentNodeIndex := t.newNode(address, matchCount)
		newCommonParentNode := &t.nodes[newCommonParentNodeIndex]

		// shift
		address.ShiftLeft(matchCount)

		newNodeIndex := t.newNode(address, address.Length)
		countIncreased := t.addTag(tag, newNodeIndex, matchFunc, updateFunc)

		// see where the existing node fits - left or right
		node.ShiftPrefix(matchCount)
		if !node.IsLeftBitSet() {
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
		return countIncreased, t.nodes[newNodeIndex].TagCount
	}
}

// Delete a tag from the tree if it matches matchVal, as determined by matchFunc. Returns how many tags are removed
// - use DeleteWithBuffer if you can reuse slices, to cut down on allocations
func (t *TreeV4) Delete(address patricia.IPv4Address, matchFunc MatchesFunc, matchVal GeneratedType) int {
	return t.DeleteWithBuffer(nil, address, matchFunc, matchVal)
}

// DeleteWithBuffer a tag from the tree if it matches matchVal, as determined by matchFunc. Returns how many tags are removed
// - uses input slice to reduce allocations
func (t *TreeV4) DeleteWithBuffer(buf []GeneratedType, address patricia.IPv4Address, matchFunc MatchesFunc, matchVal GeneratedType) int {
	// traverse the tree, finding the node and its parent
	root := &t.nodes[1]
	var parentIndex uint
	var parent *treeNodeV4
	var targetNode *treeNodeV4
	var targetNodeIndex uint

	if address.Length == 0 {
		// caller just looking for root tags
		targetNode = root
		targetNodeIndex = 1
	} else {
		nodeIndex := uint(0)

		parentIndex = 1
		parent = root
		if !address.IsLeftBitSet() {
			nodeIndex = root.Left
		} else {
			nodeIndex = root.Right
		}

		// traverse the tree
		for {
			if nodeIndex == 0 {
				return 0
			}

			node := &t.nodes[nodeIndex]
			matchCount := node.MatchCount(address)
			if matchCount < node.prefixLength {
				// didn't match the entire node - we're done
				return 0
			}

			if matchCount == address.Length {
				// exact match - we're done
				targetNode = node
				targetNodeIndex = nodeIndex
				break
			}

			// there's still more address - keep traversing
			parentIndex = nodeIndex
			parent = node
			address.ShiftLeft(matchCount)
			if !address.IsLeftBitSet() {
				nodeIndex = node.Left
			} else {
				nodeIndex = node.Right
			}
		}
	}

	if targetNode == nil || targetNode.TagCount == 0 {
		// no tags found
		return 0
	}

	// delete matching tags
	deleteCount, remainingTagCount := t.deleteTag(buf, targetNodeIndex, matchVal, matchFunc)
	if remainingTagCount > 0 {
		// target node still has tags - we're not deleting it
		return deleteCount
	}
	t.deleteNode(targetNodeIndex, targetNode, parentIndex, parent)
	return deleteCount
}

// deleteNode removes the provided node and compact the tree.
func (t *TreeV4) deleteNode(targetNodeIndex uint, targetNode *treeNodeV4, parentIndex uint, parent *treeNodeV4) (result deleteNodeResult) {
	result = notDeleted
	if targetNodeIndex == 1 {
		// can't delete the root node
		return result
	}

	// compact the tree, if possible
	if targetNode.Left != 0 && targetNode.Right != 0 {
		// target has two children - nothing we can do - not deleting the node
		return result
	} else if targetNode.Left != 0 {
		// target node only has only left child
		result = deletedNodeReplacedByChild
		if parent.Left == targetNodeIndex {
			parent.Left = targetNode.Left
		} else {
			parent.Right = targetNode.Left
		}

		// need to update the child node prefix to include target node's
		tmpNode := &t.nodes[targetNode.Left]
		tmpNode.MergeFromNodes(targetNode, tmpNode)
	} else if targetNode.Right != 0 {
		// target node has only right child
		result = deletedNodeReplacedByChild
		if parent.Left == targetNodeIndex {
			parent.Left = targetNode.Right
		} else {
			parent.Right = targetNode.Right
		}

		// need to update the child node prefix to include target node's
		tmpNode := &t.nodes[targetNode.Right]
		tmpNode.MergeFromNodes(targetNode, tmpNode)
	} else {
		// target node has no children - straight-up remove this node
		result = deletedNodeJustRemoved
		if parent.Left == targetNodeIndex {
			parent.Left = 0
			if parentIndex > 1 && parent.TagCount == 0 && parent.Right != 0 {
				// parent isn't root, has no tags, and there's a sibling - merge sibling into parent
				result = deletedNodeParentReplacedBySibling
				siblingIndexToDelete := parent.Right
				tmpNode := &t.nodes[siblingIndexToDelete]
				parent.MergeFromNodes(parent, tmpNode)

				// move tags
				t.moveTags(siblingIndexToDelete, parentIndex)

				// parent now gets target's sibling's children
				parent.Left = t.nodes[siblingIndexToDelete].Left
				parent.Right = t.nodes[siblingIndexToDelete].Right

				t.availableIndexes = append(t.availableIndexes, siblingIndexToDelete)
			}
		} else {
			parent.Right = 0
			if parentIndex > 1 && parent.TagCount == 0 && parent.Left != 0 {
				// parent isn't root, has no tags, and there's a sibling - merge sibling into parent
				result = deletedNodeParentReplacedBySibling
				siblingIndexToDelete := parent.Left
				tmpNode := &t.nodes[siblingIndexToDelete]
				parent.MergeFromNodes(parent, tmpNode)

				// move tags
				t.moveTags(siblingIndexToDelete, parentIndex)

				// parent now gets target's sibling's children
				parent.Right = t.nodes[parent.Left].Right
				parent.Left = t.nodes[parent.Left].Left

				t.availableIndexes = append(t.availableIndexes, siblingIndexToDelete)
			}
		}
	}

	targetNode.Left = 0
	targetNode.Right = 0
	t.availableIndexes = append(t.availableIndexes, targetNodeIndex)
	return result
}

// FindTagsWithFilter finds all matching tags that passes the filter function
// - use FindTagsWithFilterAppend if you can reuse slices, to cut down on allocations
func (t *TreeV4) FindTagsWithFilter(address patricia.IPv4Address, filterFunc FilterFunc) []GeneratedType {
	ret := make([]GeneratedType, 0)
	return t.FindTagsWithFilterAppend(ret, address, filterFunc)
}

// FindTagsAppend finds all matching tags for given address and appends them to ret
func (t *TreeV4) FindTagsAppend(ret []GeneratedType, address patricia.IPv4Address) []GeneratedType {
	return t.FindTagsWithFilterAppend(ret, address, nil)
}

// FindTags finds all matching tags for given address
// - use FindTagsAppend if you can reuse slices, to cut down on allocations
func (t *TreeV4) FindTags(address patricia.IPv4Address) []GeneratedType {
	ret := make([]GeneratedType, 0)
	return t.FindTagsAppend(ret, address)
}

// FindTagsWithFilterAppend finds all matching tags that passes the filter function
// - results are appended to the input slice
func (t *TreeV4) FindTagsWithFilterAppend(ret []GeneratedType, address patricia.IPv4Address, filterFunc FilterFunc) []GeneratedType {
	var matchCount uint
	root := &t.nodes[1]

	if root.TagCount > 0 {
		ret = t.tagsForNode(ret, 1, filterFunc)
	}

	if address.Length == 0 {
		// caller just looking for root tags
		return ret
	}

	var nodeIndex uint
	if !address.IsLeftBitSet() {
		nodeIndex = root.Left
	} else {
		nodeIndex = root.Right
	}

	// traverse the tree
	for {
		if nodeIndex == 0 {
			return ret
		}
		node := &t.nodes[nodeIndex]

		matchCount = node.MatchCount(address)
		if matchCount < node.prefixLength {
			// didn't match the entire node - we're done
			return ret
		}

		// matched the full node - get its tags, then chop off the bits we've already matched and continue
		if node.TagCount > 0 {
			ret = t.tagsForNode(ret, nodeIndex, filterFunc)
		}

		if matchCount == address.Length {
			// exact match - we're done
			return ret
		}

		// there's still more address - keep traversing
		address.ShiftLeft(matchCount)
		if !address.IsLeftBitSet() {
			nodeIndex = node.Left
		} else {
			nodeIndex = node.Right
		}
	}
}

// FindDeepestTag finds a tag at the deepest level in the tree, representing the closest match.
// - if that target node has multiple tags, the first in the list is returned
func (t *TreeV4) FindDeepestTag(address patricia.IPv4Address) (bool, GeneratedType) {
	root := &t.nodes[1]
	var found bool
	var ret GeneratedType

	if root.TagCount > 0 {
		ret = t.firstTagForNode(1)
		found = true
	}

	if address.Length == 0 {
		// caller just looking for root tags
		return found, ret
	}

	var nodeIndex uint
	if !address.IsLeftBitSet() {
		nodeIndex = root.Left
	} else {
		nodeIndex = root.Right
	}

	// traverse the tree
	for {
		if nodeIndex == 0 {
			return found, ret
		}
		node := &t.nodes[nodeIndex]

		matchCount := node.MatchCount(address)
		if matchCount < node.prefixLength {
			// didn't match the entire node - we're done
			return found, ret
		}

		// matched the full node - get its tags, then chop off the bits we've already matched and continue
		if node.TagCount > 0 {
			ret = t.firstTagForNode(nodeIndex)
			found = true
		}

		if matchCount == address.Length {
			// exact match - we're done
			return found, ret
		}

		// there's still more address - keep traversing
		address.ShiftLeft(matchCount)
		if !address.IsLeftBitSet() {
			nodeIndex = node.Left
		} else {
			nodeIndex = node.Right
		}
	}
}

// FindDeepestTags finds all tags at the deepest level in the tree, representing the closest match
// - use FindDeepestTagsAppend if you can reuse slices, to cut down on allocations
func (t *TreeV4) FindDeepestTags(address patricia.IPv4Address) (bool, []GeneratedType) {
	ret := make([]GeneratedType, 0)
	return t.FindDeepestTagsWithFilterAppend(ret, address, nil)
}

// FindDeepestTagsWithFilter finds all tags at the deepest level in the tree, matching the provided filter, representing the closest match
// - use FindDeepestTagsWithFilterAppend if you can reuse slices, to cut down on allocations
// - returns true regardless of the result of the filtering function
func (t *TreeV4) FindDeepestTagsWithFilter(address patricia.IPv4Address, filterFunc FilterFunc) (bool, []GeneratedType) {
	ret := make([]GeneratedType, 0)
	return t.FindDeepestTagsWithFilterAppend(ret, address, filterFunc)
}

// FindDeepestTagsAppend finds all tags at the deepest level in the tree, representing the closest match
// - appends results to the input slice
func (t *TreeV4) FindDeepestTagsAppend(ret []GeneratedType, address patricia.IPv4Address) (bool, []GeneratedType) {
	return t.FindDeepestTagsWithFilterAppend(ret, address, nil)
}

// FindDeepestTagsWithFilterAppend finds all tags at the deepest level in the tree, matching the provided filter, representing the closest match
// - appends results to the input slice
// - returns true regardless of the result of the filtering function
func (t *TreeV4) FindDeepestTagsWithFilterAppend(ret []GeneratedType, address patricia.IPv4Address, filterFunc FilterFunc) (bool, []GeneratedType) {
	root := &t.nodes[1]
	var found bool
	var retTagIndex uint

	if root.TagCount > 0 {
		retTagIndex = 1
		found = true
	}

	if address.Length == 0 {
		// caller just looking for root tags
		return found, t.tagsForNode(ret, retTagIndex, filterFunc)
	}

	var nodeIndex uint
	if !address.IsLeftBitSet() {
		nodeIndex = root.Left
	} else {
		nodeIndex = root.Right
	}

	// traverse the tree
	for {
		if nodeIndex == 0 {
			return found, t.tagsForNode(ret, retTagIndex, filterFunc)
		}
		node := &t.nodes[nodeIndex]

		matchCount := node.MatchCount(address)
		if matchCount < node.prefixLength {
			// didn't match the entire node - we're done
			return found, t.tagsForNode(ret, retTagIndex, filterFunc)
		}

		// matched the full node - get its tags, then chop off the bits we've already matched and continue
		if node.TagCount > 0 {
			retTagIndex = nodeIndex
			found = true
		}

		if matchCount == address.Length {
			// exact match - we're done
			return found, t.tagsForNode(ret, retTagIndex, filterFunc)
		}

		// there's still more address - keep traversing
		address.ShiftLeft(matchCount)
		if !address.IsLeftBitSet() {
			nodeIndex = node.Left
		} else {
			nodeIndex = node.Right
		}
	}
}

// TreeIteratorV4 is a stateful iterator over a tree.
type TreeIteratorV4 struct {
	t           *TreeV4
	nodeIndex   uint
	nodeHistory []uint
	next        treeIteratorNext
}

// Iterate returns an iterator to find all nodes from a tree. It is
// important for the tree to not be modified while using the iterator.
func (t *TreeV4) Iterate() *TreeIteratorV4 {
	return &TreeIteratorV4{
		t:           t,
		nodeIndex:   1,
		nodeHistory: []uint{},
		next:        nextSelf,
	}
}

// Next jumps to the next element of a tree. It returns false if there
// is none.
func (iter *TreeIteratorV4) Next() bool {
	for {
		node := &iter.t.nodes[iter.nodeIndex]
		if iter.next == nextSelf {
			iter.next = nextLeft
			if node.TagCount != 0 {
				return true
			}
		}
		if iter.next == nextLeft {
			if node.Left != 0 {
				iter.nodeHistory = append(iter.nodeHistory, iter.nodeIndex)
				iter.nodeIndex = node.Left
				iter.next = nextSelf
			} else {
				iter.next = nextRight
			}
		}
		if iter.next == nextRight {
			if node.Right != 0 {
				iter.nodeHistory = append(iter.nodeHistory, iter.nodeIndex)
				iter.nodeIndex = node.Right
				iter.next = nextSelf
			} else {
				// We need to backtrack
				iter.next = nextUp
			}
		}
		if iter.next == nextUp {
			nodeHistoryLen := len(iter.nodeHistory)
			if nodeHistoryLen == 0 {
				return false
			}
			previousIndex := iter.nodeHistory[nodeHistoryLen-1]
			previousNode := iter.t.nodes[previousIndex]
			iter.nodeHistory = iter.nodeHistory[:nodeHistoryLen-1]
			if previousNode.Left == iter.nodeIndex {
				iter.nodeIndex = previousIndex
				iter.next = nextRight
			} else if previousNode.Right == iter.nodeIndex {
				iter.nodeIndex = previousIndex
				iter.next = nextUp
			} else {
				panic("unexpected state")
			}
		}
	}
}

// Tags returns the current tags for the iterator. This is not a copy
// and the result should not be used outside the iterator.
func (iter *TreeIteratorV4) Tags() []GeneratedType {
	return iter.TagsWithBuffer(nil)
}

// TagsWithBuffer returns the current tags for the iterator. To avoid
// allocation, it uses the provided buffer.
func (iter *TreeIteratorV4) TagsWithBuffer(ret []GeneratedType) []GeneratedType {
	return iter.t.tagsForNode(ret, uint(iter.nodeIndex), nil)
}

// Delete a tag from the current node if it matches matchVal, as
// determined by matchFunc. Returns how many tags are removed
// - use DeleteWithBuffer if you can reuse slices, to cut down on allocations
func (iter *TreeIteratorV4) Delete(matchFunc MatchesFunc, matchVal GeneratedType) int {
	return iter.DeleteWithBuffer(nil, matchFunc, matchVal)
}

// DeleteWithBuffer a tag from the current node if it matches
// matchVal, as determined by matchFunc. Returns how many tags are
// removed
// - uses input slice to reduce allocations
func (iter *TreeIteratorV4) DeleteWithBuffer(buf []GeneratedType, matchFunc MatchesFunc, matchVal GeneratedType) int {
	deleteCount, remainingTagCount := iter.t.deleteTag(buf, iter.nodeIndex, matchVal, matchFunc)
	if remainingTagCount > 0 || iter.nodeIndex == 1 {
		return deleteCount
	}
	nodeHistoryLen := len(iter.nodeHistory)
	currentIndex := iter.nodeIndex
	current := &iter.t.nodes[currentIndex]
	parentIndex := iter.nodeHistory[nodeHistoryLen-1]
	parent := &iter.t.nodes[parentIndex]
	wasLeft := false
	if parent.Left == currentIndex {
		wasLeft = true
	}
	result := iter.t.deleteNode(currentIndex, current, parentIndex, parent)
	switch result {
	case notDeleted:
		return deleteCount
	case deletedNodeReplacedByChild:
		// Continue with the child
		if wasLeft {
			iter.nodeIndex = parent.Left
		} else {
			iter.nodeIndex = parent.Right
		}
		iter.next = nextSelf
	case deletedNodeParentReplacedBySibling:
		iter.nodeIndex = parentIndex
		iter.nodeHistory = iter.nodeHistory[:nodeHistoryLen-1]
		if wasLeft {
			// Parent replaced by right sibling, to visit
			iter.next = nextSelf
		} else {
			// Parent replaced by left sibling, already visited
			iter.next = nextUp
		}
	case deletedNodeJustRemoved:
		iter.nodeIndex = parentIndex
		iter.nodeHistory = iter.nodeHistory[:nodeHistoryLen-1]
		if wasLeft {
			// Visit our sibling
			iter.next = nextRight
		} else {
			// Go up
			iter.next = nextUp
		}
	}
	return deleteCount
}

// note: this is only used for unit testing
// nolint
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

// note: this is only used for unit testing
// nolint
func (t *TreeV4) countTags(nodeIndex uint) int {
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
