package main

import (
	common "commons"
	"fmt"
	"log"
	"math/rand"
	"strconv"
)

func (s *ScrapperNode) DummyHandler(url string) {
	fmt.Println("Dummy Handler")
	product := common.Product{}

	product.Name = strconv.Itoa(rand.Int())
	product.Price = 12.31
	product.Rating = "4.5"
	product.Description = "This is a description"
	product.URL = url

	log.Printf("Inserting key with name %s and url %s", product.Name, product.URL)

	//body, err := json.Marshal(product)
	//if err != nil {
	//	log.Fatal(err)
	//}

	addr := s.StorageAddress + ":" + strconv.Itoa(s.StoragePort)
	err := insertProduct(product, addr)
	if err != nil {
		fmt.Printf("Error while inserting product %s: %s", product.Name, err.Error())
		return
	}

	// Put to the database
	//group := &sync.WaitGroup{}
	//group.Add(1)

	//put_pair(addr, product.Name, string(body), group)
}
