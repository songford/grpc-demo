package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	pqpb "grpc-demo/print_queue_protobuf"
	"io"
	"log"
	"os"
)

var (
	serverAddr        = flag.String("server_addr", "localhost:12345", "")
	messageOutChannel = make(chan pqpb.Text)
)

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	reader := bufio.NewReader(os.Stdin)

	defer conn.Close()
	client := pqpb.NewGrpcCatServiceClient(conn)

	ctx := context.Background()
	stream := initiateRouteChat(client, ctx)

	go sendChat(stream)
	go receiveChat(stream)

	for {
		fmt.Println("Your name is?")
		sender, _ := reader.ReadString('\n')
		fmt.Println("Your message is?")
		message, _ := reader.ReadString('\n')
		text := pqpb.Text{
			Text:   message,
			Sender: sender,
		}
		fmt.Printf("%+v\n", text)
		messageOutChannel <- text
	}
}

func initiateRouteChat(client pqpb.GrpcCatServiceClient, ctx context.Context) pqpb.GrpcCatService_ChatClient {
	stream, err := client.Chat(ctx)
	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}
	return stream
}

func sendChat(stream pqpb.GrpcCatService_ChatClient) {
	for {
		text := <-messageOutChannel
		if err := stream.Send(&text); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}
}

func receiveChat(stream pqpb.GrpcCatService_ChatClient) {
	for {
		waitc := make(chan struct{})
		t, err := stream.Recv()
		if err != nil {
			log.Fatalf("Failed to receive a chat message : %v", err)
		}
		if err == io.EOF {
			// read done.
			close(waitc)
			return
		}
		fmt.Sprintf("Got message pinged back, sent by %s, content is: %s", t.Sender, t.Text)
		<-waitc
	}
}
