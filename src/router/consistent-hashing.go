package main

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"
)

type Node struct {
	name string
	hash uint64
	ip   string
}

type ConsistentHash struct {
	nodes map[uint64]Node
	ring  []uint64
}

func hash(s string) uint64 {
	hash := sha256.Sum256([]byte(s))
	return binary.BigEndian.Uint64(hash[:])
}

func NewNode(name string, address string) *Node {
	return &Node{
		name: name,
		hash: hash(name),
		ip:   address,
	}
}

func NewConsistentHash() *ConsistentHash {
	return &ConsistentHash{
		nodes: make(map[uint64]Node),
		ring:  make([]uint64, 0),
	}
}

func (c *ConsistentHash) AddNode(node *Node) {
	c.nodes[node.hash] = *node
	c.ring = append(c.ring, node.hash)
	sort.Slice(c.ring, func(i, j int) bool {
		return c.ring[i] < c.ring[j]
	})
}

func (c *ConsistentHash) RemoveNode(node Node) {
	delete(c.nodes, node.hash)
	for i := 0; i < len(c.ring); i++ {
		if c.ring[i] == node.hash {
			c.ring = append(c.ring[:i], c.ring[i+1:]...)
			break
		}
	}
}

func (c *ConsistentHash) GetNode(key string) Node {
	hash := hash(key)
	idx := sort.Search(len(c.ring), func(i int) bool {
		return c.ring[i] >= hash
	})

	if idx == len(c.ring) {
		idx = 0
	}

	return c.nodes[c.ring[idx]]
}
