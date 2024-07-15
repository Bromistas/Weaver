package main

import (
	common "commons"
	"encoding/json"
	"github.com/gocolly/colly/v2"
	"log"
	"strconv"
	"strings"
	"sync"
)

func (s *ScrapperNode) NeweggProductHandler(url string) {
	product := common.Product{URL: url}

	// Create a new collector
	c := colly.NewCollector()

	// Scrape the product name
	c.OnHTML(".product-title", func(e *colly.HTMLElement) {
		product.Name = e.Text
	})

	// Scrape the product price
	c.OnHTML(".price-current", func(e *colly.HTMLElement) {
		temp, _ := strconv.ParseFloat(strings.TrimSpace(e.Text), 32)
		product.Price = float32(temp)
	})

	// Scrape the product description
	c.OnHTML("#product-details .product-bullets", func(e *colly.HTMLElement) {
		product.Description = e.Text
	})

	// Scrape the product rating
	c.OnHTML(".product-rating", func(e *colly.HTMLElement) {
		product.Rating = e.ChildAttr("i", "aria-label")
	})

	// Start scraping the page
	err := c.Visit(url)
	if err != nil {
		log.Printf("Failed to scrape Newegg product URL %s: %v", url, err)
	}

	// Put the product in the database
	addr := s.StorageAddress + ":" + strconv.Itoa(s.StoragePort)
	// Marshal and stringify product
	marshal, err := json.Marshal(product)
	if err != nil {
		log.Printf("Failed to marshal product: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	put_pair(addr, product.Name, string(marshal), wg)
	wg.Wait()
}
