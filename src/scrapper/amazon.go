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

func (s ScrapperNode) AmazonProductHandler(url string) {
	fmt.Println("Amazon Product Handler")

	// Scrap and create a product with name, price and rating
	// Send the product to the storage service

	//c := colly.NewCollector()
	product := common.Product{}
	//
	//c.OnHTML("#productTitle", func(e *colly.HTMLElement) {
	//	if strings.TrimSpace(e.Text) != "" {
	//		product.Name = strings.TrimSpace(e.Text)
	//	}
	//})

	product.Name = strconv.Itoa(rand.Int())
	product.Price = 12.31
	product.Rating = "4.5"
	product.Description = "This is a description"
	product.URL = url

	//err := c.Visit(url)
	//if err != nil {
	//	log.Fatal(err)
	//}

	body, err := json.Marshal(product)
	if err != nil {
		log.Fatal(err)
	}

	addr := s.StorageAddress + ":" + strconv.Itoa(s.StoragePort)

	// Put to the database
	group := &sync.WaitGroup{}
	group.Add(1)
	put_pair(addr, product.Name, string(body), group)
}
