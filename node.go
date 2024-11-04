package hashy

import "time"

// Node is a host that an arbitrary string may hash to.
type Node struct {
	// Name is a hostname at this node.
	Name string

	// Group is the logical group this node belongs to.  Most
	// commonly, group corresponds to datacenter.
	Group string

	// TTL is the DNS time-to-live for this node entry.
	TTL time.Duration
}

// Nodes is an aggregate of nodes.
type Nodes []Node

// Append returns a possibly different Nodes with the
// given nodes appended.
func (ns Nodes) Append(more ...Node) Nodes {
	return append(ns, more...)
}
