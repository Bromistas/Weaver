package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"src/common"
)

var consistentHash *ConsistentHash

func init() {
	// Initialize the consistent hash object here.
	// Add your nodes to the consistent hash.
	consistentHash = NewConsistentHash()
	consistentHash.AddNode(NewNode("node1", "http://localhost:8080"))
	// consistentHash.AddNode(NewNode("node2", "http://localhost:8082"))
	// Add as many nodes as you have in your system.
}

func Route(product common.Product) {
	node := consistentHash.GetNode(product.Name)

	jsonProduct, err := json.Marshal(product)
	if err != nil {
		log.Printf("Failed to marshal product: %s", err)
		return
	}

	resp, err := http.Post(node.ip+"/product", "application/json", bytes.NewBuffer(jsonProduct))
	if err != nil {
		log.Printf("Failed to route product: %s", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-OK response while routing product: %d", resp.StatusCode)
	}
}
