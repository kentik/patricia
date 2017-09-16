// Package template is the base of code generation for type-specific trees
package int32_ptr_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []*int32
}

func (n *treeNode) AddTag(tag *int32) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []*int32{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
