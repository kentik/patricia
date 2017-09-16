// Package template is the base of code generation for type-specific trees
package template

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    []GeneratedType
}

func (n *treeNode) AddTag(tag GeneratedType) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = []GeneratedType{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
