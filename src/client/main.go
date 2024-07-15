package main

import (
	"bufio"
	"bytes"
	common "commons"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type Product struct {
	Name        string
	Price       float64
	URL         string
	Description string
	Rating      string
	Addresses   []string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func getQueueUrl() string {
	foundAddr, err := common.NetDiscover("9000", "QUEUE", true, false)
	if err != nil || foundAddr[0] == "" {
		log.Fatalf("Not found queue %s", err)
	}

	return foundAddr[0]
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
	case "gather":
		gather()
	case "help":
		fmt.Println("Available commands: scrap, gather, help, exit")
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func scrap(params []string) {
	if len(params) != 1 {
		fmt.Println("Usage: cli scrap <query>")
		return
	}

	query := params[0]
	// Construct search URLs for Amazon and Newegg
	amazonURL := fmt.Sprintf("https://www.amazon.com/s?k=%s", url.QueryEscape(query))
	neweggURL := fmt.Sprintf("https://www.newegg.com/p/pl?d=%s", url.QueryEscape(query))

	// Create URL messages for Amazon and Newegg with the appropriate types
	amazonURLMessage := common.URLMessage{URL: amazonURL, URLType: common.Dummy}
	neweggURLMessage := common.URLMessage{URL: neweggURL, URLType: common.NeweggRoot}

	// Insert Amazon and Newegg URL messages into the queue
	insertIntoQueue(amazonURLMessage)
	insertIntoQueue(neweggURLMessage)

	fmt.Println("Search URL messages inserted into queue successfully")
}

// insertIntoQueue inserts the given URL message into the queue
func insertIntoQueue(urlMessage common.URLMessage) {
	baseUrl := getQueueUrl()
	url := fmt.Sprintf("http://%s:9000/put", baseUrl)
	jsonBody, err := common.EncodeURLMessage(urlMessage)
	failOnError(err, "Failed to marshal JSON")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	failOnError(err, "Failed to send POST request")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}
}

func gather() {
	storeIps, err := common.NetDiscover("10000", "STORAGE", false, true)
	if err != nil {
		log.Fatalf("Error while discovering storage nodes %s", err.Error())
	}

	fmt.Printf("[*] While gathering, found %d different storage nodes\n", len(storeIps))

	productMap := make(map[string]*Product)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, ip := range storeIps {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			url := fmt.Sprintf("http://%s:10000/gather", ip)
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Error fetching from %s: %s", ip, err.Error())
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response from %s: %s", ip, err.Error())
				return
			}

			var products []Product
			if err := json.Unmarshal(body, &products); err != nil {
				log.Printf("Error unmarshalling response from %s: %s", ip, err.Error())
				return
			}

			mu.Lock()
			defer mu.Unlock()
			for _, product := range products {
				if existingProduct, ok := productMap[product.URL]; ok {
					existingProduct.Addresses = append(existingProduct.Addresses, ip)
				} else {
					newProduct := Product{
						Name:        product.Name,
						Price:       product.Price,
						URL:         product.URL,
						Description: product.Description,
						Rating:      product.Rating,
						Addresses:   []string{ip}, // Initialize with current IP
					}

					productMap[product.URL] = &newProduct
				}
			}
		}(ip)
	}

	wg.Wait()

	// Printing the table
	fmt.Println("URL\tName\tDescription\tPrice\tAddresses")
	for _, product := range productMap {
		fmt.Printf("%s\t%s\t%s\t%.2f\t%s\n", product.URL, product.Name, product.Description, product.Price, strings.Join(product.Addresses, ", "))
	}
}

// contains checks if a slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
