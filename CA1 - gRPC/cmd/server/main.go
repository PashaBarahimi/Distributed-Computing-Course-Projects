package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"dist-grpc/pkg/matcher"
	pb "dist-grpc/pkg/proto"
	"dist-grpc/pkg/utils"
)

const port = 8080

type orderManagementServer struct {
	pb.UnimplementedOrderManagementServer
}

func (s *orderManagementServer) GetOrderUnary(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	res := matcher.MatchOrder(req.GetQuery())
	return &pb.Response{
		Results:   res,
		Timestamp: timestamppb.Now(),
	}, nil
}

func (s *orderManagementServer) GetOrderServerStream(req *pb.Request, stream pb.OrderManagement_GetOrderServerStreamServer) error {
	res := matcher.MatchOrder(req.GetQuery())
	for _, v := range res {
		resp := pb.Response{
			Results:   []string{v},
			Timestamp: timestamppb.Now(),
		}
		if err := stream.Send(&resp); err != nil {
			return err
		}
	}
	return nil
}

func (s *orderManagementServer) GetOrderClientStream(stream pb.OrderManagement_GetOrderClientStreamServer) error {
	var res []string
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		list := matcher.MatchOrder(req.GetQuery())
		res = append(res, list...)
	}
	return stream.SendAndClose(&pb.Response{
		Results:   utils.RemoveDuplicates(res),
		Timestamp: timestamppb.Now(),
	})
}

func (s *orderManagementServer) GetOrderBiDiStream(stream pb.OrderManagement_GetOrderBiDiStreamServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		res := matcher.MatchOrder(req.GetQuery())
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
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("net listen failed: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterOrderManagementServer(grpcServer, &orderManagementServer{})
	grpcServer.Serve(listener)
}
