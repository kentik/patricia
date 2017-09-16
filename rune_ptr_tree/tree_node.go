// Package template is the base of code generation for type-specific trees
package rune_ptr_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []*rune
}

func (n *treeNode) AddTag(tag *rune) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []*rune{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
