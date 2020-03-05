package main

import (
	"flag"
	"fmt"
	pqpb "github.com/songford/grpc-demo/print_queue_protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"net"
	"sync"
)

var port = flag.Int("port", 12345, "")

type catServiceServer struct {
	pqpb.UnimplementedGrpcCatServiceServer

	mu         sync.Mutex // protects routeNotes
	textThread map[string][]*pqpb.Text
}

func (css *catServiceServer) Chat(stream pqpb.GrpcCatService_ChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		key := in.Sender
		css.mu.Lock()
		css.textThread[key] = append(css.textThread[key], in)
		t := make([]*pqpb.Text, len(css.textThread[key]))
		copy(t, css.textThread[key])
		css.mu.Unlock()

		for _, text := range t {
			fmt.Printf("Received %s from %s\n", text.Text, text.Sender)
			if err := stream.Send(text); err != nil {
				return err
			}
		}
	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pqpb.RegisterGrpcCatServiceServer(grpcServer, newServer())
	reflection.Register(grpcServer)
	log.Printf("Server started. ")
	grpcServer.Serve(lis)
}

func newServer() *catServiceServer {
	s := &catServiceServer{textThread: make(map[string][]*pqpb.Text)}
	return s
}
