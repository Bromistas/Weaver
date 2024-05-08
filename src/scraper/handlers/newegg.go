package handlers

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"strings"
)

func NeweggProductHandler(url string) {
	c := colly.NewCollector()

	c.OnHTML(".product-title", func(e *colly.HTMLElement) {
		fmt.Println("Name: ", e.Text)
	})

	c.OnHTML(".product-bullets", func(e *colly.HTMLElement) {
		fmt.Println("Description: ", e.Text)
	})

	c.OnHTML(".product-pane .product-price .price .price-current strong", func(e *colly.HTMLElement) {
		fmt.Println("Price: ", e.Text)
	})

	c.OnHTML(".product-wrap .product-reviews .product-rating i", func(e *colly.HTMLElement) {
		fmt.Println("Rating: ", strings.TrimSpace(e.Attr("title")))
	})

	err := c.Visit(url)
	if err != nil {
		log.Fatal(err)
	}
}
