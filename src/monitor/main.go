package main

import (
	"fmt"
	"github.com/bromistas/weaver-commons"
	"time"
)

func main() {
	node1 := common.NewNode(1)
	node2 := common.NewNode(3)
	node3 := common.NewNode(5)
	node4 := common.NewNode(7)

	node2.Join(node1)
	node3.Join(node1)
	node4.Join(node2)

	go node1.Stabilize()
	go node2.Stabilize()
	go node3.Stabilize()

	// Print everybody's list of nodes and maps
	time.Sleep(1 * time.Second)
	//for {
	fmt.Printf("Node 1: %v\n", node1)

	fmt.Printf("Node 2: %v\n", node2)
	fmt.Printf("Node 3: %v\n", node3)
	fmt.Printf("Node 4: %v\n", node4)

	fmt.Printf("Lookup for 2 :%d\n", node1.Lookup(2).ID)
	fmt.Printf("Lookup for 4 :%d\n", node1.Lookup(4).ID)
	fmt.Printf("Lookup for 5 :%d\n", node1.Lookup(5).ID)
	fmt.Printf("Lookup for 10 :%d\n", node1.Lookup(10).ID)

}
