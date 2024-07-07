package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Product struct {
	Name        string
	Price       float64
	URL         string
	Description string
	Rating      string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type URLMessage struct {
	URL     string
	URLType URLType
}

type URLType int

const (
	AmazonProduct URLType = iota
	NeweggProduct
	NeweggRoot
)

func main() {
	message := "https://primeng.org/installation"
	baseUrl := "http://localhost:9000"
	url := fmt.Sprintf("%s/put", baseUrl)
	body := map[string]string{"message": message}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		panic(err)
	}
}
