// Package template is the base of code generation for type-specific trees
package complex128_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []complex128
}

func (n *treeNode) AddTag(tag complex128) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []complex128{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
