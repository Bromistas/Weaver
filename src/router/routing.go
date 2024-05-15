package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

var consistentHash *ConsistentHash

func init() {
	// Initialize the consistent hash object here.
	// Add your nodes to the consistent hash.

	consistentHash = NewConsistentHash()
	captain := NewCaptainNode("node1", "http://localhost:8080")
	captain.AddReplication(*NewNode("node2", "http://localhost:8070"))
	consistentHash.AddNode(captain)

	// consistentHash.AddNode(NewNode("node2", "http://localhost:8082"))
	// Add as many nodes as you have in your system.
}

func Send(product Product, addr string) {
	jsonProduct, err := json.Marshal(product)
	if err != nil {
		log.Printf("Failed to marshal product: %s", err)
		return
	}

	resp, err := http.Post(addr+"/product", "application/json", bytes.NewBuffer(jsonProduct))
	if err != nil {
		log.Printf("Failed to route product: %s", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-OK response while routing product: %d", resp.StatusCode)
	}
}

func Route(product Product) {
	node := consistentHash.GetNode(product.Name)

	Send(product, node.ip)

	// Send to replicas
	for _, replica := range node.replicationNodes {
		Send(product, replica.ip)
	}
}
