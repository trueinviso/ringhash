package sortedslice

import (
	"cmp"
	"crypto/md5"
	"encoding/binary"
	"slices"
	"strconv"
)

// I want to test read heavy hot keys across nodes and then write heavy hot keys
// across nodes.
// I want to test quorum protocols, probably build raft from the MIT course.
// Probably the popular concert idea as well with millions of people trying to buy a seat

// For this data structure we can build it O(n), O(n^2) lookups and test with load
// Let's start with proxy pattern then move to gossip protocol like Dynamo
// I also want to test no Vnodes vs a lot etc
type PhysicalNode struct {
	ID      string
	Address string
}

type VirtualNode struct {
	Hash uint32
	*PhysicalNode
}

type HashRing struct {
	ring []VirtualNode
}

func (h *HashRing) Insert(node *PhysicalNode, vnodeCount int) {
	var hash uint32
	for i := 0; i < vnodeCount; i++ {
		checksum := md5.Sum([]byte(node.ID + "#" + strconv.Itoa(i)))
		hash = binary.BigEndian.Uint32(checksum[:4])
		vnode := VirtualNode{
			Hash:         hash,
			PhysicalNode: node,
		}
		h.ring = append(h.ring, vnode)
	}
	slices.SortFunc(h.ring, func(a, b VirtualNode) int {
		return cmp.Compare(a.Hash, b.Hash)
	})
}

func (h *HashRing) Get(key uint32) *PhysicalNode {
	pos := h.findNextNodePosition(key)

	if pos >= len(h.ring) {
		pos = 0
	}
	return h.ring[pos].PhysicalNode
}

func (h *HashRing) findNextNodePosition(key uint32) int {
	l := 0
	r := len(h.ring)

	var m int
	var curr VirtualNode
	for l < r {
		m = (l + r) / 2
		curr = h.ring[m]

		if curr.Hash < key {
			l = m + 1
		} else {
			r = m
		}
	}

	return l
}
