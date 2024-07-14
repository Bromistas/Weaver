package main

import (
	common "commons"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"node"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Queue struct {
	addr       string
	leaderAddr string
	messages   map[string]common.Message
	order      []string
	mu         sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		messages: make(map[string]common.Message),
		order:    make([]string, 0),
	}
}

func (q *Queue) Put(message string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	msg := common.Message{
		ID:      id,
		Body:    message,
		Visible: true,
	}
	q.messages[id] = msg
	q.order = append(q.order, id)

	log.Printf("Message %s added to queue with content %s", id, message)
}

func (q *Queue) Pop() (common.Message, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, id := range q.order {
		msg := q.messages[id]
		if msg.Visible {
			msg.Visible = false
			msg.Received = time.Now()
			q.messages[id] = msg
			return msg, true
		}
	}
	return common.Message{}, false
}

func (q *Queue) Ack(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.messages, id)
	for i, msgID := range q.order {
		if msgID == id {
			q.order = append(q.order[:i], q.order[i+1:]...)
			break
		}
	}
}

func (q *Queue) RequeueInvisibleMessages(timeout time.Duration) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	for id, msg := range q.messages {
		if !msg.Visible && now.Sub(msg.Received) > timeout {
			msg.Visible = true
			q.messages[id] = msg
		}
	}
}

var (
	q = NewQueue()
)

// Create a function that will constantly ping the leaderAddr for healthchecks and when failed it will change its leader to be the element in the ring with the lowest id
func (q *Queue) leaderAttention() {
	time.Sleep(5 * time.Second)

	for {
		if q.leaderAddr == "" {
			q.changeLeader(3 * time.Second)
		}

		// Create a healthcheck for q.leaderAddr
		if q.leaderAddr == q.addr {
			continue
		}

		url := fmt.Sprintf("http://%s/healthcheck", q.leaderAddr)
		resp, err := http.Get(url)

		if err != nil {
			q.changeLeader(3 * time.Second)
			fmt.Printf("error pinging leader: %s", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			q.changeLeader(3 * time.Second)
		}

		time.Sleep(1 * time.Second)
		fmt.Printf("Current leader is: %s", q.leaderAddr)
	}
}

func (q *Queue) changeLeader(waitTime time.Duration) {
	foundIp := ""
	foundPort := 0

	discoveryPort := "9001"
	discovered, err := common.NetDiscover(discoveryPort, "QUEUE", true)
	foundIp = discovered
	foundPort = 9000

	if err != nil {
		log.Printf("Error discovering a leader: %s", err)
	}

	// Convert string IPs to net.IP
	discoveredNetIP := net.ParseIP(discovered)
	currentNetIP := net.ParseIP(q.addr)

	if common.CompareIPs(currentNetIP, discoveredNetIP) <= 0 {
		q.leaderAddr = q.addr
	} else {
		finalLeader := fmt.Sprintf("%s:%s", foundIp, strconv.Itoa(foundPort))
		q.leaderAddr = finalLeader
	}

	log.Printf("Final leader for node %s is: %s", q.addr, q.leaderAddr)
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	q.Put(body.Message)
	w.WriteHeader(http.StatusNoContent)
}

func popHandler(w http.ResponseWriter, r *http.Request) {
	message, ok := q.Pop()
	if !ok {
		http.Error(w, "Queue is empty", http.StatusNoContent)
		return
	}
	json.NewEncoder(w).Encode(message)
}

func ackHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	q.Ack(body.ID)
	w.WriteHeader(http.StatusNoContent)
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

func requeueInvisibleMessages() {
	for {
		time.Sleep(10 * time.Second)
		q.RequeueInvisibleMessages(30 * time.Second)
	}
}

func serveChordWrapper(n *node.ChordNode, bootstrap *node.ChordNode, group *sync.WaitGroup) {
	log.Printf("[*] Node %s started", n.Address)

	mux := http.NewServeMux()
	// Register handlers
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/pop", popHandler)
	mux.HandleFunc("/ack", ackHandler)
	mux.HandleFunc("/healthcheck", healthCheckHandler)

	port := strings.Split(n.Address, ":")[1]

	// Create an http.Server
	server := &http.Server{
		Addr:    "0.0.0.0:" + port, // or any other port
		Handler: mux,
	}

	go requeueInvisibleMessages()
	go q.leaderAttention()

	node.ServeChord(context.Background(), n, bootstrap, group, nil, server)
}

func mainWrapper(group *sync.WaitGroup) {
	defer group.Done()

	address, err := common.GetHostIPV1()
	if err != nil {
		log.Fatalf("Failed to get host IP: %v", err)
	}

	q.addr = address
	q.leaderAddr = address

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	role := os.Getenv("ROLE")
	address += ":" + strconv.Itoa(port)

	node1 := node.NewChordNode(address, nil)

	// Create a directory this the address name if it doesnt exist already
	err = os.Mkdir(address, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory: %v", err)
	}

	foundIp := ""
	foundPort := 0

	discovered, err := common.NetDiscover(strconv.Itoa(port), role, false)
	foundIp = discovered
	foundPort = port

	if foundIp != "" {
		fmt.Println("Found queue node, joining the ring")
		node2 := node.NewChordNode(foundIp+":"+fmt.Sprint(foundPort), nil)
		go serveChordWrapper(node1, node2, group)
	} else {
		fmt.Println("No queue node found, starting a new ring")
		go serveChordWrapper(node1, nil, group)
	}

	common.ThreadBroadListen(strconv.Itoa(port), role)
}

func main() {

	group := &sync.WaitGroup{}

	// Setup Server and Node
	group.Add(1)
	go mainWrapper(group)

	// Wait for all goroutines to finish
	group.Wait()
}
