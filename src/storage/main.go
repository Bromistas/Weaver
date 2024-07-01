package main

import (
	"commons"
	"context"
	"encoding/json"
	"fmt"
	"github.com/grandcat/zeroconf"
	"google.golang.org/grpc"
	"log"
	"net"
	"node"
	"os"
	pb "protos"
	"strings"
	"sync"
	"time"
)

func put_pair(addr, k, v string, group *sync.WaitGroup) {
	defer group.Done()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("connection to %v failed: %v", addr, err)
	}
	defer conn.Close()
	c := pb.NewDHTClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	_, err = c.Put(ctx, &pb.Pair{Key: k, Value: v})
	if err != nil {
		log.Fatalf("could not put to %v: %v", addr, err)
	}
}

// CustomPut is a function that unmarshals the payload into a Product object and writes it to a JSON file
func CustomPut(ctx context.Context, pair *pb.Pair) error {
	// Unmarshal the payload into a Product object

	fmt.Println("Printing custom put")

	var product common.Product
	err := json.Unmarshal([]byte(pair.Value), &product)
	if err != nil {
		log.Fatalf("Failed to unmarshal payload: %v", err)
		return err
	}

	// Write the Product object to a JSON file
	productJson, err := json.Marshal(product)
	if err != nil {
		log.Fatalf("Failed to marshal product: %v", err)
		return err
	}

	filename := product.Name + ".json"
	err = os.WriteFile(filename, productJson, 0644)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
		return err
	}

	log.Printf("Product written to file: %v", product)

	return nil
}

func ServeChordWrapper(n *node.ChordNode, bootstrap *node.ChordNode, group *sync.WaitGroup) {
	fmt.Println("[*] Node1 started")
	node.ServeChord(context.Background(), n, bootstrap, group, nil)
	fmt.Println("Node1 joined the network")
}

func mainWrapper(address string, port int, group *sync.WaitGroup, wait time.Duration) {
	node1 := node.NewChordNode(address, CustomPut)

	found_ip := ""
	found_port := 0
	discoveryCallback := func(entry *zeroconf.ServiceEntry) {
		if strings.HasPrefix(entry.ServiceInstanceName(), "Storage") {

			if len(entry.AddrIPv4) == 0 {
				log.Printf("Found service: %s, but no IP address", entry.ServiceInstanceName(), ". Going localhost")
				found_ip = "localhost"
				found_port = entry.Port
			} else {
				found_ip = entry.AddrIPv4[0].String()
				found_port = entry.Port
			}
			log.Printf("Registered service: %s, IP: %s, Port: %d\n", entry.ServiceInstanceName(), entry.AddrIPv4, entry.Port)
		}
	}

	common.Discover("_http._tcp", "local.", 5*time.Second, discoveryCallback)

	group.Add(1)
	if found_ip != "" {
		fmt.Println("Found storage node, joining the ring")
		node2 := node.NewChordNode(found_ip+":"+fmt.Sprint(found_port), CustomPut)
		go ServeChordWrapper(node1, node2, group)
	} else {
		fmt.Println("No storage node found, starting a new ring")
		go ServeChordWrapper(node1, nil, group)
	}

	serviceName := "StorageNode"
	serviceType := "_http._tcp"
	domain := "local."
	ip := net.ParseIP("127.0.0.1")

	err := common.RegisterForDiscovery(serviceName, serviceType, domain, port, ip)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	group := &sync.WaitGroup{}

	go mainWrapper("127.0.0.1:50051", 50051, group, 5*time.Second)
	time.Sleep(7 * time.Second)
	go mainWrapper("127.0.0.1:50052", 50052, group, 5*time.Second)

	group.Wait()
}
