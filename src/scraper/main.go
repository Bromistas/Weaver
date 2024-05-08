package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"strings"
)

func main() {
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

	err := c.Visit("https://www.newegg.com/ponder-blue-asus-um3504da-nb74/p/N82E16834236473?Item=N82E16834236473&cm_sp=homepage_carousel_p3_newegghandpicked-_-item-_-34-236-473")
	if err != nil {
		log.Fatal(err)
	}
}
