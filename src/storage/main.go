package main

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	exch := "products"
	err = ch.ExchangeDeclare(
		exch,     // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		exch,   // exchange
		false,
		nil)
	failOnError(err, "Failed to bind a queue")

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

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var product Product
			err := json.Unmarshal(d.Body, &product)
			if err != nil {
				log.Fatal(err)
			}

			// Process the product here
			log.Printf("Received a product: %s", product)

			// Save the product as a JSON file
			productJson, err := json.Marshal(product)
			if err != nil {
				log.Fatal(err)
			}

			// Ensure the products directory exists
			err = os.MkdirAll("products", 0755)
			if err != nil {
				log.Fatal(err)
			}

			// Write the JSON data to a file
			err = ioutil.WriteFile(filepath.Join("products", fmt.Sprintf("%s.json", product.Name)), productJson, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	log.Printf(" [*] Waiting for products. To exit press CTRL+C")
	<-forever
}
