package main

import (
	common "commons"
	"encoding/json"
	"github.com/grandcat/zeroconf"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"strings"
	"time"
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

	node := &ScrapperNode{}

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
				node.AmazonProductHandler(urlMessage.URL)
			case NeweggProduct:
				NeweggProductHandler(urlMessage.URL)
			case NeweggRoot:
				NeweggRootHandler(urlMessage.URL, ch, q)
			default:
				log.Printf("Unknown URL type: %v", urlMessage.URLType)
			}
		}
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Make a discovery and listen for services

	discoveryCallback := func(entry *zeroconf.ServiceEntry) {
		if strings.HasPrefix(entry.ServiceInstanceName(), "Storage") {
			if len(entry.AddrIPv4) == 0 {
				log.Printf("Found service: %s, but no IP address", entry.ServiceInstanceName(), ". Going localhost")
				node.StorageAddress = "localhost"
				node.StoragePort = entry.Port
			}
			log.Printf("Registered service: %s, IP: %s, Port: %d\n", entry.ServiceInstanceName(), entry.AddrIPv4, entry.Port)
		}
	}

	common.Discover("_http._tcp", "local.", 5*time.Second, discoveryCallback)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
