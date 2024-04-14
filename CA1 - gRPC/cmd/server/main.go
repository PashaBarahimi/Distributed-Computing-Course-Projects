package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/charmbracelet/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"dist-grpc/pkg/matcher"
	pb "dist-grpc/pkg/proto"
	"dist-grpc/pkg/utils"
)

func init() {
	log.SetPrefix("GRPC Server")
	log.SetTimeFormat(time.TimeOnly)
}

const (
	defaultPort = 8080
	defaultHost = "localhost"
)

type orderManagementServer struct {
	pb.UnimplementedOrderManagementServer
}

func (s *orderManagementServer) GetOrderUnary(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	log.Info("Received unary request", "query", req.GetQuery())
	res := matcher.MatchOrder(req.GetQuery())
	log.Info("Matched orders", "orders", utils.ToString(res))
	log.Info("Sending response", "results", utils.ToString(res))
	return &pb.Response{
		Results:   res,
		Timestamp: timestamppb.Now(),
	}, nil
}

func (s *orderManagementServer) GetOrderServerStream(req *pb.Request, stream pb.OrderManagement_GetOrderServerStreamServer) error {
	log.Info("Received server stream request", "query", req.GetQuery())
	res := matcher.MatchOrder(req.GetQuery())
	log.Info("Matched orders", "orders", utils.ToString(res))
	for _, v := range res {
		resp := pb.Response{
			Results:   []string{v},
			Timestamp: timestamppb.Now(),
		}
		log.Info("Sending single response", "results", v)
		if err := stream.Send(&resp); err != nil {
			return err
		}
	}
	return nil
}

func (s *orderManagementServer) GetOrderClientStream(stream pb.OrderManagement_GetOrderClientStreamServer) error {
	log.Info("Received client stream request")
	var res []string
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Warn("Received EOF, closing stream")
			break
		}
		if err != nil {
			return err
		}
		log.Info("Received request", "query", req.GetQuery())
		list := matcher.MatchOrder(req.GetQuery())
		log.Info("Matched orders", "orders", utils.ToString(list))
		res = append(res, list...)
	}
	result := utils.RemoveDuplicates(res)
	log.Info("Sending response", "results", utils.ToString(result))
	return stream.SendAndClose(&pb.Response{
		Results:   result,
		Timestamp: timestamppb.Now(),
	})
}

func (s *orderManagementServer) GetOrderBiDiStream(stream pb.OrderManagement_GetOrderBiDiStreamServer) error {
	log.Info("Received bidirectional stream request")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Warn("Received EOF, closing stream")
			return nil
		}
		if err != nil {
			return err
		}
		log.Info("Received request", "query", req.GetQuery())
		res := matcher.MatchOrder(req.GetQuery())
		log.Info("Matched orders", "orders", utils.ToString(res))
		log.Info("Sending response", "results", utils.ToString(res))
		err = stream.Send(&pb.Response{
			Results:   res,
			Timestamp: timestamppb.Now(),
		})
		if err != nil {
			return err
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

	listenAddr := fmt.Sprintf("%s:%d", host, port)
	log.Infof("Listening on %s", listenAddr)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterOrderManagementServer(grpcServer, &orderManagementServer{})
	if err = grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
