Garbage collector-sensitive patricia tree for IP/CIDR tagging
=============================================================

![CI](https://github.com/kentik/patricia/workflows/CI/badge.svg)
[![GitHub Release](https://img.shields.io/github/release/kentik/patricia.svg?style=flat)](https://github.com/kentik/patricia/releases/latest)
[![Coverage Status](https://coveralls.io/repos/github/kentik/patricia/badge.svg?branch=main)](https://coveralls.io/github/kentik/patricia?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/kentik/patricia)](https://goreportcard.com/report/github.com/kentik/patricia)

What is this?
-------------

A Go implemenation of a [patricia tree](https://en.wikipedia.org/wiki/Radix_tree) (radix tree with radix=2), specifically for
tagging IPv4 and IPv6 addresses with CIDR bits, with a focus on producing as little garbage for the garbage collector to
manage as possible. This allows you to tag millions of IP addresses without incurring a penalty during GC scanning.

This library requires Go >= 1.9.

IP/CIDR tagging
---------------

IP addresses can be tagged by any of the built-in types that we generate trees for. It's no accident that we don't support
pointers, slices, or `interface{}` for reasons described below. Once your IPv4 or IPv6 tree is initialized, you can tag a full
32/128 bit address, or IP/CIDR.

For example, on an IPv4 tree, you can create the following tags:

- `123.0.0.0/8`:     `"HELLO"`
- `123.54.66.20/32`: `"THERE"`
- `123.54.66.0/24`:  `"GOPHERS"`
- `123.54.66.21/32`: `":)"`

Searching for:

- `123.1.2.3/32` or `123.0.0.0/8` returns `"HELLO"`
- `123.54.66.20/32` returns `["HELLO", "THERE", "GOPHERS"]`
- `123.54.66.21/32` returns `["HELLO", "GOPHERS", ":)"]`


Generated types, but why not reference types?
---------------------------------------------

The initial version of this effort included many references to structs. The nodes in the tree were all tied together with pointers,
and each node had an array of tags. Even as the tree grew, it seemed to perform well. However, CPU was higher than expected. Profiling
revealed this to be largely due to garbage collection. Even though the objects in the tree were mostly static, each one needed to be 
scanned by the garbage collector when it ran. The strategy then became: _remove all pointers possible_. 

At this point, the internal structure is tuned to be left alone by the garbage collector. Storing references in the tree would defeat 
much of the purpose of these optimizations. If you need to store references, then consider storing integers that index the data you need
in a separate structure, like a map or array.

In addition, to support custom payload types would require `interface{}`, which adds noticeable overhead at high volume. To avoid this,
a separate set of strongly-typed trees is generated for:

- `bool`
- `byte`
- `complex64`
- `complex128`
- `float32`
- `float64`
- `int8`
- `int16`
- `int32`
- `int64`
- `rune`
- `string`
- `uint`
- `uint8`
- `uint16`
- `uint32`
- `uint64`


How does this avoid garbage collection scanning?
------------------------------------------------

A scarcely-populated patricia tree will require about 2x as many nodes as addresses, and each node with tags needs to maintain that list.
This means, in a pointer-based tree of 1 million IP addresses, you'll end up with around 3 million references - this puts a considerable
load on the garbage collector.

To avoid this, the nodes in this tree are stored in a single array, by value. This array of nodes is a single reference that the GC 
needs to manage. Nodes are wired together by `uint32` indexes in that array. This has the added benefit of saving us 8 bytes
of memory per node: rather than two 64-bit pointers, we have two 32-bit integers.

The way we avoid a reference to each collection of tags is a little trickier. Thanks to an [optimization introduced in 1.5](https://github.com/golang/go/issues/9477),
the GC now ignores maps with keys and values that do not contain pointers. So, our collection of tags is flattened into a `map[uint64]GENERATED_TYPE`.
The keys into the map use the following convention:

        (nodeIndex << 32) + (tagArrayIndex...)

That is... we use a 64 bit number, setting the most significant 32 bits to the node index, then adding to it the 0-based index into the
tag array. 

With these strategies, in a tree of 1 million tags, we reduce the pointer count from 3 million to 3: the tree, its node array, 
and its tag map. Your garbage collector thanks you.


Notes
-----

- This is not thread-safe. If you need concurrency, it needs to be managed at a higher level.
- The tree is tuned for fast reads, but update performance shouldn't be too bad.
- IPv4 addresses are represented as uint32
- IPv6 addresses are represented as a pair of uint64's
- The tree maintains as few nodes as possible, deleting unnecessary ones when possible, to reduce the amount of work needed during tree search.
- The tree doesn't currently compact its array of nodes, so you could end up with a capacity that's twice as big as the max number of nodes ever seen, but 
each node is only 20 bytes. Deleted node indexes are reused.
- Code generation isn't performed with `go generate`, but rather a Makefile with some simple search and replace from the ./template directory. Development
is performed on the IPv4 tree. The IPv6 tree is generated from it, again, with simple search & replaces. 
