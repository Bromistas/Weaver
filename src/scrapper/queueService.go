package main

import (
	"bytes"
	common "commons"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type QueueService interface {
	Put(message string) error
	Pop() (common.Message, error)
	Ack(id string) error
	HealthCheck() (string, error)
	Listen(tickDuration time.Duration, node *ScrapperNode) error
}

type QueueServiceClient struct {
	baseURL string
}

func NewQueueServiceClient(baseURL string) *QueueServiceClient {
	return &QueueServiceClient{baseURL: baseURL}
}

func (q *QueueServiceClient) Put(message string) error {
	temp := strings.Split(q.baseURL, ":")[0]
	url := fmt.Sprintf("http://%s:9001/put", temp)
	body := map[string]string{"message": message}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("error while marshaling message: %s", message)
		return err
	}

	// Retry mechanism
	const maxRetries = 5
	const initialBackoff = time.Second

	var resp *http.Response
	var attempt int

	for attempt = 0; attempt < maxRetries; attempt++ {
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
		if err == nil && resp.StatusCode == http.StatusNoContent {
			// Success
			defer resp.Body.Close()
			return nil
		}

		// If an error occurred or unexpected status code, wait and retry
		time.Sleep(initialBackoff * time.Duration(1<<attempt))
	}

	if err != nil {
		return err
	}
	if resp != nil {
		defer resp.Body.Close()
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return fmt.Errorf("failed to send request after %d attempts", maxRetries)
}

func (q *QueueServiceClient) Pop() (common.Message, error) {
	url := fmt.Sprintf("http://%s/pop", q.baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return common.Message{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return common.Message{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return common.Message{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result common.Message
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return common.Message{}, err
	}

	// Send ack
	if err := q.Ack(result.ID); err != nil {
		return common.Message{}, err
	}

	return result, nil
}

func (q *QueueServiceClient) Ack(id string) error {
	url := fmt.Sprintf("http://%s/ack", q.baseURL)
	body := map[string]string{"id": id}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (q *QueueServiceClient) HealthCheck() (string, error) {
	url := fmt.Sprintf("http://%s/healthcheck", q.baseURL)
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

func (q *QueueServiceClient) Listen(tickDuration time.Duration, node *ScrapperNode) error {
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	log.Printf("Listening to queue %s", q.baseURL)
	for {
		select {
		case <-ticker.C:
			message, err := q.Pop()
			if err != nil {
				return err
			}

			if message.Body != "" {
				messageHandler(node, message.Body)
			}
		}
	}
}

func messageHandler(node *ScrapperNode, message string) {
	log.Printf("[?] Received message: %s\n", message)

	// Step 1: Unmarshal the message
	var urlMessage common.URLMessage
	err := json.Unmarshal([]byte(message), &urlMessage)
	if err != nil {
		log.Printf("Error unmarshalling URLMessage: %v", err)
		return
	}

	// Step 2: Switch statement to redirect based on the URL type
	switch urlMessage.URLType {
	case common.AmazonRoot:
		node.AmazonRootHandler(urlMessage.URL)
	case common.NeweggRoot:
		node.NeweggRootHandler(urlMessage.URL)
	case common.AmazonProduct:
		node.AmazonProductHandler(urlMessage.URL)
	case common.NeweggProduct:
		node.NeweggProductHandler(urlMessage.URL)
	case common.Dummy:
		node.DummyHandler(urlMessage.URL)

	// Add more cases as needed
	default:
		log.Printf("Unknown URL type: %v", urlMessage.URLType)
	}
}
