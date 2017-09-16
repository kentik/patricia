// Package template is the base of code generation for type-specific trees
package complex64_ptr_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []*complex64
}

func (n *treeNode) AddTag(tag *complex64) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []*complex64{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
