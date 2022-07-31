// Template file.

package template

// code common to the IPv4/IPv6 trees

// MatchesFunc is called to check if tag data matches the input value
type MatchesFunc func(payload GeneratedType, val GeneratedType) bool

// FilterFunc is called on each result to see if it belongs in the resulting set
type FilterFunc func(payload GeneratedType) bool

// treeIteratorNext is an indicator to know what Next() should return
// for the current node.
type treeIteratorNext int

const (
	nextSelf treeIteratorNext = iota
	nextLeft
	nextRight
	nextUp
)
