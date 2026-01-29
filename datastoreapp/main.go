package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"messagefeedapp/common"
	pb "messagefeedapp/datastoreapp/openmedia/datastoreapp/protobuf"
	"messagefeedapp/datastoreapp/server"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fileStorage, err := common.NewFileClient(common.GetFilePath())
	if err != nil {
		log.Fatalf("failed to create file storage client: %v", err)
	}
	s := grpc.NewServer()
	messageServer := &server.MessageServer{FileStorage: fileStorage}
	pb.RegisterMessageServiceServer(s, messageServer)
	go runClient(messageServer)
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Wait for termination signal
	<-sigChan
	fileStorage.Close()

	log.Info("shutting down server...")
	s.GracefulStop()

}
func runClient(s *server.MessageServer) {

	for {
		var userId, message string
		log.Print("Enter userId: ")
		fmt.Scanln(&userId)
		log.Print("Enter message: ")
		fmt.Scanln(&message)

		storeReq := &pb.StoreMessageRequest{
			Message: message,
		}
		storeResp, err := s.StoreMessage(context.Background(), storeReq)
		if err != nil {
			log.Errorf("failed to store message: %v", err)
			continue
		}
		log.Printf("Store result: %v", storeResp.Success)

		retrieveReq := &pb.RetrieveMessagesRequest{}
		retrieveResp, err := s.RetrieveMessages(context.Background(), retrieveReq)
		if err != nil {
			log.Errorf("failed to retrieve messages: %v", err)
			continue
		}
		log.Printf("Retrieved messages: %v", retrieveResp.Messages)
	}
}
