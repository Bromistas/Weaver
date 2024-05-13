package main

import (
	"fmt"
)

func main() {
	// Create a few nodes
	node1 := NewNode("node1", "192.168.1.1")
	node2 := NewNode("node2", "192.168.1.2")
	node3 := NewNode("node3", "192.168.1.3")

	// Create a new consistent hash ring
	ch := NewConsistentHash()

	// Add the nodes to the ring
	ch.AddNode(node1)
	ch.AddNode(node2)
	ch.AddNode(node3)

	// Test the GetNode method with a few keys
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	for _, key := range keys {
		node := ch.GetNode(key)
		fmt.Printf("Key: %s, Node: %s, IP: %s\n", key, node.name, node.ip)
	}

	// Test the RemoveNode method
	ch.RemoveNode(*node1)
	fmt.Println("After removing node1")

	// Test the GetNode method again
	for _, key := range keys {
		node := ch.GetNode(key)
		fmt.Printf("Key: %s, Node: %s, IP: %s\n", key, node.name, node.ip)
	}
}
