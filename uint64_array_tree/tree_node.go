// Package template is the base of code generation for type-specific trees
package uint64_array_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    [][]uint64
}

func (n *treeNode) AddTag(tag []uint64) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = [][]uint64{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
