package service

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/gorilla/websocket"
	gatewayv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/gateway/v1"
	locationv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/location/v1"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type GatewayService struct {
	connections     map[*websocket.Conn]struct{}
	connectionsLock sync.RWMutex
	locationClient  locationv1.LocationServiceClient
	grpcConn        *grpc.ClientConn // Shared gRPC connection
}

func NewGatewayService(locationServiceAddr string) (*GatewayService, error) {
	// Set up gRPC connection
	conn, err := grpc.Dial(locationServiceAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &GatewayService{
		connections:    make(map[*websocket.Conn]struct{}),
		locationClient: locationv1.NewLocationServiceClient(conn),
		grpcConn:       conn,
	}, nil
}

func (g *GatewayService) HandleConnection(conn *websocket.Conn, done chan struct{}) {
	g.connectionsLock.Lock()
	g.connections[conn] = struct{}{}
	g.connectionsLock.Unlock()

	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	sendChan := make(chan []byte, 100)
	defer close(sendChan)

	// Read pump
	go func() {
		defer func() {
			g.cleanupConnection(conn)
			close(done)
		}()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Printf("Read error: %v", err)
				}
				return
			}
			go g.handleIncomingMessage(message, sendChan)
		}
	}()

	// Write pump
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case msg := <-sendChan:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error: %v", err)
				return
			}
		}
	}
}

func (g *GatewayService) cleanupConnection(conn *websocket.Conn) {
	g.connectionsLock.Lock()
	delete(g.connections, conn)
	g.connectionsLock.Unlock()
	conn.Close()
}

func (g *GatewayService) Close() {
	g.grpcConn.Close()
}

func (g *GatewayService) Broadcast(msg []byte) {
	g.connectionsLock.RLock()
	defer g.connectionsLock.RUnlock()

	for conn := range g.connections {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("Broadcast failed to %v: %v", conn.RemoteAddr(), err)
		}
	}
}

func (g *GatewayService) ConnectionCount() int {
	g.connectionsLock.RLock()
	defer g.connectionsLock.RUnlock()
	return len(g.connections)
}
func (g *GatewayService) handleIncomingMessage(msg []byte, sendChan chan<- []byte) {
	var frame gatewayv1.StreamRequest
	if err := proto.Unmarshal(msg, &frame); err != nil {
		log.Printf("Failed to decode protobuf: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch payload := frame.Payload.(type) {
	case *gatewayv1.StreamRequest_LocationUpdate:
		resp, err := g.locationClient.UpdateLocation(ctx, payload.LocationUpdate)
		if err != nil {
			log.Printf("Location update failed: %v", err)
			return
		}
		response := &gatewayv1.StreamResponse{
			SessionId: frame.SessionId,
			Payload:   &gatewayv1.StreamResponse_LocationUpdate{LocationUpdate: resp},
		}
		data, err := proto.Marshal(response)
		if err == nil {
			sendChan <- data
		}

	case *gatewayv1.StreamRequest_NearbyRequest:
		resp, err := g.locationClient.GetNearbyUsers(ctx, payload.NearbyRequest)
		if err != nil {
			log.Printf("Nearby users query failed: %v", err)
			return
		}
		response := &gatewayv1.StreamResponse{
			SessionId: frame.SessionId,
			Payload:   &gatewayv1.StreamResponse_NearbyResponse{NearbyResponse: resp},
		}
		data, err := proto.Marshal(response)
		if err == nil {
			sendChan <- data
		}

	default:
		log.Printf("Unhandled message type: %T", payload)
	}
}
