package main

import (
	common "commons"
	"context"
	"encoding/json"
	"github.com/gocolly/colly/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"log"
	pb "protos"
	"strconv"
	"strings"
	"sync"
	"time"
)

const address = "localhost:50051"

func put_pair(addr, k, v string, group *sync.WaitGroup) {
	defer group.Done()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("connection to %v failed: %v", addr, err)
	}
	defer conn.Close()
	c := pb.NewDHTClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	_, err = c.Put(ctx, &pb.Pair{Key: k, Value: v})
	if err != nil {
		log.Fatalf("could not put to %v: %v", addr, err)
	}

}

func NeweggProductHandler(url string) {
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
		product.Price = float32(price)
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

	// Put to the database
	put_pair(address, product.Name, string(body), &sync.WaitGroup{})
}

func NeweggRootHandler(searchURL string, ch *amqp.Channel, q amqp.Queue) {
	c := colly.NewCollector()

	c.OnHTML(".item-cells-wrap", func(e *colly.HTMLElement) {
		// Log the search results
		log.Printf("Found something\n")
		log.Printf("Found {%s} products\n", e.Text)

	})

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
		urlMessage := URLMessage{
			URL:     productURL,
			URLType: NeweggProduct,
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
