package main

import (
	common "commons"
	"context"
	"encoding/json"
	"log"
	"node"
	"os"
	pb "protos"
)

func CustomPut(ctx context.Context, pair *pb.Pair, node *node.ChordNode) error {
	// Unmarshal the payload into a Product object

	var product common.Product
	err := json.Unmarshal([]byte(pair.Value), &product)
	if err != nil {
		log.Fatalf("Failed to unmarshal payload: %v", err)
		return err
	}

	if product.NodeAuthor == "" {
		product.NodeAuthor = node.Address
	}

	// Write the Product object to a JSON file
	productJson, err := json.MarshalIndent(product, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal product: %v", err)
		return err
	}

	nodeName := node.Address

	// If nodeName folder doesnt exist, create it
	if _, err := os.Stat(nodeName); os.IsNotExist(err) {
		err = os.Mkdir(nodeName, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
			return err
		}
	}

	filename := nodeName + "/" + product.Name + ".json"
	err = os.WriteFile(filename, productJson, 0644)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
		return err
	}

	log.Printf("Product written to file: %v in %s", product, nodeName)

	return nil
}
