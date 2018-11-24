package main

import (
	"context"
	"log"
	"time"

	"github.com/catrossim/monbeat/manager/pb"

	"google.golang.org/grpc"
)

const (
	address = "localhost:31080"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewRemoteClient(conn)

	// Contact the server and print out its response.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	text := `
	#!/bin/bash
	ls /tmp
	`
	data := []byte(text)
	r, err := c.Execute(ctx)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return
	}
	err = r.Send(&pb.Chunk{
		Content: data,
	})
	if err != nil {
		log.Fatalf("cound not send data.")
	}
	status, err := r.CloseAndRecv()
	if err != nil {
		log.Fatalf("cound not get response.")
	}
	log.Println(status.GetResult())
}
