package main

import (
	"bytes"
	"chord"
	common "commons"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"node"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var previousSuccessors map[string][]string

func init() {
	previousSuccessors = make(map[string][]string)
}

func lookupAndReplicateIfNecessary(ring *chord.Ring, product *common.Product) {
	currentSuccessors, err := ring.Lookup(3, []byte(product.Name))
	if err != nil {
		log.Fatalf("Lookup failed: %v", err)
	}

	currentSuccessorAddresses := make([]string, 3)
	for i, succ := range currentSuccessors {
		currentSuccessorAddresses[i] = succ.Host
	}

	previousSuccessorAddresses, found := previousSuccessors[product.Name]

	if !found || !equalSuccessors(previousSuccessorAddresses, currentSuccessorAddresses) {
		// Successors have changed or this is the first lookup
		for _, address := range currentSuccessorAddresses {
			err := replicateProduct(address, *product)
			if err != nil {
				return
			}
		}
		// Update the previous successors map
		previousSuccessors[product.Name] = currentSuccessorAddresses
	}
}

func equalSuccessors(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func replicateProduct(endpoint string, product common.Product) error {

	product.Replicated = true

	// Marshal the product into JSON
	payload, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product: %v", err)
	}

	// Create a new POST request with the JSON payload
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	// Print replicated data with endpoint
	log.Printf("Replicated data %s with endpoint %s\n", product.Name, endpoint)

	return nil
}

func replicateNewFiles(ctx context.Context, n *node.ChordNode, predecessor *node.ChordNode, files []os.FileInfo) {
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			data, err := ioutil.ReadFile(n.Address + "/" + file.Name())
			if err != nil {
				log.Fatal(err)
			}

			var product common.Product
			err = json.Unmarshal(data, &product)
			if err != nil {
				log.Fatal(err)
			}

			// Check if the product was created by the current node
			if product.NodeAuthor != n.Address {
				continue
			}

			if product.Replicated {
				continue
			}

			fmt.Printf("File eligible for replication: %s\n", product.Name)

			// Send it
			err = replicateProduct("http://"+predecessor.Address+"/replicate", product)
			if err != nil {
				fmt.Printf("Error while replicating %s to %s: %s", product.Name, predecessor.Address, err.Error())
				return
			}

			// Mark the product as replicated
			product.Replicated = true

			// Rewrite it
			updatedData, err := json.Marshal(product)
			if err != nil {
				log.Fatal(err)
			}
			err = ioutil.WriteFile(n.Address+"/"+file.Name(), updatedData, 0644)
			if err != nil {
				log.Fatal(err)
			}

			// Log replicated data
			log.Printf("Replicated file: %s with predecessor %v\n", file.Name(), predecessor.Address)
		}
	}
}

//func replicateOnPredChange(ctx context.Context, n *node.ChordNode, predecessor *node.ChordNode, files []os.FileInfo) {
//	for _, file := range files {
//		if filepath.Ext(file.Name()) == ".json" {
//			data, err := ioutil.ReadFile(n.Address + "/" + file.Name())
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			var product common.Product
//			err = json.Unmarshal(data, &product)
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			// Check if the product was created by the current node
//			if product.NodeAuthor != n.Address {
//				continue
//			}
//
//			err = replicateProduct("http://"+predecessor.Address, product)
//			if err != nil {
//				fmt.Printf("Error while replicating %s to %s: %s", product.Name, predecessor.Address, err.Error())
//				return
//			}
//
//			// Log replicated data
//			log.Printf("Replicated data with predecessor %v, node address %v\n", predecessor.Address, n.Address)
//		}
//	}
//}

var previousStorageNodes = make(map[string][]string)

func shouldReplicate(productName string, currentStorageNodes []string) bool {
	previousNodes, exists := previousStorageNodes[productName]
	if !exists {
		// If it's the first time, always replicate
		return true
	}

	if len(previousNodes) != len(currentStorageNodes) {
		// Different number of nodes, should replicate
		return true
	}

	// Check if all nodes are the same
	for _, node := range currentStorageNodes {
		if !contains(previousNodes, node) {
			return true
		}
	}

	// If all nodes are the same, no need to replicate
	return false
}

func blackholeReplicate(ring *chord.Ring, host string, files []os.FileInfo) {
	// Get all the storage nodes

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			data, err := os.ReadFile(host + "/" + file.Name())
			if err != nil {
				log.Fatal(err)
			}

			var product common.Product
			err = json.Unmarshal(data, &product)
			if err != nil {
				log.Fatal(err)
			}

			storageNodes := getAllPredecessors(ring, []byte(product.Name))

			if shouldReplicate(product.Name, storageNodes) {

				previousStorageNodes[product.Name] = storageNodes[:]

				for _, address := range storageNodes {

					fullEndpoint := "http://" + address + ":10001/replicate"
					err = replicateProduct(fullEndpoint, product)
					if err != nil {
						fmt.Printf("Error while replicating %s to %s: %s\n", product.Name, fullEndpoint, err.Error())
						continue
					}

					// Log replicated data
					log.Printf("Replicated data with node %v, node address %v\n", address, host)
				}
			} else {
				log.Printf("No need for replication on %s\n", product.Name)
			}

		}
	}

}

func ReplicateData(ctx context.Context, n *chord.Ring, address string, interval time.Duration) {

	//lastPredecessor := &node.ChordNode{}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		files, err := ioutil.ReadDir("./" + address)
		if err != nil {
			log.Fatal(err)
		}

		// blachole replicate
		blackholeReplicate(ring, address, files)
	}
}

func getAllPredecessors(ring *chord.Ring, key []byte) []string {
	present := make(map[string]bool)
	predecessors := make([]string, 0)
	for _, n := range ring.Vnodes {
		// Ping the predecessor if not nil plase
		if n.Predecessor != nil {
			ok, err := ring.Transport.Ping(n.Predecessor)
			if err == nil && !present[n.Predecessor.Host] && ok {
				present[n.Predecessor.Host] = true
				predecessors = append(predecessors, n.Predecessor.Host)
			}
		}
	}

	// Call lookup to get the closest nodes
	successors, err := ring.Lookup(8, key)
	if err != nil {
		log.Printf("Lookup failed: %v", err)
	}

	for _, succ := range successors {

		if !present[succ.Host] {
			present[succ.Host] = true
			predecessors = append(predecessors, succ.Host)
		}
	}

	// Just ot be safe go through every predecessor, split by : and get the first part
	for i, pred := range predecessors {
		predecessors[i] = strings.Split(pred, ":")[0]
	}

	return predecessors
}
