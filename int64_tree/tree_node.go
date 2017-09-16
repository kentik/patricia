// Package template is the base of code generation for type-specific trees
package int64_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []int64
}

func (n *treeNode) AddTag(tag int64) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []int64{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
