package main

import (
	common "commons"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"strconv"
)

func (s *ScrapperNode) AmazonProductHandler(url string) common.Product {
	log.Printf("Amazon Product Handler for URL: %s", url)
	c := colly.NewCollector(colly.AllowURLRevisit())
	product := common.Product{URL: url}

	c.OnHTML(".product-title-word-break", func(e *colly.HTMLElement) {
		product.Name = e.Text
	})

	c.OnHTML("selector-for-description", func(e *colly.HTMLElement) {
		product.Description = e.Text
	})

	c.OnHTML(".a-section .a-price span .a-price-whole", func(e *colly.HTMLElement) {
		t, err := strconv.ParseFloat(e.Text, 32)
		if err != nil {
			log.Println(err)
		}
		product.Price = float32(t)
	})

	c.OnHTML("#averageCustomerReviews .a-icon-star a-icon-alt", func(e *colly.HTMLElement) {
		product.Rating = e.Text
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start the scraping process.
	err := s.retryVisit(c, url, 5) // Retry up to 5 times
	if err != nil {
		log.Printf("Failed to scrape Amazon url %s: %v", url, err)
	}

	return product
}
