package main

import (
	"bytes"
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
	"time"
)

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
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

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

func replicateOnPredChange(ctx context.Context, n *node.ChordNode, predecessor *node.ChordNode, files []os.FileInfo) {
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

			err = replicateProduct("http://"+predecessor.Address, product)
			if err != nil {
				fmt.Printf("Error while replicating %s to %s: %s", product.Name, predecessor.Address, err.Error())
				return
			}

			// Log replicated data
			log.Printf("Replicated data with predecessor %v, node address %v\n", predecessor.Address, n.Address)
		}
	}
}

func ReplicateData(ctx context.Context, n *node.ChordNode, interval time.Duration) {
	var lastPredecessor = &node.ChordNode{Address: n.Address}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		predecessor, err := n.FindPredecessor(ctx, n)

		if err != nil {
			log.Fatalf("Failed to find successor: %v", err)
		}

		if predecessor == nil {
			continue
		}

		files, err := ioutil.ReadDir("./" + n.Address)
		if err != nil {
			log.Fatal(err)
		}

		// If predecessor change then replicate, else only replicate new files
		if predecessor.Address != lastPredecessor.Address {
			log.Printf("Identified different predecessor. Current: %v, Last: %v, Node address: %v\n", predecessor.Address, lastPredecessor.Address, n.Address)
			replicateOnPredChange(ctx, n, predecessor, files)
			lastPredecessor = predecessor
		} else {
			replicateNewFiles(ctx, n, predecessor, files)
		}
	}
}
