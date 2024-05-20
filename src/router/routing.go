package main

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
	"strings"
)

var consistentHash *ConsistentHash
var ctx = context.Background()

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

	addr, _ = strings.CutPrefix(addr, "http://")

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err = client.Set(ctx, product.Name, jsonProduct, 0).Err()
	if err != nil {
		log.Printf("Failed to set product: %s", err)
		return
	}

	log.Printf("Product %s inserted correctly in node %s", product.Name, addr)
}

func Route(product Product) {
	node := consistentHash.GetNode(product.Name)

	Send(product, node.ip)

	// Send to replicas
	for _, replica := range node.replicationNodes {
		Send(product, replica.ip)
	}
}
