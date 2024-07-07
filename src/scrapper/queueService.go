package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type QueueService interface {
	Put(message string) error
	Pop() (string, error)
	HealthCheck() (string, error)
	Listen(tickRate time.Duration) error
}

type QueueServiceClient struct {
	baseURL string
}

func NewQueueServiceClient(baseURL string) *QueueServiceClient {
	return &QueueServiceClient{baseURL: baseURL}
}

func (q *QueueServiceClient) Put(message string) error {
	url := fmt.Sprintf("%s/put", q.baseURL)
	body := map[string]string{"message": message}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (q *QueueServiceClient) Pop() (string, error) {
	url := fmt.Sprintf("http://%s/pop", q.baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return "", nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Message, nil
}

func (q *QueueServiceClient) HealthCheck() (string, error) {
	url := fmt.Sprintf("%s/healthcheck", q.baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return fmt.Sprintf("Status: %s, Time: %s", result.Status, result.Time), nil
}

func (q *QueueServiceClient) Listen(tickRate time.Duration, node *ScrapperNode) error {
	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			message, err := q.Pop()
			if err != nil {
				return err
			}
			if message != "" {
				messageHandler(node, message)
			}
		}
	}
}

func messageHandler(node *ScrapperNode, message string) {
	log.Printf("[?] Received message: %s\n", message)
	node.AmazonProductHandler(message)
}
