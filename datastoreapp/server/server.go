package server

import (
	"context"

	log "github.com/sirupsen/logrus"

	"messagefeedapp/common"
	pb "messagefeedapp/datastoreapp/openmedia/datastoreapp/protobuf"
)

type MessageServer struct {
	pb.UnimplementedMessageServiceServer
	// Add your storage map here
	FileStorage *common.FileClient
}

func (s *MessageServer) StoreMessage(ctx context.Context, req *pb.StoreMessageRequest) (*pb.StoreMessageResponse, error) {
	err := s.FileStorage.WriteMessageToFile(req.GetMessage())
	if err != nil {
		log.WithContext(ctx).Errorf("failed to write message to file: %v", err)
		return &pb.StoreMessageResponse{Success: false}, err
	}
	return &pb.StoreMessageResponse{Success: true}, nil
}

func (s *MessageServer) RetrieveMessages(ctx context.Context, req *pb.RetrieveMessagesRequest) (*pb.RetrieveMessagesResponse, error) {
	var messages []string
	messages, err := s.FileStorage.RetrieveMessageFromFile(10)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to retrieve messages from file: %v", err)
		return &pb.RetrieveMessagesResponse{Messages: nil}, err
	}
	return &pb.RetrieveMessagesResponse{Messages: messages}, nil
}
