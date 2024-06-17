package common

import (
	"sync"
)

type Node struct {
	ID          uint64
	Successor   *Node
	Predecessor *Node
	Data        map[string]string
	mutex       sync.Mutex
}

func NewNode(id uint64) *Node {
	return &Node{
		ID:          id,
		Successor:   nil,
		Predecessor: nil,
		Data:        make(map[string]string),
	}
}

func (n *Node) Insert(newNode *Node) {
	//n.mutex.Lock()
	//defer n.mutex.Unlock()

	if n.Successor == nil {
		// If the node is alone, it points to itself
		n.Successor = newNode
		n.Predecessor = newNode
		newNode.Successor = n
		newNode.Predecessor = n
		return
	}

	// Find the correct position for the new node
	current := n
	for {
		if between(current.ID, newNode.ID, current.Successor.ID) {
			newNode.Successor = current.Successor
			newNode.Predecessor = current
			current.Successor.Predecessor = newNode
			current.Successor = newNode
			break
		}
		current = current.Successor
		if current == n {
			break
		}
	}

	// Stabilize after insertion
	n.Stabilize()
}

func between(start, id, end uint64) bool {
	if start < end {
		return start < id && id < end
	}
	return start < id || id < end
}

func (n *Node) Stabilize() {
	//n.mutex.Lock()
	//defer n.mutex.Unlock()

	x := n.Successor.Predecessor
	if between(n.ID, x.ID, n.Successor.ID) && x != n {
		n.Successor = x
	}
	n.Successor.Notify(n)
}

func (n *Node) Notify(node *Node) {
	//n.mutex.Lock()
	//defer n.mutex.Unlock()

	if n.Predecessor == nil || (between(n.Predecessor.ID, node.ID, n.ID) && node != n) {
		n.Predecessor = node
	}
}

func (n *Node) FindSuccessor(key uint64) *Node {
	node := n.FindPredecessor(key)
	return node.Successor
}

func (n *Node) FindPredecessor(key uint64) *Node {
	//n.mutex.Lock()
	//defer n.mutex.Unlock()

	if between(n.ID, key, n.Successor.ID) || n == n.Successor {
		return n
	}

	current := n
	for !between(current.ID, key, current.Successor.ID) && current.ID != key {
		current = current.Successor
	}

	return current
}
