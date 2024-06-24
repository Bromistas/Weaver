package common

import "sync"

type Node struct {
	ID      int
	Address string
	IDList  []int
	IDMap   map[int]string
	Mutex   sync.RWMutex
}

type NodeFacet struct {
	ID      int
	Address string
}

func NewNode(id int, address string) *Node {
	n := &Node{
		ID:      id,
		Address: address,
		IDList:  []int{},
		IDMap:   make(map[int]string),
	}

	n.IDList = append(n.IDList, n.ID)
	n.IDMap[n.ID] = n.Address

	return n
}

func (n *Node) Insert(node *Node) {
	n.Mutex.Lock()
	defer n.Mutex.Unlock()

	if n.IDMap[node.ID] == "" {
		n.IDList = insertIntoSorted(n.IDList, node.ID)
		n.IDMap[node.ID] = node.Address
		//n.notifyInsertion(node)
	}
}

func (n *Node) GetMap() map[int]string {
	return n.IDMap
}

func (n *Node) Remove(node *Node) {
	n.Mutex.Lock()
	defer n.Mutex.Unlock()

	if _, ok := n.IDMap[node.ID]; ok {
		delete(n.IDMap, node.ID)
		n.IDList = removeFromSorted(n.IDList, node.ID)
	}
}

func (n *Node) Lookup(key int) string {
	i := searchInSorted(n.IDList, key)
	i -= 1
	return n.IDMap[n.IDList[i]]
}
