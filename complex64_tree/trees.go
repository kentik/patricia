// Code generated by automation. DO NOT EDIT

package complex64_tree

// code common to the IPv4/IPv6 trees

// MatchesFunc is called to check if tag data matches the input value
type MatchesFunc func(payload complex64, val complex64) bool

// FilterFunc is called on each result to see if it belongs in the resulting set
type FilterFunc func(payload complex64) bool

// UpdatesFunc is called to update the tag value
type UpdatesFunc func(payload complex64) complex64

// treeIteratorNext is an indicator to know what Next() should return
// for the current node.
type treeIteratorNext int

const (
	nextSelf treeIteratorNext = iota
	nextLeft
	nextRight
	nextUp
)
