package main

import (
	"context"
	"dist-grpc/pkg/utils"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "dist-grpc/pkg/proto"
)

func init() {
	log.SetPrefix("GRPC Client")
	log.SetTimeFormat(time.TimeOnly)
}

const (
	defaultPort = 8080
	defaultHost = "localhost"
)

func printResponse(res *pb.Response) {
	fmt.Println("Response:")
	fmt.Println("\tTimestamp:", res.Timestamp.AsTime())
	fmt.Println("\tResults:", utils.ToString(res.Results))
}

func getOrderUnary(client pb.OrderManagementClient, query string) {
	log.Infof("Sending query as unary RPC: %s", query)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := pb.Request{Query: query}
	res, err := client.GetOrderUnary(ctx, &req)
	if err != nil {
		log.Fatalf("Failed to get order: %v", err)
	}
	log.Info("Received response")
	printResponse(res)
}

func getOrderServerStream(client pb.OrderManagementClient, query string) {
	log.Infof("Sending query as server stream RPC: %s", query)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := pb.Request{Query: query}
	stream, err := client.GetOrderServerStream(ctx, &req)
	if err != nil {
		log.Fatalf("Failed to get order: %v", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Warn("Received EOF, closing stream")
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive response: %v", err)
		}
		log.Info("Received single response")
		printResponse(resp)
	}
}

func getOrderClientStream(client pb.OrderManagementClient, queryChan chan string) {
	log.Info("Sending queries as client stream RPC")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.GetOrderClientStream(ctx)
	if err != nil {
		log.Fatalf("Failed to get order: %v", err)
		return
	}

	for query := range queryChan {
		log.Infof("Sending query: %s", query)
		if err := stream.Send(&pb.Request{Query: query}); err != nil {
			log.Fatalf("Failed to send request: %v", err)
		}
	}

	log.Info("Closing stream and receiving response")
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to receive response: %v", err)
	}
	log.Info("Received response")
	printResponse(resp)
}

func GetOrderBiDiStream(client pb.OrderManagementClient, queryChan chan string) {
	log.Info("Sending queries as bidirectional stream RPC")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.GetOrderBiDiStream(ctx)
	if err != nil {
		log.Fatalf("Failed to get order: %v", err)
		return
	}

	waitc := make(chan struct{})
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Warn("Received EOF, closing stream")
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive response: %v", err)
			}
			log.Info("Received single response")
			printResponse(resp)
		}
	}()

	for query := range queryChan {
		log.Printf("Sending query: %s", query)
		if err := stream.Send(&pb.Request{Query: query}); err != nil {
			log.Fatalf("Failed to send request: %v", err)
		}
	}
	err = stream.CloseSend()
	if err != nil {
		log.Fatalf("Failed to close send: %v", err)
	}
	<-waitc
}

func unaryRPC(client pb.OrderManagementClient) {
	var order string
	fmt.Print("> Enter query: ")
	scan, err := fmt.Scan(&order)
	if err != nil || scan != 1 {
		log.Error("Failed to read query")
		return
	}
	getOrderUnary(client, order)
}

func serverStreamRPC(client pb.OrderManagementClient) {
	var order string
	fmt.Print("> Enter query: ")
	scan, err := fmt.Scan(&order)
	if err != nil || scan != 1 {
		log.Error("Failed to read query")
		return
	}
	getOrderServerStream(client, order)
}

func clientStreamRPC(client pb.OrderManagementClient) {
	queryChan := make(chan string)
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			var order string
			fmt.Print("> Enter query (enter 'exit' to finish): ")
			scan, err := fmt.Scan(&order)
			if err != nil || scan != 1 {
				log.Error("Failed to read query")
				return
			}
			if order == "exit" {
				break
			}
			queryChan <- order
		}
		close(queryChan)
	}()
	getOrderClientStream(client, queryChan)
}

func bidirectionalStreamRPC(client pb.OrderManagementClient) {
	queryChan := make(chan string)
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			var order string
			fmt.Print("> Enter query (enter 'exit' to finish): ")
			scan, err := fmt.Scan(&order)
			if err != nil || scan != 1 {
				log.Error("Failed to read query")
				return
			}
			if order == "exit" {
				break
			}
			queryChan <- order
		}
		close(queryChan)
	}()
	GetOrderBiDiStream(client, queryChan)

}

func launchMenu(client pb.OrderManagementClient) {
	for {
		fmt.Println()
		fmt.Println("Choose the desired communication pattern:")
		fmt.Println("\t1. Unary RPC")
		fmt.Println("\t2. Server Streaming RPC")
		fmt.Println("\t3. Client Streaming RPC")
		fmt.Println("\t4. Bidirectional Streaming RPC")
		fmt.Println("\t5. Exit")

		fmt.Print("> Enter choice: ")
		var choice int
		scan, err := fmt.Scan(&choice)
		if err != nil || scan != 1 {
			log.Error("Failed to read choice")
			continue
		}

		switch choice {
		case 1:
			unaryRPC(client)
		case 2:
			serverStreamRPC(client)
		case 3:
			clientStreamRPC(client)
		case 4:
			bidirectionalStreamRPC(client)
		case 5:
			return
		default:
			log.Error("Invalid choice")
		}
	}
}

func main() {
	log.Info("Starting...")

	portPtr := flag.Int("port", defaultPort, "port to listen on")
	hostPtr := flag.String("host", defaultHost, "host to listen on")
	flag.Parse()
	port := *portPtr
	host := *hostPtr

	dialAddr := fmt.Sprintf("%s:%d", host, port)
	log.Infof("Dialing %s", dialAddr)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(dialAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("Failed to close connection: %v", err)
		}
	}()

	client := pb.NewOrderManagementClient(conn)
	launchMenu(client)
	log.Warn("Exiting...")
}
