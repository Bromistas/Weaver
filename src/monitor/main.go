package main

import (
	"fmt"
	"github.com/bromistas/weaver-commons"
	"time"
)

func main() {
	node1 := common.NewNode(1)
	node2 := common.NewNode(2)
	node3 := common.NewNode(3)
	//node4 := common.NewNode(4)

	node2.Join(node1)
	node3.Join(node1)

	go node1.Stabilize()
	go node2.Stabilize()
	go node3.Stabilize()

	// Print everybody's list of nodes and maps
	time.Sleep(1 * time.Second)
	//for {
	fmt.Printf("Node 1: %v\n", node1)

	fmt.Printf("Node 2: %v\n", node2)
	fmt.Printf("Node 3: %v\n", node3)

}
