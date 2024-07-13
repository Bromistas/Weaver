package main

import (
	common "commons"
	"log"
	"os"
	"strconv"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func discoverQueue(node *ScrapperNode, panicOnFail bool) {
	//port := os.Getenv("PORT")
	foundAddr, _ := common.NetDiscover("9000", "QUEUE")

	node.QueueAddress = foundAddr
	node.QueuePort = 9000

	//temp, err := strconv.Atoi(foundPort)
	//if err != nil {
	//	log.Panicf("Failed to convert in queue discover port to int: %v", err)
	//}
	//

	//discoverQueueCallback := func(entry *zeroconf.ServiceEntry) {
	//	if strings.HasPrefix(entry.ServiceInstanceName(), "Queue") {
	//		if len(entry.AddrIPv4) == 0 {
	//			log.Printf("Found service: %s, but no IP address. Going localhost", entry.ServiceInstanceName())
	//			node.QueueAddress = "localhost"
	//			node.QueuePort = entry.Port
	//		}
	//		log.Printf("Registered service: %s, IP: %s, Port: %d\n", entry.ServiceInstanceName(), entry.AddrIPv4, entry.Port)
	//	}
	//}
	//
	//common.Discover("_http._tcp", "local.", 5*time.Second, discoverQueueCallback)
	//
	//if panicOnFail && (node.QueueAddress == "" || node.QueuePort == 0) {
	//	log.Panicf("Failed to discover Queue service")
	//}
}

func discoverStorage(node *ScrapperNode, panicOnFail bool) {

	port := os.Getenv("PORT")
	foundAddr, _ := common.NetDiscover(port, "STORAGE")

	//
	node.StorageAddress = foundAddr
	node.StoragePort = 10000
	//
	//temp, err := strconv.Atoi(foundPort)
	//if err != nil {
	//	log.Panicf("Failed to convert in storage discover port to int: %v", err)
	//}
	//
	//node.QueuePort = temp
	//
	//discoveryCallback := func(entry *zeroconf.ServiceEntry) {
	//	if strings.HasPrefix(entry.ServiceInstanceName(), "Storage") {
	//		if len(entry.AddrIPv4) == 0 {
	//			log.Printf("Found service: %s, but no IP address. Going localhost", entry.ServiceInstanceName())
	//			node.StorageAddress = "localhost"
	//			node.StoragePort = entry.Port
	//		}
	//		log.Printf("Registered service: %s, IP: %s, Port: %d\n", entry.ServiceInstanceName(), entry.AddrIPv4, entry.Port)
	//	}
	//}
	//
	//common.Discover("_http._tcp", "local.", 5*time.Second, discoveryCallback)
	//
	//if panicOnFail && (node.StorageAddress == "" || node.StoragePort == 0) {
	//	log.Panicf("Failed to discover Storage service")
	//}
}

func healthCheckQueue(node *ScrapperNode) {
	for {
		queueService := NewQueueServiceClient(node.QueueAddress + ":" + strconv.Itoa(node.QueuePort))
		_, err := queueService.HealthCheck()
		if err != nil {
			log.Println("Queue service health check failed. Rediscovering...")
			discoverQueue(node, false)
			queueService = NewQueueServiceClient(node.QueueAddress + ":" + strconv.Itoa(node.QueuePort))
		}
		time.Sleep(1 * time.Second) // Adjust the sleep duration as needed
	}
}

func healthCheckStorage(node *ScrapperNode) {
	for {
		storageService := NewStorageServiceClient(node.StorageAddress + ":" + strconv.Itoa(node.StoragePort))
		_, err := storageService.HealthCheck()
		if err != nil {
			log.Println("Storage service health check failed. Rediscovering...")
			discoverStorage(node, false)
			storageService = NewStorageServiceClient(node.StorageAddress + ":" + strconv.Itoa(node.StoragePort))
		}

		time.Sleep(1 * time.Second) // Adjust the sleep duration as needed
	}
}

func main() {
	node := &ScrapperNode{}

	discoverQueue(node, false)
	discoverStorage(node, false)

	queueService := NewQueueServiceClient(node.QueueAddress + ":" + strconv.Itoa(node.QueuePort))

	time.Sleep(1 * time.Second)

	//time.Sleep(6 * time.Second)
	//go healthCheckQueue(node)
	//go healthCheckStorage(node)

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")

	err := queueService.Listen(time.Second, node)
	if err != nil {
		log.Panicf("[!] Error listening to queue: %s", err)
	}

}
