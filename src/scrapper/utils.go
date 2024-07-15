package main

import (
	"bytes"
	common "commons"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// sendProductRequest marshals the product and address, then sends them to the /replicate endpoint
func insertProduct(product common.Product, address string) error {
	endpoint := "http://" + address + "/insert"
	payloadBytes, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	maxRetries := 3
	backoff := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			if err == io.EOF && attempt < maxRetries-1 {
				time.Sleep(backoff)
				backoff *= 2 // Exponential backoff
				continue
			} else {
				return fmt.Errorf("failed to send HTTP request: %v", err)
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected response status: %s", resp.Status)
		}
		break // Request was successful
	}

	return nil
}
