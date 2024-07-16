package main

import (
	"chord"
	common "commons"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

func serveChordWrapper(conf *chord.Config, trans chord.Transport, address string, bootstrap string) {
	log.Printf("[*] Node %s started", address)

	go requeueInvisibleMessages()
	//go q.leaderAttention()

	var err error

	if bootstrap == "" {
		_, err = chord.Create(conf, trans)
		if err != nil {
			log.Fatalf("Failed to create ring: %v", err)
		} else {
			log.Printf("Succesfully created the ring")
		}
	} else {
		_, err = chord.Join(conf, trans, bootstrap)
		if err != nil {
			log.Fatalf("Failed to join ring: %v", err)
		} else {
			log.Printf("Successfully joined the network of %s", bootstrap)
		}
	}

	//node.ServeChord(context.Background(), n, bootstrap, group, nil, server)
}

func setupServer(address string) *http.Server {
	mux := http.NewServeMux()
	// Register handlers
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/pop", popHandler)
	mux.HandleFunc("/ack", ackHandler)
	mux.HandleFunc("/healthcheck", healthCheckHandler)

	// Create a http.Server
	server := &http.Server{
		Addr:    address, // or any other port
		Handler: mux,
	}

	return server
}

func mainWrapper(group *sync.WaitGroup) {
	defer group.Done()

	address, err := common.GetHostIPV1()
	if err != nil {
		log.Fatalf("Failed to get host IP: %v", err)
	}

	q.addr = address

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	role := os.Getenv("ROLE")
	address += ":" + strconv.Itoa(port)

	config := chord.DefaultConfig(address)
	transport, err := chord.InitTCPTransport(address, 6*time.Second)
	if err != nil {
		log.Fatalf("Failed to create transport: %v", err)
	}

	server := setupServer(q.addr + ":" + strconv.Itoa(port+1))
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Create a directory this the address name if it doesnt exist already
	err = os.Mkdir(address, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory: %v", err)
	}

	foundIp := ""
	foundPort := 0

	discovered, err := common.NetDiscover(strconv.Itoa(port), role, false, false)

	if len(discovered) > 0 {
		foundIp = discovered[0]
		foundPort = port
	}

	if foundIp != "" {
		fmt.Println("Found queue node, joining the ring")
		//node2 := node.NewChordNode(foundIp+":"+fmt.Sprint(foundPort), nil)
		go serveChordWrapper(config, transport, address, foundIp+":"+fmt.Sprint(foundPort))
	} else {
		fmt.Println("No queue node found, starting a new ring")
		go serveChordWrapper(config, transport, address, "")
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
