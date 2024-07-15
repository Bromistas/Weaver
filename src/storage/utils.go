package main

import (
	"bytes"
	common "commons"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func CheckIps() (net.IP, error) {
	// Get the default network interface
	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Get the IP addresses associated with this interface
	addrs, err := iface.Addrs()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var ip net.IP
	for _, addr := range addrs {

		if strings.Contains(addr.String(), "::") {
			continue
		}

		switch v := addr.(type) {
		case *net.IPNet:
			//if strings.Contains(v.String(), "::"){
			ip = v.IP
			//}
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil {
			fmt.Printf("%s has IP %s\n", iface.Name, ip.String())
		}
	}

	return ip, nil
}

func TakeBytesAndTake1(b []byte) []byte {
	hexa := hex.EncodeToString(b)
	parseInt, err := strconv.ParseInt(hexa, 16, 64)

	if err != nil {
		log.Fatalf("Failed to parse int: %v", err)
	}

	parseInt = parseInt - 1
	hexa = strconv.FormatInt(parseInt, 16)

	return []byte(hexa)
}

// SendProductRequest marshals the product and address, then sends them to the /replicate endpoint
func SendProductRequest(product common.Product, address string) error {
	// Create the request payload
	endpoint := "http://" + address + "/replicate"

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
