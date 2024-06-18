package common

type Node struct {
	ID     int
	IDList []int
	IDMap  map[int]*Node
}

func NewNode(id int) *Node {
	n := &Node{
		ID:     id,
		IDList: []int{},
		IDMap:  make(map[int]*Node),
	}

	n.IDList = append(n.IDList, n.ID)
	n.IDMap[n.ID] = n

	return n
}

func (n *Node) Insert(node *Node) {
	if n.IDMap[node.ID] == nil {
		n.IDList = insertIntoSorted(n.IDList, node.ID)
		n.IDMap[node.ID] = node
		n.notifyInsertion(node)
	}
}

func (n *Node) GetMap() map[int]*Node {
	//TODO
	return n.IDMap
}

func (n *Node) Join(node *Node) {
	node.Insert(n)
	for _, id := range node.GetMap() {
		n.Insert(id)
	}
}

func (n *Node) notifyInsertion(node *Node) {
	// TODO
	for _, id := range n.IDMap {
		id.Insert(node)
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
	// TODO
	if n.IDMap[node.ID] != nil {
		return true
	} else {
		return false
	}
}

func (n *Node) notifyRemoval(node *Node) {
	// TODO
	for _, item := range n.IDMap {
		item.Remove(node)
	}
}

func (n *Node) Remove(node *Node) {
	if _, ok := n.IDMap[node.ID]; ok {
		delete(n.IDMap, node.ID)

		for _, id := range n.IDList {
			if id == node.ID {
				n.IDList = removeFromSorted(n.IDList, node.ID)
				break
			}
		}
	}
}
