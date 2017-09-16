// Package template is the base of code generation for type-specific trees
package uint32_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []uint32
}

func (n *treeNode) AddTag(tag uint32) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []uint32{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
