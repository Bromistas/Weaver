package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	pb "protos"
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
