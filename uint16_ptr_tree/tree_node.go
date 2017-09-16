// Package template is the base of code generation for type-specific trees
package uint16_ptr_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []*uint16
}

func (n *treeNode) AddTag(tag *uint16) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []*uint16{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
