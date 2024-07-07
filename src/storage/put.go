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

	if product.NodeAuthor == "" /* Then the product is not yet assigned to a node */ {
		product.NodeAuthor = node.Address

		// Find the successor and replicate the product to the successor

		// TODO: Replication on insertion
		//next := TakeBytesAndAdd1(node.Id)
		//successor, err := node.FindSuccessor(ctx, next)
		//if err != nil {
		//	log.Fatalf("Failed to find successor: %v", err)
		//	return err
		//}
		//
		//if successor.Address != node.Address {
		//	// Replicate the product to the successor
		//	_, err = node.Put(ctx, &pb.Pair{Key: pair.Key, Value: pair.Value})
		//	if err != nil {
		//		log.Fatalf("Failed to replicate product: %v", err)
		//		return err
		//	}
		//}

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
