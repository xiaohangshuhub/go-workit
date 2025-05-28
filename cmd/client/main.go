package main

import (
	"context"
	"log"
	"time"

	pb "github.com/lxhanghub/go-workit/internal/service1/grpcapi/hello" // 替换为你的实际 import 路径
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewHelloServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "ChatGPT"})
	if err != nil {
		log.Fatalf("error calling SayHello: %v", err)
	}

	log.Printf("Response: %s", resp.Message)
}
