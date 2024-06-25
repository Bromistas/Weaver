package main

import (
	"commons"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"node"
	"os"
	pb "protos"
	"sync"
	"time"
)

func put_pair(addr, k, v string, group *sync.WaitGroup) {
	defer group.Done()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("connection to %v failed: %v", addr, err)
	}
	defer conn.Close()
	c := pb.NewDHTClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	_, err = c.Put(ctx, &pb.Pair{Key: k, Value: v})
	if err != nil {
		log.Fatalf("could not put to %v: %v", addr, err)
	}
}

// CustomPut is a function that unmarshals the payload into a Product object and writes it to a JSON file
func CustomPut(ctx context.Context, pair *pb.Pair) error {
	// Unmarshal the payload into a Product object

	fmt.Println("Printing custom put")

	var product common.Product
	err := json.Unmarshal([]byte(pair.Value), &product)
	if err != nil {
		log.Fatalf("Failed to unmarshal payload: %v", err)
		return err
	}

	// Write the Product object to a JSON file
	productJson, err := json.Marshal(product)
	if err != nil {
		log.Fatalf("Failed to marshal product: %v", err)
		return err
	}

	filename := product.Name + ".json"
	err = os.WriteFile(filename, productJson, 0644)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
		return err
	}

	log.Printf("Product written to file: %v", product)

	return nil
}

func main() {
	group := &sync.WaitGroup{}
	node1 := node.NewChordNode("127.0.0.1:50051", CustomPut)

	// Start serving each node in a separate goroutine
	group.Add(1)
	go func() {
		fmt.Println("Node1 started")
		node.ServeChord(context.Background(), node1, nil, group, nil)
		fmt.Println("Node1 joined the network")
	}()

	time.Sleep(1 * time.Second)

	key := "product1"
	value := "{\"id\": \"123\", \"name\": \"the\", \"price\":1444}"
	put_pair(node1.Address, key, value, group)

	// Test querying for product1

	group.Wait()
}
