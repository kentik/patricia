// Package template is the base of code generation for type-specific trees
package uintptr_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []uintptr
}

func (n *treeNode) AddTag(tag uintptr) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []uintptr{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
