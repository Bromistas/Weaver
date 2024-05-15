package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const STORAGE_PATH = "storage"

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	failOnError(err, "Failed to read request body")

	var product Product
	err = json.Unmarshal(body, &product)
	failOnError(err, "Failed to unmarshal product")

	// Do something with the product
	log.Printf("Received product: %+v", product)

	// Save the product to a JSON file
	productJson, err := json.MarshalIndent(product, "", "  ")
	failOnError(err, "Failed to marshal product to JSON")

	dir := fmt.Sprintf("%s/network-%s", STORAGE_PATH, r.Host)
	err = os.MkdirAll(dir, 0755)
	failOnError(err, "Failed to create directory")

	filePath := filepath.Join(dir, fmt.Sprintf("%s.json", product.Name))
	err = os.WriteFile(filePath, productJson, 0644)
	failOnError(err, "Failed to write product to file")

	w.WriteHeader(http.StatusOK)
}

func startServer(addr string, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("[*] Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func main() {
	var wg sync.WaitGroup

	wg.Add(2)

	http.HandleFunc("/product", productHandler)

	go startServer(":8080", &wg)
	go startServer(":8070", &wg)

	wg.Wait()
}
