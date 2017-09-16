// Package template is the base of code generation for type-specific trees
package int_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []int
}

func (n *treeNode) AddTag(tag int) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []int{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
