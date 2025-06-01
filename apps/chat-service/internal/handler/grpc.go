package handler

import (
	"context"
	"log"

	"github.com/raghav1030/NaboServer/apps/chat-service/internal/service"
	chatv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/chat/v1"
)

type ChatServer struct {
	chatv1.UnimplementedChatServiceServer
	chatService *service.ChatService
}

func NewChatServer(chatService *service.ChatService) *ChatServer {
	return &ChatServer{chatService: chatService}
}

func (s *ChatServer) SendMessage(ctx context.Context, req *chatv1.SendMessageRequest) (*chatv1.SendMessageResponse, error) {
	log.Println("Received SendMessage")
	return s.chatService.HandleSendMessage(ctx, req)
}
