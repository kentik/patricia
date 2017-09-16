// Package template is the base of code generation for type-specific trees
package byte_array_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    [][]byte
}

func (n *treeNode) AddTag(tag []byte) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = [][]byte{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
