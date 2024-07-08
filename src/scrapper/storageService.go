package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type StorageService interface {
	HealthCheck() (string, error)
}

type StorageServiceClient struct {
	baseURL string
}

func NewStorageServiceClient(baseURL string) *StorageServiceClient {
	return &StorageServiceClient{baseURL: baseURL}
}

func (s *StorageServiceClient) HealthCheck() (string, error) {
	url := fmt.Sprintf("http://%s/healthcheck", s.baseURL)
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
