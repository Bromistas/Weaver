package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
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
	s.GetMapHandler(w, r)
}

func (s *Server) GetMapHandler(w http.ResponseWriter, r *http.Request) {

	mapData := s.Node.GetMap()

	mapData[s.Node.ID] = nil

	data, err := json.Marshal(mapData)
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

	for _, existingNode := range s.Node.IDMap {

		if existingNode.ID == s.Node.ID || existingNode.ID == newNode.ID {
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
		resp, err := http.Post("http://"+existingNode.Address+"/insert", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error sending POST request")
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

func (s Server) Join(addr net.Addr) error {
	// Create a copy of the node
	nodeCopy := *s.Node

	// Set the map entry for the node itself to nil in the copy
	nodeCopy.IDMap[nodeCopy.ID] = nil

	// Marshal the copy to JSON
	jsonData, err := json.Marshal(nodeCopy)
	if err != nil {
		fmt.Println("Error marshalling JSON data")
		return err
	}

	// Send the marshalled JSON data
	resp, err := http.Post("http://"+addr.String()+"/join", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending POST request")
		return err
	}

	// The response from the join endpoint is a map of the network. We will unmarshal this response into a map of nodes.
	decoder := json.NewDecoder(resp.Body)
	var mapData map[int]Node
	err = decoder.Decode(&mapData)
	if err != nil {
		fmt.Println("Error decoding JSON data")
		return err
	}

	for _, node := range mapData {
		s.Node.Insert(&node)
	}

	// Close the response body
	defer resp.Body.Close()

	return nil
}
