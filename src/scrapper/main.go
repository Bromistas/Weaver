package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	baseUrlMessage := URLMessage{
		URL:     baseSearchURL,
		URLType: NeweggRoot,
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
		"scrap", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

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

	//exch := "products"

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var urlMessage URLMessage

			// Log message
			log.Printf("Received a message: %s", d.Body)

			err := json.Unmarshal(d.Body, &urlMessage)
			if err != nil {
				log.Fatal(err)
			}

			switch urlMessage.URLType {
			case AmazonProduct:
				AmazonProductHandler(urlMessage.URL)
			case NeweggProduct:
				NeweggProductHandler(urlMessage.URL)
			case NeweggRoot:
				NeweggRootHandler(urlMessage.URL, ch, q)
			default:
				log.Printf("Unknown URL type: %v", urlMessage.URLType)
			}
		}
	}()

	// Server
	http.HandleFunc("/query/", func(w http.ResponseWriter, r *http.Request) {
		param := strings.TrimPrefix(r.URL.Path, "/query/")
		newQueryMessageSend(param, ch, q)
		fmt.Fprintf(w, "Received param: %s. Processing request", param)
	})

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
