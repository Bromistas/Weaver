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
	var lastPredecessor *node.ChordNode = nil

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {

		// Find the successor of the node
		predecessor, err := n.FindPredecessor(ctx, n)

		if err != nil {
			log.Fatalf("Failed to find successor: %v", err)
		}

		if predecessor == nil {
			continue
		}

		if lastPredecessor == nil || predecessor.Address != lastPredecessor.Address {

			log.Printf("Identified different predecessor with predecessor %v, last predecessor %v, node address %v\n", predecessor, lastPredecessor, n.Address)

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

					// Get predecessor Id

					// Use the predecessor's ID as the key
					next := TakeBytesAndTake1(predecessor.Id)

					pair := &pb.Pair{
						Key:   hex.EncodeToString(next),
						Value: string(data),
					}

					ctx := context.Background()
					_, err = predecessor.Put(ctx, pair)
					if err != nil {
						log.Fatal(err)
					}

					// Log replicated data
					log.Printf("Replicated data with predecessor %v, last predecessor %v, node address %v\n", predecessor, lastPredecessor, n.Address)
				}
			}

			lastPredecessor = predecessor
		}
	}
}
