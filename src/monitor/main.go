package main

import (
	"fmt"
	common "github.com/bromistas/weaver-commons"
	"log"
	"time"
)

func main() {

	node1 := common.NewNode(1, "localhost:8080")
	server1 := common.NewServer(node1, "localhost:8080")
	go func() {
		fmt.Println("[x] Server 1 started")
		err := server1.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	node2 := common.NewNode(2, "localhost:8081")
	server2 := common.NewServer(node2, "localhost:8081")
	go func() {
		fmt.Println("[x] Server 2 started")
		err := server2.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	node3 := common.NewNode(3, "localhost:8082")
	server3 := common.NewServer(node3, "localhost:8082")
	go func() {
		fmt.Println("[x] Server 3 started")
		err := server3.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(1 * time.Second)
	//addr, err := net.ResolveTCPAddr("tcp", "localhost:8081")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err := server1.Join("localhost:8081")
	//if err != nil {
	//	log.Fatal(err)
	//}

	// Keep the main function running
	select {}

}
