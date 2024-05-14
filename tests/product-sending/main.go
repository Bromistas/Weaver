package main

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Product struct {
	Name        string
	Price       float64
	URL         string
	Description string
	Rating      string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"store-products", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	product := &Product{
		Name:        "Test Product 2",
		Price:       100.0,
		URL:         "http://example.com",
		Description: "This is a test product",
		Rating:      "5 stars",
	}

	body, err := json.Marshal(product)
	failOnError(err, "Failed to encode product into JSON")

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

	log.Printf(" [x] Sent %s", body)
}
