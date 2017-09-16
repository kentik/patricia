// Package template is the base of code generation for type-specific trees
package uint_ptr_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []*uint
}

func (n *treeNode) AddTag(tag *uint) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []*uint{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
