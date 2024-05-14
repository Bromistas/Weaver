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

type CaptainNode struct {
	Node
	replicationNodes []Node
}

type ConsistentHash struct {
	nodes map[uint64]CaptainNode
	ring  []uint64
}

func hash(s string) uint64 {
	hash := sha256.Sum256([]byte(s))
	return binary.BigEndian.Uint64(hash[:])
}

func (c *CaptainNode) AddReplication(node Node) {
	c.replicationNodes = append(c.replicationNodes, node)
}

func NewCaptainNode(name string, address string) *CaptainNode {
	return &CaptainNode{
		Node: Node{
			name: name,
			hash: hash(name),
			ip:   address,
		},
		replicationNodes: make([]Node, 0),
	}
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
		nodes: make(map[uint64]CaptainNode),
		ring:  make([]uint64, 0),
	}
}

func (c *ConsistentHash) AddNode(node *CaptainNode) {
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

func (c *ConsistentHash) GetNode(key string) CaptainNode {
	hash := hash(key)
	idx := sort.Search(len(c.ring), func(i int) bool {
		return c.ring[i] >= hash
	})

	if idx == len(c.ring) {
		idx = 0
	}

	return c.nodes[c.ring[idx]]
}
