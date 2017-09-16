// Package template is the base of code generation for type-specific trees
package int8_array_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    [][]int8
}

func (n *treeNode) AddTag(tag []int8) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = [][]int8{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
