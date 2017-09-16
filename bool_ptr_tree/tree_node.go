// Package template is the base of code generation for type-specific trees
package bool_ptr_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []*bool
}

func (n *treeNode) AddTag(tag *bool) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []*bool{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
