package main

import (
	"bufio"
	"bytes"
	common "commons"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

func getQueueUrl() string {
	foundAddr, err := common.NetDiscover("9000", "QUEUE", true)
	if err != nil || foundAddr == "" {
		log.Fatalf("Not found queue %s", err)
	}

	return foundAddr
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to the CLI tool! Type 'exit' to quit.")

	for {
		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Add your command processing logic here
		processInput(input)
	}
}

func processInput(input string) {
	args := strings.Fields(input)
	if len(args) == 0 {
		return
	}

	command := args[0]
	params := args[1:]

	switch command {
	case "scrap":
		scrap(params)
	case "help":
		fmt.Println("Available commands: hello, echo, help, exit")
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func scrap(params []string) {
	if len(params) != 1 {
		fmt.Println("Usage: cli scrap <url>")
		return
	}

	urlToScrap := params[0]
	baseUrl := getQueueUrl()
	url := fmt.Sprintf("http://%s:9000/put", baseUrl)
	body := map[string]string{"message": urlToScrap}
	jsonBody, err := json.Marshal(body)
	failOnError(err, "Failed to marshal JSON")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	failOnError(err, "Failed to send POST request")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	fmt.Println("URL scrapped successfully")
}
