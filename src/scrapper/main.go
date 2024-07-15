package main

import (
	common "commons"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var globalQueueAddresses = make(map[string]bool)
var globalQueueAddressesMutex = &sync.Mutex{}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func discoverStorage(node *ScrapperNode, panicOnFail bool) {

	port := os.Getenv("PORT")
	foundAddr, _ := common.NetDiscover(port, "STORAGE", false, false)

	node.StorageAddress = foundAddr[0]
	node.StoragePort = 10000

	log.Printf("Found storage node with address %s", node.StorageAddress)
}

func healthCheckStorage(node *ScrapperNode) {
	storageService := NewStorageServiceClient(node.StorageAddress + ":" + strconv.Itoa(node.StoragePort))
	for {
		_, err := storageService.HealthCheck()
		if err != nil {
			log.Println("Storage service health check failed. Rediscovering...")
			discoverStorage(node, false)
			storageService = NewStorageServiceClient(node.StorageAddress + ":" + strconv.Itoa(node.StoragePort))
		}
	}
}

func discoverQueue(node *ScrapperNode) {
	for {
		foundAddr, _ := common.NetDiscover("9000", "QUEUE", false, true)
		globalQueueAddressesMutex.Lock()
		for _, addr := range foundAddr {
			if _, exists := globalQueueAddresses[addr]; !exists {
				globalQueueAddresses[addr] = true
				go startReadingFromQueue(addr, node)
			}
		}
		globalQueueAddressesMutex.Unlock()
		time.Sleep(10 * time.Second) // Adjust the sleep duration as needed
	}
}

func readFromQueue(queueService *QueueServiceClient, node *ScrapperNode, wg *sync.WaitGroup, addr string) {
	defer wg.Done()
	err := queueService.Listen(time.Second, node)
	if err != nil {
		log.Printf("Error listening to queue: %s", err)
		globalQueueAddressesMutex.Lock()
		delete(globalQueueAddresses, addr)
		globalQueueAddressesMutex.Unlock()
	}
}

func startReadingFromQueue(addr string, node *ScrapperNode) {
	log.Printf("Starting to read from new queue: %s", addr+":9000")
	queueService := NewQueueServiceClient(addr + ":9000")
	var wg sync.WaitGroup
	wg.Add(1)
	go readFromQueue(queueService, node, &wg, addr)
	wg.Wait()
}

func main() {
	defer fmt.Printf("Stopping scrapper node")

	log.Printf("Launching scrapper")

	time.Sleep(7 * time.Second)

	node := &ScrapperNode{}
	discoverStorage(node, false)
	go healthCheckStorage(node)

	discoverQueue(node)
}
