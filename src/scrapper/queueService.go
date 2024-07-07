package main

import (
	"bytes"
	common "commons"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	url := fmt.Sprintf("http://%s/put", q.baseURL)
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

	return result, nil
}

func (q *QueueServiceClient) Ack(id string) error {
	url := fmt.Sprintf("%s/ack", q.baseURL)
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
	node.AmazonProductHandler(message)
}