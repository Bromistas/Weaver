package main

import (
	"chord"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		// No IP argument provided, create the first node's configuration and transport
		config := chord.DefaultConfig("127.0.0.1:8000")
		transport, err := chord.InitTCPTransport("127.0.0.1:8000", 4*time.Second)
		if err != nil {
			log.Fatalf("Failed to create transport: %v", err)
		}
		node, err := chord.Create(config, transport)
		if err != nil {
			log.Fatalf("Failed to create node: %v", err)
		}
		defer node.Leave()
		fmt.Println("Created a ring on 127.0.0.1:8000")

		for {
			printRingInfo(node)
			fmt.Println()
			time.Sleep(5 * time.Second)
		}

	} else {
		// IP argument provided, create a node with the provided IP and join it to the ring on port 8000
		ip := os.Args[1]
		config := chord.DefaultConfig(ip)
		transport, err := chord.InitTCPTransport(ip, 4*time.Second)
		if err != nil {
			log.Fatalf("Failed to create transport for node: %v", err)
		}
		node, err := chord.Join(config, transport, "127.0.0.1:8000")
		if err != nil {
			log.Fatalf("Failed for node to join the ring: %v", err)
		}
		defer node.Leave()
		fmt.Printf("Node with IP %s has successfully joined the ring on 127.0.0.1:8000\n", ip)
		//ticks, err := strconv.Atoi(os.Args[2])
		i := 0
		for {
			i += 1
			printRingInfo(node)
			fmt.Println()
			time.Sleep(5 * time.Second)
			//if i == ticks {
			//	node.Leave()
			//}
		}

	}

	// Keep the nodes running
	select {}
}

// Function to look for a key in the ring
func lookupKey(ring *chord.Ring, key []byte) {
	successors, err := ring.Lookup(1, key) // Assuming you want the closest successor
	if err != nil {
		log.Fatalf("Lookup failed: %v", err)
	}
	for _, succ := range successors {
		fmt.Printf("Successor ID: %s\n", succ.Id)
	}
}

// Function to print predecessors and successors for all nodes in the ring
func printRingInfo(ring *chord.Ring) {
	for _, vnode := range ring.Vnodes {
		fmt.Printf("Vnode ID: %v\n", vnode.Id)
		if vnode.Predecessor != nil {
			fmt.Printf("\tPredecessor: %v\n", vnode.Predecessor.Host)
		}
	}
}
