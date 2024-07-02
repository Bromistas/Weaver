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
	"strconv"
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

func CheckIps() (net.IP, error) {
	// Get the default network interface
	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Get the IP addresses associated with this interface
	addrs, err := iface.Addrs()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var ip net.IP
	for _, addr := range addrs {

		if strings.Contains(addr.String(), "::") {
			continue
		}

		switch v := addr.(type) {
		case *net.IPNet:
			//if strings.Contains(v.String(), "::"){
			ip = v.IP
			//}
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil {
			fmt.Printf("%s has IP %s\n", iface.Name, ip.String())
		}
	}

	return ip, nil
}

func mainWrapper(group *sync.WaitGroup) {
	defer group.Done()

	ip, _ := CheckIps()
	//ip := net.ParseIP("127.0.0.1")

	//address := os.Getenv("ADDRESS")
	address := ip.String() + ":" + os.Getenv("PORT")
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	waitTime, _ := time.ParseDuration(os.Getenv("WAIT_TIME"))

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
				temp := entry.AddrIPv4[0].String()

				if !strings.Contains(found_ip, "::") {
					found_ip = temp
					found_port = entry.Port
				}
			}
			log.Printf("Registered service: %s, IP: %s, Port: %d\n", entry.ServiceInstanceName(), entry.AddrIPv4, entry.Port)
		}
	}

	common.Discover("_http._tcp", "local.", waitTime, discoveryCallback)

	if found_ip != "" {
		fmt.Println("Found storage node, joining the ring")
		node2 := node.NewChordNode(found_ip+":"+fmt.Sprint(found_port), CustomPut)
		go ServeChordWrapper(node1, node2, group)
	} else {
		fmt.Println("No storage node found, starting a new ring")
		go ServeChordWrapper(node1, nil, group)
	}

	// TODO: change this is hardcoded for localhost
	serviceName := "StorageNode"
	serviceType := "_http._tcp"
	domain := "local."

	err := common.RegisterForDiscovery(serviceName, serviceType, domain, port, ip)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	group := &sync.WaitGroup{}

	//os.Setenv("ADDRESS", "127.0.0.1:50051")
	//os.Setenv("PORT", "50051")
	//os.Setenv("WAIT_TIME", "2s")

	group.Add(1)
	go mainWrapper(group)

	group.Wait()
}
