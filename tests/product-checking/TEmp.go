package main

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
)

var ctx = context.Background()

func main() {

	product := Product{
		Name:        "Test Product 2",
		Price:       100.0,
		URL:         "http://example.com",
		Description: "This is a test product",
		Rating:      "4 stars",
	}

	Check("localhost:8080", product)
	Check("localhost:8070", product)

}

func Check(addr string, product Product) {
	jsonProduct, err := json.Marshal(product)
	if err != nil {
		log.Printf("Failed to marshal product: %s", err)
		return
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Verify the product was inserted correctly
	result, err := client.Get(ctx, product.Name).Result()
	if err != nil {
		log.Printf("Failed to get product: %s", err)
		return
	}

	if result != string(jsonProduct) {
		log.Printf("Product data mismatch: got %s, want %s", result, string(jsonProduct))
	} else {
		log.Printf("Product inserted correctly")
	}
}
