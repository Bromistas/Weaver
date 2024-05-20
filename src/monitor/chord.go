package main

import (
	"crypto/sha1"
	"math/big"
	"net"
	"sync"
)

type Node struct {
	ID   *big.Int
	Addr net.Addr
}

type Ring struct {
	Nodes []*Node
	Mutex sync.RWMutex
}

func NewNode(addr net.Addr) *Node {
	hash := sha1.New()
	hash.Write([]byte(addr.String()))
	id := big.NewInt(0).SetBytes(hash.Sum(nil))
	return &Node{
		ID:   id,
		Addr: addr,
	}
}

func (r *Ring) AddNode(node *Node) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.Nodes = append(r.Nodes, node)
}

func (r *Ring) RemoveNode(node *Node) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	for i, n := range r.Nodes {
		if n.Addr.String() == node.Addr.String() {
			r.Nodes = append(r.Nodes[:i], r.Nodes[i+1:]...)
			break
		}
	}
}

func (r *Ring) GetNodeForKey(key string) *Node {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()
	hash := sha1.New()
	hash.Write([]byte(key))
	id := big.NewInt(0).SetBytes(hash.Sum(nil))
	for _, node := range r.Nodes {
		if node.ID.Cmp(id) >= 0 {
			return node
		}
	}
	return r.Nodes[0]
}

func (r *Ring) JoinNetwork(addr net.Addr) {
	node := NewNode(addr)
	r.AddNode(node)
}
