package main

import (
	"chord"
	common "commons"
	"fmt"
	"log"
)

func insertInStore(ring *chord.Ring, product common.Product, host string) {
	key := []byte(product.Name)
	closestAddr := lookupKey(ring, key, host)

	err := SendProductRequest(product, closestAddr)
	if err != nil {
		fmt.Printf("Error while sending the insertion request for %s: %s", product.Name, err)
		return
	}
}

// Function to look for a key in the ring
func lookupKey(ring *chord.Ring, key []byte, host string) string {
	successors, err := ring.Lookup(1, key) // Assuming you want the closest successor
	if err != nil {
		log.Fatalf("Lookup failed: %v", err)
	}
	for _, succ := range successors {
		fmt.Printf("Successor ID: %s\n", succ.Id)
	}

	return successors[0].Host
}
