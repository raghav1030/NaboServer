package main

import (
	"log"
	"net/http"

	"github.com/raghav1030/NaboServer/apps/ws-gateway/internal/router"
	"github.com/raghav1030/NaboServer/apps/ws-gateway/internal/service"
)

func main() {
	
	gateway, err := service.NewGatewayService("localhost:50051") 
	if err != nil {
		log.Fatalf("Failed to initialize gateway: %v", err)
	}
	defer gateway.Close()

	// Set up HTTP server
	mux := router.SetupRouter(gateway)
	
	log.Println("WebSocket Gateway listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
