package main

import (
	"commons"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"node"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	addr string
)

func ServeChordWrapper(n *node.ChordNode, bootstrap *node.ChordNode, group *sync.WaitGroup) {
	log.Printf("[*] Node %s started", n.Address)
	go ReplicateData(context.Background(), n, 5*time.Second)

	// Create your HTTP server
	httpServer := &http.Server{
		Handler: http.HandlerFunc(replicateHandler),
	}

	node.ServeChord(context.Background(), n, bootstrap, group, nil, httpServer)
}

func mainWrapper(group *sync.WaitGroup) {
	defer group.Done()

	address, err := common.GetHostIPV1()
	if err != nil {
		log.Fatalf("Failed to get host IP: %v", err)
	}

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	role := os.Getenv("ROLE")
	address += ":" + strconv.Itoa(port)

	addr = address
	node1 := node.NewChordNode(address, CustomPut)

	// Create a directory this the address name if it doesnt exist already
	err = os.Mkdir(address, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory: %v", err)
	}

	found_ip := ""
	found_port := 0

	discovered, err := common.NetDiscover(strconv.Itoa(port), role, false, false)

	if err != nil {
		log.Fatalf("Failed to discover: %v", err)
	}

	if len(discovered) > 0 {
		found_ip = discovered[0]
		found_port = port
	}

	if found_ip != "" {
		fmt.Println("Found storage node, joining the ring")
		node2 := node.NewChordNode(found_ip+":"+fmt.Sprint(found_port), CustomPut)
		go ServeChordWrapper(node1, node2, group)
	} else {
		fmt.Println("No storage node found, starting a new ring")
		go ServeChordWrapper(node1, nil, group)
	}

	common.ThreadBroadListen(strconv.Itoa(port), role)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	}{
		Status: "healthy",
		Time:   time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(response)
}

func replicateHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read and parse the JSON payload
	var payload common.Product
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Generate a filename based on the current timestamp
	filename := fmt.Sprintf("%s.json", payload.Name)
	fp := filepath.Join(addr, filename)

	// Convert the payload to JSON
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write the JSON content to the file
	err = os.WriteFile(fp, data, 0644)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Respond to the client
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "File replicated successfully: %s\n", filename)
	log.Printf("File replicated successfully: %s\n", filename)
}

func setupServer() {
	http.HandleFunc("/healthcheck", healthCheckHandler)
	http.HandleFunc("/replicate", replicateHandler)
}

func main() {
	group := &sync.WaitGroup{}

	setupServer()

	group.Add(1)
	go mainWrapper(group)

	group.Wait()
}
