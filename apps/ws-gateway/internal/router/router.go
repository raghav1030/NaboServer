package router

import (
	"net/http"

	"github.com/raghav1030/NaboServer/apps/ws-gateway/internal/handler"
	"github.com/raghav1030/NaboServer/apps/ws-gateway/internal/service"
)

func SetupRouter(gateway *service.GatewayService) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handler.WebSocketHandler(gateway))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	return mux
}
