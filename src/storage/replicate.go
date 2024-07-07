package main

import (
	common "commons"
	"context"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"node"
	"path/filepath"
	pb "protos"
	"time"
)

func ReplicateData(ctx context.Context, n *node.ChordNode, interval time.Duration) {
	var lastSuccessor *node.ChordNode = nil

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {

		// Find the successor of the node
		next := TakeBytesAndAdd1(n.Id)
		successor, err := n.FindSuccessor(ctx, next)

		if err != nil {
			log.Fatalf("Failed to find successor: %v", err)
		}

		if successor == nil {
			continue
		}

		if lastSuccessor == nil || successor.Address != lastSuccessor.Address {

			files, err := ioutil.ReadDir("./" + n.Address)
			if err != nil {
				log.Fatal(err)
			}

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

					// Get successor Id

					// Use the successor's ID as the key
					pair := &pb.Pair{
						Key:   hex.EncodeToString(next),
						Value: string(data),
					}

					ctx := context.Background()
					_, err = successor.Put(ctx, pair)
					if err != nil {
						log.Fatal(err)
					}

					// Log replicated data
					log.Printf("Replicated data with successor %v, last successor %v, node address %v\n", successor, lastSuccessor, n.Address)
				}
			}

			// Log replicated data with successor, last successor and node address
			log.Printf("Identified different successor with successor %v, last successor %v, node address %v\n", successor, lastSuccessor, n.Address)

			lastSuccessor = successor
		}
	}
}
