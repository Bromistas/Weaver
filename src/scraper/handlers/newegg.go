package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"scraper/common"
	"strconv"
	"strings"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

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

func NeweggRootHandler(searchURL string, ch *amqp.Channel, q amqp.Queue) {
	c := colly.NewCollector()

	// This selector will need to be adjusted to match the actual structure of the Newegg search results page
	c.OnHTML(".item-cell", func(e *colly.HTMLElement) {
		// Extract product details
		productURL := e.ChildAttr("a.item-title", "href")
		reviewCount := e.ChildText(".item-rating-num")

		// Check if the product has more than 10 reviews
		str := strings.Trim(reviewCount, "()")
		reviews, err := strconv.Atoi(str)

		if err != nil || reviews <= 10 {
			return
		}

		// Add the product URL to the RabbitMQ queue
		urlMessage := common.URLMessage{
			URL:     productURL,
			URLType: common.NeweggProduct,
		}
		body, err := json.Marshal(urlMessage)
		if err != nil {
			log.Fatal(err)
		}

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s\n", urlMessage.URL)
	})

	c.Visit(searchURL)
}
