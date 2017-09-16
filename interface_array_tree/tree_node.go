// Package template is the base of code generation for type-specific trees
package interface_array_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    [][]interface{}
}

func (n *treeNode) AddTag(tag []interface{}) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = [][]interface{}{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
