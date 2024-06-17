package main

import (
	"fmt"
	"github.com/bromistas/weaver-commons"
)

func main() {
	node1 := common.NewNode(1)
	node2 := common.NewNode(2)
	node3 := common.NewNode(3)
	//node4 := common.NewNode(4)

	node1.Insert(node2)
	//node1.Insert(node2)
	node1.Insert(node3)

	node1.Stabilize()
	node2.Stabilize()
	node3.Stabilize()
	//node4.Stabilize()

	//node1.Stabilize()
	//node2.Stabilize()
	//node3.Stabilize()
	//node4.Stabilize()

	fmt.Printf("Node 1: ID = %d, Successor = %d, Predecessor = %d\n", node1.ID, node1.Successor.ID, node1.Predecessor.ID)
	fmt.Printf("Node 2: ID = %d, Successor = %d, Predecessor = %d\n", node2.ID, node2.Successor.ID, node2.Predecessor.ID)
	fmt.Printf("Node 3: ID = %d, Successor = %d, Predecessor = %d\n", node3.ID, node3.Successor.ID, node3.Predecessor.ID)
	//fmt.Printf("Node 4: ID = %d, Successor = %d, Predecessor = %d\n", node4.ID, node4.Successor.ID, node4.Predecessor.ID)
}
