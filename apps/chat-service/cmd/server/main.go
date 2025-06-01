package main

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/raghav1030/NaboServer/apps/chat-service/internal/db"
	"github.com/raghav1030/NaboServer/apps/chat-service/internal/handler"
	"github.com/raghav1030/NaboServer/apps/chat-service/internal/service"
	chatv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/chat/v1"
	"google.golang.org/grpc"
)

func main() {
	pg, err := db.NewPostgresManager(os.Getenv("POSTGRES_DSN"))
	if err != nil {
		log.Fatal("Failed to connect to Postgres:", err)
	}
	kafka := db.NewKafkaManager(strings.Split(os.Getenv("KAFKA_BROKERS"), ","))

	chatService := service.NewChatService(pg, kafka)
	chatServer := handler.NewChatServer(chatService)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	chatv1.RegisterChatServiceServer(grpcServer, chatServer)

	log.Println("ChatService gRPC server listening on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
