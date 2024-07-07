package main

import (
	common "commons"
	"context"
	"encoding/json"
	"fmt"
	"github.com/grandcat/zeroconf"
	"log"
	"net"
	"net/http"
	"node"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Queue struct {
	messages []string
	mu       sync.Mutex
}

func (q *Queue) Put(message string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = append(q.messages, message)
}

func (q *Queue) Pop() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.messages) == 0 {
		return "", false
	}
	msg := q.messages[0]
	q.messages = q.messages[1:]
	return msg, true
}

func serveChordWrapper(n *node.ChordNode, bootstrap *node.ChordNode, group *sync.WaitGroup) {
	log.Printf("[*] Node %s started", n.Address)
	node.ServeChord(context.Background(), n, bootstrap, group, nil)
}

func setupServer(queue *Queue) {
	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		queue.Put(body.Message)
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/pop", func(w http.ResponseWriter, r *http.Request) {
		msg, ok := queue.Pop()
		if !ok {
			http.Error(w, "Queue is empty", http.StatusNoContent)
			return
		}
		response := struct {
			Message string `json:"message"`
		}{
			Message: msg,
		}
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Status string `json:"status"`
			Time   string `json:"time"`
		}{
			Status: "healthy",
			Time:   time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	})
}

func setupDiscovery(n *node.ChordNode, address string, waitTime time.Duration, group *sync.WaitGroup) {

	addr := strings.Split(address, ":")
	port, _ := strconv.Atoi(addr[1])
	ip := net.ParseIP(addr[0])

	foundIp := ""
	foundPort := 0
	discoveryCallback := func(entry *zeroconf.ServiceEntry) {
		if strings.HasPrefix(entry.ServiceInstanceName(), "Queue") {

			if len(entry.AddrIPv4) == 0 {
				log.Printf("Found service: %s, but no IP address. Going localhost", entry.ServiceInstanceName())
				foundIp = "localhost"
				foundPort = entry.Port
			} else {
				temp := entry.AddrIPv4[0].String()

				if !strings.Contains(foundIp, "::") {
					foundIp = temp
					foundPort = entry.Port
				}
			}
			log.Printf("Registered service: %s, IP: %s, Port: %d\n", entry.ServiceInstanceName(), entry.AddrIPv4, entry.Port)
		}
	}

	common.Discover("_http._tcp", "local.", waitTime, discoveryCallback)

	if foundIp != "" {
		log.Println("Found queue node, joining the ring")
		other := node.NewChordNode(foundIp+":"+fmt.Sprint(foundPort), nil)
		go serveChordWrapper(n, other, group)
	} else {
		log.Println("No queue node found, starting a new ring")
		go serveChordWrapper(n, nil, group)
	}

	//common.ThreadBroadListen(strconv.Itoa(port))
	serviceName := "QueueNode"
	serviceType := "_http._tcp"
	domain := "local."

	err := common.RegisterForDiscovery(serviceName, serviceType, domain, port, ip)
	if err != nil {
		log.Fatalln(err)
	}
}

func setupNode(address string, port int, group *sync.WaitGroup, waitTime time.Duration, first bool) {
	defer group.Done()

	queue := &Queue{
		messages: make([]string, 0),
	}
	n := node.NewChordNode(address, nil)

	if first {
		setupServer(queue)
	}

	go setupDiscovery(n, address, waitTime, group)

	time.Sleep(7 * time.Second)

	// Print predecessor of n
	pred, _ := n.FindPredecessor(context.Background(), n)
	fmt.Printf("Predecessor of %s is %s\n", n.Address, pred.Address)

	// Listen and serve
	// TODO: Change to use instead of localhost
	log.Printf("[*] Listening to messages in port %d", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func main() {
	group := &sync.WaitGroup{}
	group.Add(2)

	go setupNode("127.0.0.1:9000", 9000, group, 1*time.Second, true)
	//go setupNode("127.0.0.1:9001", 9001, group, 3*time.Second, false)
	//go setupNode("127.0.0.1:9002", 9002, group, 5*time.Second, false)

	// Wait indefinitely
	group.Wait()
}
