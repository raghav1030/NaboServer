package service

import (
	"context"
	"fmt"
	"log"

	"github.com/raghav1030/NaboServer/apps/chat-service/internal/db"
	chatv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/chat/v1"
)

type ChatService struct {
	kafkaManager *db.KafkaManager
	// Add room manager, DB, etc. as needed
}

func NewChatService(kafkaManager *db.KafkaManager) *ChatService {
	return &ChatService{kafkaManager: kafkaManager}
}

func (s *ChatService) HandleSendMessage(ctx context.Context, req *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	var topic string
	var msgType string

	switch m := req.Message.(type) {
	case *chatv1.SendMessageRequest_Direct:
		topic = fmt.Sprintf("user-%s", m.Direct.ToUserId) // One-to-one topic
		msgType = "direct"
	case *chatv1.SendMessageRequest_Group:
		topic = fmt.Sprintf("group-%s", m.Group.GroupId) // Group chat topic
		msgType = "group"
	case *chatv1.SendMessageRequest_Ephemeral:
		topic = fmt.Sprintf("activity-%s", m.Ephemeral.ActivityId) // Ephemeral chat topic
		msgType = "ephemeral"
	default:
		return &chatv1.SendMessageResponse{Success: false, Message: "Unknown message type"}, nil
	}

	// Serialize the message (use protobuf marshal)
	data, err := s.kafkaManager.MarshalMessage(req)
	if err != nil {
		log.Printf("Failed to marshal chat message: %v", err)
		return &chatv1.SendMessageResponse{Success: false, Message: "Serialization error"}, err
	}

	// Publish to Kafka
	if err := s.kafkaManager.PublishMessage(ctx, topic, data); err != nil {
		log.Printf("Failed to publish to Kafka: %v", err)
		return &chatv1.SendMessageResponse{Success: false, Message: "Kafka publish error"}, err
	}

	return &chatv1.SendMessageResponse{Success: true, Message: fmt.Sprintf("%s message sent", msgType)}, nil
}
