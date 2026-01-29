package server

import (
	"context"
	"messagefeedapp/common"
	pb "messagefeedapp/datastoreapp/openmedia/datastoreapp/protobuf"

	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StoreMessage(t *testing.T) {
	// Initialize FileClient with a test file path
	fileStorage, _ := common.NewFileClient("test_messages.txt")
	messageServer := &MessageServer{FileStorage: fileStorage}
	resp, _ := messageServer.StoreMessage(context.Background(), &pb.StoreMessageRequest{Message: "Test Message 1"})
	assert.Equal(t, resp.Success, true)
}
