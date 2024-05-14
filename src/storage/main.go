package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"src/common"
)

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

	var product common.Product
	err = json.Unmarshal(body, &product)
	failOnError(err, "Failed to unmarshal product")

	// Do something with the product
	log.Printf("Received product: %+v", product)

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/product", productHandler)

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
