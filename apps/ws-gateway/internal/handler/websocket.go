package handler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/raghav1030/NaboServer/apps/ws-gateway/internal/middleware"
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
		userID, err := middleware.AuthenticateUser(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}
		defer func() {
			gateway.RemoveConnection(userID) // Now calls GatewayService's method
			conn.Close()
		}()

		gateway.AddConnection(userID, conn)    // Now calls GatewayService's method
		gateway.HandleConnection(userID, conn) // Updated method signature
	}
}
