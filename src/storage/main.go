package main

import (
	"chord"
	"commons"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	ring *chord.Ring
	addr string
)

func ServeChordWrapper(conf *chord.Config, trans chord.Transport, address string, bootstrap string) {
	log.Printf("[*] Node %s started", address)
	// go ReplicateData(context.Background(), n, 5*time.Second)

	var err error

	if bootstrap == "" {
		ring, err = chord.Create(conf, trans)
		if err != nil {
			log.Fatalf("Failed to create ring: %v", err)
		} else {
			log.Printf("Succesfully created the ring")
		}
	} else {
		for {
			ring, err = chord.Join(conf, trans, bootstrap)
			i := 0
			if err != nil {
				log.Printf("Failed to join ring attempt %d: %v", i, err)
				time.Sleep(1 * time.Second)
				i++
			} else {
				log.Printf("Successfully joined the network of %s", bootstrap)
				break
			}
		}
	}
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
	//node1 := node.NewChordNode(address, CustomPut)
	config := chord.DefaultConfig(address)
	server := setupServer(address)
	transport, err := chord.InitTCPTransport(address, 4*time.Second, server)

	if err != nil {
		log.Fatalf("Failed to create transport: %v", err)
	}

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
		log.Printf("Found storage node, joining the ring %s", found_ip)
		//node2 := node.NewChordNode(found_ip+":"+fmt.Sprint(found_port), CustomPut)
		go ServeChordWrapper(config, transport, address, found_ip+":"+fmt.Sprint(found_port))
	} else {
		log.Println("No storage node found, starting a new ring")
		go ServeChordWrapper(config, transport, address, "")
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

func gatherHandler(w http.ResponseWriter, r *http.Request) {
	//// Only allow GET method
	//if r.Method != http.MethodGet {
	//	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	//	return
	//}

	// Read directory
	files, err := ioutil.ReadDir(addr)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var products []common.Product

	// Iterate over files
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			// Open and decode JSON
			fp := filepath.Join(addr, file.Name())
			data, err := os.ReadFile(fp)
			if err != nil {
				continue // Skip files that can't be read
			}

			var product common.Product
			if err := json.Unmarshal(data, &product); err != nil {
				continue // Skip files that can't be decoded
			}

			products = append(products, product)
		}
	}

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, "Failed to encode products", http.StatusInternalServerError)
	}
}

func insertHandler(w http.ResponseWriter, r *http.Request) {
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

	insertInStore(ring, payload, addr)

	// Respond to the client
	w.WriteHeader(http.StatusTemporaryRedirect)
	//fmt.Fprintf(w, "File replicated successfully: %s\n", filename)
	//log.Printf("File replicated successfully: %s\n", filename)
}

func setupServer(address string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/insert", insertHandler)
	mux.HandleFunc("/healthcheck", healthCheckHandler)
	mux.HandleFunc("/replicate", replicateHandler)
	mux.HandleFunc("/gather", gatherHandler)

	httpServer := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	return httpServer
}

func main() {
	group := &sync.WaitGroup{}

	group.Add(1)
	go mainWrapper(group)

	group.Wait()
}
