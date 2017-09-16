// Package template is the base of code generation for type-specific trees
package float64_array_tree

// treeNode represents a 128-bit node in the Patricia tree
type treeNode struct {
	HasTags bool
	Tags    [][]float64
}

func (n *treeNode) AddTag(tag []float64) {
	n.HasTags = true
	if n.Tags == nil {
		n.Tags = [][]float64{tag}
	} else {
		n.Tags = append(n.Tags, tag)
	}
}
