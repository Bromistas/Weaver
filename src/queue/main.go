package main

import (
	"encoding/json"
	"log"
	"net/http"
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

func main() {
	queue := &Queue{
		messages: make([]string, 0),
	}

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

	log.Printf("[*] Listening to messages in port %d", 9000)
	http.ListenAndServe(":9000", nil)
}
