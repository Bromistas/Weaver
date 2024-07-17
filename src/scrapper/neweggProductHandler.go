package main

import (
	common "commons"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"strconv"
	"strings"
)

func (s *ScrapperNode) NeweggProductHandler(url string) {
	log.Println("Newegg Product Handler")

	product := common.Product{URL: url}

	// Create a new collector
	c := colly.NewCollector()

	// Scrape the product name
	c.OnHTML(".product-title", func(e *colly.HTMLElement) {
		product.Name = e.Text
	})

	// Scrape the product price
	c.OnHTML(".product-buy-box .price-current", func(e *colly.HTMLElement) {

		if product.Price != 0 {
			return
		}

		// Get the child strong element and its text
		strong := e.DOM.Find("strong")
		text := strong.Text()

		temp, _ := strconv.ParseFloat(strings.TrimSpace(text), 32)
		product.Price = float32(temp)
	})

	// Scrape the product description
	c.OnHTML("#product-details .product-bullets", func(e *colly.HTMLElement) {
		product.Description = e.Text
	})

	// Scrape the product rating
	c.OnHTML(".product-rating", func(e *colly.HTMLElement) {
		product.Rating = e.ChildAttr("i", "title")
	})

	// Start scraping the page
	err := c.Visit(url)
	if err != nil {
		log.Printf("Failed to scrape Newegg product URL %s: %v", url, err)
	}

	// Put the product in the database
	addr := s.StorageAddress + ":" + strconv.Itoa(s.StoragePort)
	// Marshal and stringify product

	err = insertProduct(product, addr)
	if err != nil {
		fmt.Printf("Error while inserting product: %s", err)
	}
}
