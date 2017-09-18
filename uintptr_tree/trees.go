package uintptr_tree

// code common to the IPv4/IPv6 trees

// MatchesFunc is called to check if tag data matches the input value
type MatchesFunc func(payload uintptr, val uintptr) bool

// FilterFunc is called on each result to see if it belongs in the resulting set
type FilterFunc func(payload uintptr) bool
