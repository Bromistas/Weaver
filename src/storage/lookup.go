package main

import (
	"chord"
	common "commons"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func insertInStore(ring *chord.Ring, product common.Product, host string, amount int) {
	key := []byte(product.Name)
	successors := lookupKey(ring, key, host, amount)

	for _, closestAddr := range successors {

		ip := strings.Split(closestAddr, ":")[0]
		port := strings.Split(closestAddr, ":")[1]
		intPort, err := strconv.Atoi(port)
		if err != nil {
			fmt.Printf("Error while converting port to int: %s", err)
			return
		}

		product.Replicated = true
		err = SendProductRequest(product, ip+":"+strconv.Itoa(intPort+1))
		if err != nil {
			fmt.Printf("Error while sending the insertion request for %s: %s", product.Name, err)
		}
	}

}

// Function to look for a key in the ring
func lookupKey(ring *chord.Ring, key []byte, host string, amount int) []string {
	successors, err := ring.Lookup(8, key) // Request 8 successors
	if err != nil {
		log.Fatalf("Lookup failed: %v", err)
	}

	uniqueSuccessors := make(map[string]bool)
	var result []string

	for _, succ := range successors {
		// Print them
		log.Printf("Successor for key %s: %s", string(key), succ.Host)
	}

	for _, succ := range successors {
		if !uniqueSuccessors[succ.Host] {
			uniqueSuccessors[succ.Host] = true
			result = append(result, succ.Host)
		}
		if len(result) == amount {
			break
		}
	}

	if len(result) == 0 {
		log.Fatalf("No unique successors found")
	}

	// Print the unique successors
	for _, host := range result {
		fmt.Printf("Successor Host: %s\n", host)
	}

	if len(result) < amount && !contains(result, host) {
		result = append(result, host)
	}

	return result // Return the first unique successor
}

func contains(a []string, v string) bool {
	for _, b := range a {
		if b == v {
			return true
		}
	}
	return false
}

func write(product common.Product) {
	// Generate a filename based on the current timestamp
	filename := fmt.Sprintf("%s.json", product.Name)
	fp := filepath.Join(addr, filename)

	// Convert the product to JSON
	data, err := json.MarshalIndent(product, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal product on insert: %v", err)
		return
	}

	// Write the JSON content to the file
	err = os.WriteFile(fp, data, 0644)
	if err != nil {
		log.Printf("Failed to write file: %v", err)
		return
	}

	// Respond to the client
	log.Printf("File replicated successfully: %s\n", filename)
}
