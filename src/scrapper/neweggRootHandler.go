package main

import (
	common "commons"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"strconv"
)

func (s *ScrapperNode) NeweggRootHandler(url string) {
	fmt.Println("Newegg Root Handler")

	c := colly.NewCollector(colly.AllowURLRevisit())
	var productURLs []string

	// Find and visit all links
	c.OnHTML(".item-container", func(e *colly.HTMLElement) {
		if len(productURLs) < 6 {
			productURL := e.ChildAttr(".item-title", "href")
			productURLs = append(productURLs, productURL)
		}
	})

	// Start scraping the page
	err := c.Visit(url)
	if err != nil {
		log.Fatal(err)
	}

	queueClient := NewQueueServiceClient(s.QueueAddress + ":" + strconv.Itoa(s.QueuePort))

	// Print the top 3 products
	for _, productURL := range productURLs {
		// Create a URL message for the product URL with the appropriate type
		urlMessage := common.URLMessage{URL: productURL, URLType: common.NeweggProduct}
		marshaledMessage, err := json.Marshal(urlMessage)
		if err != nil {
			log.Printf("Failed to marshal %s URL message: %v", productURL, err)
			continue
		}

		err = queueClient.Put(string(marshaledMessage))
		if err != nil {
			log.Printf("Failed to put %s into the queue: %v", productURL, err)
		}
	}
}
