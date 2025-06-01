package service

import (
	"context"
	"fmt"

	"github.com/raghav1030/NaboServer/apps/chat-service/internal/db"
	chatv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/chat/v1"
)

type ChatService struct {
	pg    *db.PostgresManager
	kafka *db.KafkaManager
}

func NewChatService(pg *db.PostgresManager, kafka *db.KafkaManager) *ChatService {
	return &ChatService{pg: pg, kafka: kafka}
}

func (s *ChatService) HandleSendMessage(ctx context.Context, req *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	var conversationID, senderID, text, msgType string
	var attachment *chatv1.FileAttachment
	var topic string

	switch m := req.Message.(type) {
	case *chatv1.SendMessageRequest_Direct:
		conversationID = fmt.Sprintf("dm-%s-%s", m.Direct.FromUserId, m.Direct.ToUserId)
		senderID = m.Direct.FromUserId
		text = m.Direct.Text
		msgType = "direct"
		attachment = m.Direct.Attachment
		topic = fmt.Sprintf("dm-%s", m.Direct.ToUserId)
	case *chatv1.SendMessageRequest_Group:
		conversationID = fmt.Sprintf("group-%s", m.Group.GroupId)
		senderID = m.Group.FromUserId
		text = m.Group.Text
		msgType = "group"
		attachment = m.Group.Attachment
		topic = fmt.Sprintf("group-%s", m.Group.GroupId)
	case *chatv1.SendMessageRequest_Ephemeral:
		conversationID = fmt.Sprintf("activity-%s", m.Ephemeral.ActivityId)
		senderID = m.Ephemeral.FromUserId
		text = m.Ephemeral.Text
		msgType = "ephemeral"
		attachment = m.Ephemeral.Attachment
		topic = fmt.Sprintf("activity-%s", m.Ephemeral.ActivityId)
	default:
		return &chatv1.SendMessageResponse{Success: false, Message: "Unknown message type"}, nil
	}

	_, err := s.pg.SaveMessage(ctx, conversationID, senderID, text, msgType, attachment)
	if err != nil {
		return &chatv1.SendMessageResponse{Success: false, Message: "DB error"}, err
	}

	if err := s.kafka.Publish(ctx, topic, req); err != nil {
		return &chatv1.SendMessageResponse{Success: false, Message: "Kafka error"}, err
	}

	return &chatv1.SendMessageResponse{Success: true, Message: "Message sent"}, nil
}
