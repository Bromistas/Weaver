package common

import (
	"encoding/json"
	"log"
	"net"
	"sort"
)

func insertIntoSorted(slice []int, item int) []int {
	i := sort.Search(len(slice), func(i int) bool { return slice[i] >= item })
	slice = append(slice, 0)
	copy(slice[i+1:], slice[i:])
	slice[i] = item
	return slice
}

func removeFromSorted(slice []int, item int) []int {
	i := sort.Search(len(slice), func(i int) bool { return slice[i] > item })
	if i < len(slice) && slice[i] == item {
		slice = append(slice[:i], slice[i+1:]...)
	}
	return slice
}

func searchInSorted(slice []int, item int) int {
	return sort.Search(len(slice), func(i int) bool { return slice[i] > item })
}

func CompareIPs(ip1, ip2 net.IP) int {
	for i := 0; i < len(ip1) && i < len(ip2); i++ {
		if ip1[i] < ip2[i] {
			return -1
		}
		if ip1[i] > ip2[i] {
			return 1
		}
	}
	return 0
}

type queueMessageFormat struct {
	Message string `json:"message"`
}

func EncodeURLMessage(urlMessage URLMessage) ([]byte, error) {
	// Step 1: Marshal the URLMessage into a JSON string
	urlMessageJSON, err := json.Marshal(urlMessage)
	if err != nil {
		log.Printf("Error marshalling URLMessage: %v", err)
		return nil, err
	}

	// Step 2: Construct the new struct with the marshalled URLMessage as the Message property
	queueMessage := queueMessageFormat{
		Message: string(urlMessageJSON),
	}

	// Step 3: Marshal the new struct into a JSON string
	queueMessageJSON, err := json.Marshal(queueMessage)
	if err != nil {
		log.Printf("Error marshalling queue message: %v", err)
		return nil, err
	}

	return queueMessageJSON, nil
}

// DecodeMessage decodes a JSON string into a URLMessage
func DecodeMessage(jsonStr string) (URLMessage, error) {
	var qm queueMessageFormat
	err := json.Unmarshal([]byte(jsonStr), &qm)
	if err != nil {
		log.Printf("Error unmarshalling queue message: %v", err)
		return URLMessage{}, err
	}

	var urlMessage URLMessage
	err = json.Unmarshal([]byte(qm.Message), &urlMessage)
	if err != nil {
		log.Printf("Error unmarshalling URLMessage: %v", err)
		return URLMessage{}, err
	}

	return urlMessage, nil
}
