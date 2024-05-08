package main

import (
	"context"
	"fmt"
	"github.com/gocolly/colly/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strings"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func scrap(url string) {
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

	urlsToScrap := []string{
		"https://www.newegg.com/abyss-blue-lenovo-ideapad-slim-5i-82xf002sus-home-personal/p/N82E16834840212",
		"https://www.newegg.com/indie-black-asus-f1605va-ds74-home-personal/p/N82E16834236434?Item=9SIA7ABK6P6769",
		"https://www.newegg.com/natural-silver-natural-silver-with-a-vertical-brushed-pattern-hp-15-dy2088ca-mainstream/p/N82E16834272966?Item=9SIA9UWK084845",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, url := range urlsToScrap {
		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(url),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s\n", url)
	}

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

	var forever chan struct{}

	go func() {
		for d := range msgs {
			url := string(d.Body)
			scrap(url)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
