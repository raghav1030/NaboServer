package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/raghav1030/NaboServer/apps/ws-gateway/internal/handler"
	"github.com/raghav1030/NaboServer/apps/ws-gateway/internal/service"
)

func main() {
	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	gateway, err := service.NewGatewayService(
		"localhost:50051", // Location service address
		kafkaBrokers,
	)
	if err != nil {
		log.Fatalf("Failed to initialize gateway: %v", err)
	}
	defer gateway.Close()

	// Start Kafka consumer
	gateway.StartKafkaConsumer()

	http.HandleFunc("/ws", handler.WebSocketHandler(gateway))
	log.Println("WebSocket Gateway listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
