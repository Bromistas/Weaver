package handlers

import (
	"encoding/json"
	"github.com/gocolly/colly/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"src/common"
	"strconv"
	"strings"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func NeweggProductHandler(url string, ch *amqp.Channel, exch string) {
	c := colly.NewCollector()
	product := common.Product{}

	c.OnHTML(".product-title", func(e *colly.HTMLElement) {
		product.Name = e.Text
	})

	c.OnHTML(".product-bullets", func(e *colly.HTMLElement) {
		product.Description = e.Text
	})

	c.OnHTML(".product-pane .product-price .price .price-current strong", func(e *colly.HTMLElement) {
		cleanedText := strings.Replace(e.Text, ",", "", -1)
		price, err := strconv.ParseFloat(cleanedText, 64)
		if err != nil {
			log.Printf("Failed to parse price: %v", err)
			return
		}
		product.Price = price
	})

	c.OnHTML(".product-wrap .product-reviews .product-rating i", func(e *colly.HTMLElement) {
		product.Rating = strings.TrimSpace(e.Attr("title"))
	})

	err := c.Visit(url)
	if err != nil {
		log.Fatal(err)
	}

	body, err := json.Marshal(product)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.Publish(
		exch,  // exchange
		"",    // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")
}

func NeweggRootHandler(searchURL string, ch *amqp.Channel, q amqp.Queue) {
	c := colly.NewCollector()

	// This selector will need to be adjusted to match the actual structure of the Newegg search results page
	c.OnHTML(".item-cell", func(e *colly.HTMLElement) {
		// Extract product details
		productURL := e.ChildAttr("a.item-title", "href")
		reviewCount := e.ChildText(".item-rating-num")

		log.Printf("Scanning {%s}")

		// Check if the product has more than 10 reviews
		str := strings.Trim(reviewCount, "()")
		reviews, err := strconv.Atoi(str)

		if err != nil || reviews <= 10 {
			log.Printf("Rejected")
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
