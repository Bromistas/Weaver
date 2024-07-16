package main

import (
	common "commons"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"math"
	"strconv"
	"time"
)

func (s *ScrapperNode) retryVisit(c *colly.Collector, url string, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = c.Visit(url)
		if err == nil {
			return nil // Success, no need to retry
		}
		// Exponential backoff calculation
		backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
		time.Sleep(backoff)
	}
	return err // Return the last error encountered
}

func (s *ScrapperNode) AmazonRootHandler(url string) {
	log.Printf("Amazon Root Handler for URL: %s", url)
	c := colly.NewCollector()
	var productURLs []string

	// Limit the number of products to scrape.
	const maxProducts = 3

	c.OnHTML("div.s-result-item", func(e *colly.HTMLElement) {
		if len(productURLs) >= maxProducts {
			return // Stop if we already have 3 products.
		}
		// Extract the product URL and append it to the list.
		productURL := e.ChildAttr("a.a-link-normal", "href")
		productURLs = append(productURLs, productURL)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start the scraping process.
	err := s.retryVisit(c, url, 5) // Retry up to 5 times
	if err != nil {
		log.Printf("Failed to scrape Amazon url %s: %v", url, err)
	}

	// Put the urls back in the queue with the type AmazonProduct
	queueClient := NewQueueServiceClient(s.QueueAddress + ":" + strconv.Itoa(s.QueuePort))

	for _, productURL := range productURLs {
		// Create a URL message for the product URL with the appropriate type
		urlMessage := common.URLMessage{URL: productURL, URLType: common.AmazonProduct}
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
