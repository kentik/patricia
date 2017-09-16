// Package template is the base of code generation for type-specific trees
package uint8_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []uint8
}

func (n *treeNode) AddTag(tag uint8) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []uint8{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
