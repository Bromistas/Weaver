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

	//ip := net.ParseIP("127.0.0.1")
	//ip, _ := CheckIps()
	//ip := net.ParseIP("127.0.0.1")

	address, err := common.GetHostIPV1()
	if err != nil {
		log.Fatalf("Failed to get host IP: %v", err)
	}

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	role := os.Getenv("ROLE")
	address += ":" + strconv.Itoa(port)

	log.Println("Port: ", port)
	//waitTime, _ := time.ParseDuration(os.Getenv("WAIT_TIME"))
	//address := ip.String() + ":" + os.Getenv("PORT")

	node1 := node.NewChordNode(address, CustomPut)

	// Create a directory this the address name if it doesnt exist already
	err = os.Mkdir(address, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory: %v", err)
	}

	found_ip := ""
	found_port := 0
	//discoveryCallback := func(entry *zeroconf.ServiceEntry) {
	//	if strings.HasPrefix(entry.ServiceInstanceName(), "Storage") {
	//
	//		if len(entry.AddrIPv4) == 0 {
	//			log.Printf("Found service: %s, but no IP address", entry.ServiceInstanceName(), ". Going localhost")
	//			found_ip = "localhost"
	//			found_port = entry.Port
	//		} else {
	//			temp := entry.AddrIPv4[0].String()
	//
	//			if !strings.Contains(found_ip, "::") {
	//				found_ip = temp
	//				found_port = entry.Port
	//			}
	//		}
	//		log.Printf("Registered service: %s, IP: %s, Port: %d\n", entry.ServiceInstanceName(), entry.AddrIPv4, entry.Port)
	//	}
	//}

	//common.Discover("_http._tcp", "local.", waitTime, discoveryCallback)
	discovered, err := common.NetDiscover(strconv.Itoa(port), role)
	found_ip = discovered
	found_port = port

	if found_ip != "" {

		//found_ip = strings.Split(chordAddr, ":")[0]
		//found_port, _ = strconv.Atoi(strings.Split(chordAddr, ":")[1])
		fmt.Println("Found storage node, joining the ring")

		node2 := node.NewChordNode(found_ip+":"+fmt.Sprint(found_port), CustomPut)
		go ServeChordWrapper(node1, node2, group)
	} else {
		fmt.Println("No storage node found, starting a new ring")
		go ServeChordWrapper(node1, nil, group)
	}

	common.ThreadBroadListen(strconv.Itoa(port), role)
	//serviceName := "StorageNode"
	//serviceType := "_http._tcp"
	//domain := "local."
	//
	//err = common.RegisterForDiscovery(serviceName, serviceType, domain, port, ip)
	//if err != nil {
	//	log.Fatalln(err)
	//}
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
	filepath := filepath.Join(addr, filename)

	// Convert the payload to JSON
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write the JSON content to the file
	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Respond to the client
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "File created successfully: %s\n", filename)
}

func setupServer() {
	http.HandleFunc("/healthcheck", healthCheckHandler)
	http.HandleFunc("/replicate", replicateHandler)
}

func main() {
	//if len(os.Args) != 4 {
	//	fmt.Println("Usage: program <address> <port> <waitTime>")
	//	os.Exit(1)
	//}

	//addr = os.Args[1]

	//address := os.Args[1]
	//portStr := os.Args[2]
	//waitTimeStr := os.Args[3]
	//
	//port, err := strconv.Atoi(portStr)
	//if err != nil {
	//	log.Fatalf("Invalid port: %v", err)
	//}
	//
	//waitTime, err := time.ParseDuration(waitTimeStr)
	//if err != nil {
	//	log.Fatalf("Invalid wait time: %v", err)
	//}

	group := &sync.WaitGroup{}

	setupServer()

	group.Add(1)
	go mainWrapper(group)

	group.Wait()
}
