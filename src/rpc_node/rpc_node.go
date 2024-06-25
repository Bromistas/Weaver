package main

import (
	"context"
	"fmt"
	"node"
	pb "protos"
	"sync"
	"time"
)

const (
	address1 = "127.0.0.1:50051"
	address2 = "127.0.0.1:50052"
	address3 = "127.0.0.1:50053"
)

func main() {
	group := &sync.WaitGroup{}

	// Create three nodes
	node1 := node.NewChordNode(address1, nil)
	node2 := node.NewChordNode(address2, nil)
	node3 := node.NewChordNode(address3, nil)

	fmt.Println("Nodes created")

	// Start serving each node in a separate goroutine
	group.Add(1)
	go func() {
		fmt.Println("Node1 started")
		node.ServeChord(context.Background(), node1, nil, group, nil)
		fmt.Println("Node1 joined the network")
	}()

	group.Add(1)
	go func() {
		fmt.Println("Node2 started")
		node.ServeChord(context.Background(), node2, node1, group, nil)
		fmt.Println("Node2 joined the network")
	}()

	group.Add(1)
	go func() {
		fmt.Println("Node3 started")
		node.ServeChord(context.Background(), node3, node1, group, nil)
		fmt.Println("Node3 joined the network")
	}()

	time.Sleep(1 * time.Second)

	fmt.Println(node1.FindPredecessor(context.Background(), node1))
	fmt.Println(node2.FindPredecessor(context.Background(), node2))
	fmt.Println(node3.FindPredecessor(context.Background(), node3))

	// Insert a key into the chord ring
	key := "testKey"
	value := "testValue"
	pair := &pb.Pair{Key: key, Value: value}
	_, err := node1.Put(context.Background(), pair)
	if err != nil {
		fmt.Printf("Failed to insert key: %v\n", err)
	} else {
		fmt.Printf("Key %s inserted successfully\n", key)
	}

	group.Wait()

	// Print out the state of the network
	fmt.Println("Network state:")
	fmt.Println("Node1:", node1)
	fmt.Println("Node2:", node2)
	fmt.Println("Node3:", node3)
}
