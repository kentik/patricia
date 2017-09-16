// Package template is the base of code generation for type-specific trees
package string_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []string
}

func (n *treeNode) AddTag(tag string) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []string{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
