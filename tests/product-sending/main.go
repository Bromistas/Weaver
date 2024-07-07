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

type URLMessage struct {
	URL     string
	URLType URLType
}

type URLType int

const (
	AmazonProduct URLType = iota
	NeweggProduct
	NeweggRoot
)

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

	urlMessage := URLMessage{
		URL:     "https://www.amazon.com/iPhone-Charger-Anker-AirPods-Included/dp/B0C8HHV9DK/?content-id=amzn1.sym.64be5821-f651-4b0b-8dd3-4f9b884f10e5",
		URLType: AmazonProduct,
	}

	body, err := json.Marshal(urlMessage)
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
