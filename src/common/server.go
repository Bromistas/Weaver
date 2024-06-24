package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var registered bool = false

type Server struct {
	Node   *Node
	Server *http.Server
}

func NewServer(node *Node, address string) *Server {
	s := &Server{
		Node: node,
		Server: &http.Server{
			Addr: address,
		},
	}

	if !registered {
		http.HandleFunc("/getMap", s.GetMapHandler)
		http.HandleFunc("/insert", s.InsertHandler)
		http.HandleFunc("/remove", s.RemoveHandler)

		http.HandleFunc("/join", s.JoinHandler)
		http.HandleFunc("/healthcheck", s.HealthCheckHandler)

		registered = true
	}

	return s
}

func (s *Server) Start() error {
	return s.Server.ListenAndServe()
}

func (s *Server) JoinHandler(w http.ResponseWriter, r *http.Request) {
	s.InsertHandler(w, r)

	mapData := s.Node.GetMap()
	answerList := make([]NodeFacet, 0)

	for key, value := range mapData {
		answerList = append(answerList, NodeFacet{ID: key, Address: value})
	}

	data, err := json.Marshal(answerList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) GetMapHandler(w http.ResponseWriter, r *http.Request) {

	mapData := s.Node.GetMap()
	answerList := make([]NodeFacet, 0)

	for key, value := range mapData {
		answerList = append(answerList, NodeFacet{ID: key, Address: value})
	}

	data, err := json.Marshal(answerList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) InsertHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var newNode Node
	err := decoder.Decode(&newNode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.Node.Insert(&newNode)

	for key, address := range s.Node.IDMap {

		if key == s.Node.ID || key == newNode.ID {
			continue
		}

		// Marshal newNode to JSON
		jsonData, err := json.Marshal(newNode)
		if err != nil {
			fmt.Println("Error marshalling JSON data")
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}

		// Send the marshalled JSON data
		resp, err := http.Post("http://"+address+"/insert", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error sending POST request from ", s.Node.ID, " ", err)
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			continue
		}
		// Close the response body
		defer resp.Body.Close()
	}
}

func (s *Server) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	node := NewNode(id, "")
	s.Node.Remove(node)
}

func (s *Server) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s Server) Join(newNode *Node) error {
	// Marshal the new node to JSON
	jsonData, err := json.Marshal(newNode)
	if err != nil {
		fmt.Println("Error marshalling JSON data")
		return err
	}

	// Send the marshalled JSON data
	resp, err := http.Post("http://"+newNode.Address+"/join", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending POST request from ", s.Node.ID, " ", err)
		return err
	}

	// The response from the join endpoint is a map of the network. We will unmarshal this response into a map of nodes.
	decoder := json.NewDecoder(resp.Body)
	var mapData []NodeFacet
	err = decoder.Decode(&mapData)
	if err != nil {
		fmt.Println("Error decoding JSON data")
		return err
	}

	for _, item := range mapData {

		if item.ID == s.Node.ID {
			continue
		}

		n := Node{
			ID:      item.ID,
			Address: item.Address,
		}

		s.Node.Insert(&n)
	}

	// Close the response body
	defer resp.Body.Close()

	return nil
}

func (s Server) Stabilize() {
	for {
		s.Node.Mutex.RLock()
		ids := s.Node.IDList
		s.Node.Mutex.RUnlock()

		for _, key := range ids {

			if key == s.Node.ID {
				continue
			}

			s.Node.Mutex.RLock()
			address := s.Node.IDMap[key]
			s.Node.Mutex.RUnlock()

			resp, err := http.Get("http://" + address + "/healthcheck")
			if err != nil {
				fmt.Println("Error sending GET request from ", s.Node.ID, " ", err, " to ", key, " with address ", address)
				s.Node.Remove(NewNode(key, ""))
				continue
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Println("Error in healthcheck")
				s.Node.Remove(NewNode(key, ""))
				continue
			}

			time.Sleep(50)

			defer resp.Body.Close()
		}

		time.Sleep(1 * time.Second)
	}
}
