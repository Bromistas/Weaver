package common

import (
	"net"
	"strconv"
)

type Node struct {
	ID      int
	Address string
	IDList  []int
	IDMap   map[int]*Node
	ConnMap map[int]net.Conn
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
		IDMap:   make(map[int]*Node),
		ConnMap: make(map[int]net.Conn),
	}

	n.IDList = append(n.IDList, n.ID)
	n.IDMap[n.ID] = n

	return n
}

func (n *Node) Connect(node *Node) error {
	conn, err := net.Dial("tcp", node.Address)
	if err != nil {
		return err
	}

	n.ConnMap[node.ID] = conn
	return nil
}

func (n *Node) Insert(node *Node) {
	if n.IDMap[node.ID] == nil {
		n.IDList = insertIntoSorted(n.IDList, node.ID)
		n.IDMap[node.ID] = node
		//n.notifyInsertion(node)
	}
}

func (n *Node) GetMap() map[int]*Node {
	return n.IDMap
}

func (n *Node) notifyInsertion(node *Node) {
	for _, id := range n.IDMap {
		conn, ok := n.ConnMap[id.ID]
		if ok {
			_, _ = conn.Write([]byte("insert:" + strconv.Itoa(node.ID)))
		}
	}
}

func (n *Node) Stabilize() {
	for {
		for _, node := range n.IDMap {
			if !n.Ping(node) {
				n.Remove(node)
				n.notifyRemoval(node)
			}
		}
	}
}

func (n *Node) Ping(node *Node) bool {
	conn, ok := n.ConnMap[node.ID]
	if !ok {
		return false
	}

	_, err := conn.Write([]byte("ping"))
	if err != nil {
		return false
	}

	buf := make([]byte, 4)
	_, err = conn.Read(buf)
	if err != nil {
		return false
	}

	return string(buf) == "pong"
}

func (n *Node) notifyRemoval(node *Node) {
	for _, id := range n.IDMap {
		conn, ok := n.ConnMap[id.ID]
		if ok {
			_, _ = conn.Write([]byte("remove:" + strconv.Itoa(node.ID)))
		}
	}
}

func (n *Node) Remove(node *Node) {
	if _, ok := n.IDMap[node.ID]; ok {
		delete(n.IDMap, node.ID)
		n.IDList = removeFromSorted(n.IDList, node.ID)
	}
}

func (n *Node) Lookup(key int) *Node {
	i := searchInSorted(n.IDList, key)
	i -= 1
	return n.IDMap[n.IDList[i]]
}
