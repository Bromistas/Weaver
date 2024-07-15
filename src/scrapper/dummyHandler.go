package main

import (
	common "commons"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
)

func (s *ScrapperNode) DummyHandler(url string) {
	fmt.Println("Dummy Handler")
	product := common.Product{}

	product.Name = strconv.Itoa(rand.Int())
	product.Price = 12.31
	product.Rating = "4.5"
	product.Description = "This is a description"
	product.URL = url

	body, err := json.Marshal(product)
	if err != nil {
		log.Fatal(err)
	}

	addr := s.StorageAddress + ":" + strconv.Itoa(s.StoragePort)

	// Put to the database
	group := &sync.WaitGroup{}
	group.Add(1)

	log.Printf("Inserting key with name %s and url %s", product.Name, product.URL)
	put_pair(addr, product.Name, string(body), group)
}
