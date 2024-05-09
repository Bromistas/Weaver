package main

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/url"
	"scraper/common"
	"scraper/handlers"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func generateSearchURL(query string) string {
	// URL encode the query
	encodedQuery := url.QueryEscape(query)
	// Generate the Newegg search URL
	return "https://www.newegg.com/p/pl?d=" + encodedQuery
}

func newQueryMessageSend(query string, ch *amqp.Channel, q amqp.Queue) {
	baseSearchURL := generateSearchURL(query)
	baseUrlMessage := common.URLMessage{
		URL:     baseSearchURL,
		URLType: common.NeweggRoot,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(baseUrlMessage)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", baseUrlMessage.URL)
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	app := fiber.New()

	app.Get("/query/:param", func(c *fiber.Ctx) error {
		query := c.Params("param")
		newQueryMessageSend(query, ch, q)
		return c.SendString("Received param: " + query + ". Processing request")
	})

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	exch := "products"

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var urlMessage common.URLMessage
			err := json.Unmarshal(d.Body, &urlMessage)
			if err != nil {
				log.Fatal(err)
			}

			switch urlMessage.URLType {
			case common.AmazonProduct:
				handlers.AmazonProductHandler(urlMessage.URL)
			case common.NeweggProduct:
				handlers.NeweggProductHandler(urlMessage.URL, ch, exch)
			case common.NeweggRoot:
				handlers.NeweggRootHandler(urlMessage.URL, ch, q)
			default:
				log.Printf("Unknown URL type: %v", urlMessage.URLType)
			}
		}
	}()

	go func() {
		log.Fatal(app.Listen(":4000"))
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
