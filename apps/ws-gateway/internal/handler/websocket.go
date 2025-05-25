package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/raghav1030/NaboServer/apps/ws-gateway/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking in production!
		return true
	},
}

func WebSocketHandler(gateway *service.GatewayService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}
		defer conn.Close()

		// Create a channel to wait for connection closure
		done := make(chan struct{})
		
		// Start connection handling
		go gateway.HandleConnection(conn, done)
		
		// Block until connection is closed
		<-done
		log.Println("Connection closed")
	}
}
