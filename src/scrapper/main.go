package main

import (
	common "commons"
	"github.com/grandcat/zeroconf"
	"log"
	"strconv"
	"strings"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func discoverQueue(node *ScrapperNode, panicOnFail bool) {
	discoverQueueCallback := func(entry *zeroconf.ServiceEntry) {
		if strings.HasPrefix(entry.ServiceInstanceName(), "Queue") {
			if len(entry.AddrIPv4) == 0 {
				log.Printf("Found service: %s, but no IP address. Going localhost", entry.ServiceInstanceName())
				node.QueueAddress = "localhost"
				node.QueuePort = entry.Port
			}
			log.Printf("Registered service: %s, IP: %s, Port: %d\n", entry.ServiceInstanceName(), entry.AddrIPv4, entry.Port)
		}
	}

	common.Discover("_http._tcp", "local.", 5*time.Second, discoverQueueCallback)

	if panicOnFail && (node.QueueAddress == "" || node.QueuePort == 0) {
		log.Panicf("Failed to discover Queue service")
	}
}

func discoverStorage(node *ScrapperNode, panicOnFail bool) {
	discoveryCallback := func(entry *zeroconf.ServiceEntry) {
		if strings.HasPrefix(entry.ServiceInstanceName(), "Storage") {
			if len(entry.AddrIPv4) == 0 {
				log.Printf("Found service: %s, but no IP address. Going localhost", entry.ServiceInstanceName())
				node.StorageAddress = "localhost"
				node.StoragePort = entry.Port
			}
			log.Printf("Registered service: %s, IP: %s, Port: %d\n", entry.ServiceInstanceName(), entry.AddrIPv4, entry.Port)
		}
	}

	common.Discover("_http._tcp", "local.", 5*time.Second, discoveryCallback)

	if panicOnFail && (node.StorageAddress == "" || node.StoragePort == 0) {
		log.Panicf("Failed to discover Storage service")
	}
}

func main() {
	node := &ScrapperNode{}

	discoverQueue(node, true)
	discoverStorage(node, true)

	queueService := NewQueueServiceClient(node.QueueAddress + ":" + strconv.Itoa(node.QueuePort))

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")

	err := queueService.Listen(time.Second, node)
	if err != nil {
		log.Panicf("[!] Error listening to queue: %s", err)
	}

}
