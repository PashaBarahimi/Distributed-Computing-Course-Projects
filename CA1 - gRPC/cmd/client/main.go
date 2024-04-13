package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "dist-grpc/pkg/proto"
)

const port = 8080

func getOrderUnary(client pb.OrderManagementClient, order string) {
	log.Printf("getting order: %s", order)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := pb.Request{Query: order}
	res, err := client.GetOrderUnary(ctx, &req)
	if err != nil {
		log.Fatalf("get order unary failed: %v", err)
	}
	log.Println(res)
}

func getOrderServerStream(client pb.OrderManagementClient, order string) {
	log.Printf("getting order: %s", order)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := pb.Request{Query: order}
	stream, err := client.GetOrderServerStream(ctx, &req)
	if err != nil {
		log.Fatalf("get order server stream failed: %v", err)
		return
	}

	res := []*pb.Response{}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("get order server stream failed: %v", err)
		}
		res = append(res, resp)
	}
	log.Println(res)
}

func getOrderClientStream(client pb.OrderManagementClient, orders []string) {
	log.Printf("getting orders: %s", orders)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqs := []pb.Request{}
	for _, v := range orders {
		reqs = append(reqs, pb.Request{Query: v})
	}

	stream, err := client.GetOrderClientStream(ctx)
	if err != nil {
		log.Fatalf("get order client stream failed: %v", err)
		return
	}
	for _, req := range reqs {
		if err := stream.Send(&req); err != nil {
			log.Fatalf("get order client stream send failed: %v", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("get order client stream response failed: %v", err)
	}
	log.Println(resp)
}

func GetOrderBiDiStream(client pb.OrderManagementClient, orders []string) {
	log.Printf("getting orders: %s", orders)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqs := []pb.Request{}
	for _, v := range orders {
		reqs = append(reqs, pb.Request{Query: v})
	}

	stream, err := client.GetOrderBiDiStream(ctx)
	if err != nil {
		log.Fatalf("get order bidi stream failed: %v", err)
		return
	}

	waitc := make(chan struct{})
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("get order bidi stream recv failed: %v", err)
			}
			log.Println(resp)
		}
	}()

	for _, req := range reqs {
		time.Sleep(1 * time.Second)
		log.Printf("sending order: %s", req.GetQuery())
		if err := stream.Send(&req); err != nil {
			log.Fatalf("get order bidi stream send failed: %v", err)
		}
	}
	stream.CloseSend()
	<-waitc
}

func main() {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), opts...)
	if err != nil {
		log.Fatalf("grpc dial failed: %v", err)
	}
	defer conn.Close()
	client := pb.NewOrderManagementClient(conn)
	getOrderUnary(client, "apple")
	getOrderServerStream(client, "apple")
	getOrderClientStream(client, []string{"apple", "banana"})
	GetOrderBiDiStream(client, []string{"apple", "banana", "cherry", "kiwi"})
}
