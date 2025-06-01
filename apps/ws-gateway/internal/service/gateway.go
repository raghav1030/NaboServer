package service

import (
	"context"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/gorilla/websocket"
	chatv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/chat/v1"
	gatewayv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/gateway/v1"
	locationv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/location/v1"
	"github.com/segmentio/kafka-go"
)

const (
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	writeTimeout   = 10 * time.Second
	maxMessageSize = 512
)

type ConnectionManager struct {
	connections map[string]*websocket.Conn // userID → WebSocket connection
	mu          sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*websocket.Conn),
	}
}

func (cm *ConnectionManager) Add(userID string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.connections[userID] = conn
}

func (cm *ConnectionManager) Remove(userID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.connections, userID)
}

func (cm *ConnectionManager) Get(userID string) *websocket.Conn {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.connections[userID]
}

func (cm *ConnectionManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.connections)
}

func (cm *ConnectionManager) Broadcast(msg []byte) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for userID, conn := range cm.connections {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("Broadcast failed to %s: %v", userID, err)
		}
	}
}

type GatewayService struct {
	connManager    *ConnectionManager
	locationClient locationv1.LocationServiceClient
	grpcConn       *grpc.ClientConn
	kafkaReader    *kafka.Reader
}

func NewGatewayService(locationServiceAddr string, kafkaBrokers []string) (*GatewayService, error) {
	conn, err := grpc.Dial(locationServiceAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &GatewayService{
		connManager:    NewConnectionManager(),
		locationClient: locationv1.NewLocationServiceClient(conn),
		grpcConn:       conn,
		kafkaReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: kafkaBrokers,
			GroupID: "ws-gateway-group",
		}),
	}, nil
}

func (gs *GatewayService) AddConnection(userID string, conn *websocket.Conn) {
	gs.connManager.Add(userID, conn)
}

func (gs *GatewayService) RemoveConnection(userID string) {
	gs.connManager.Remove(userID)
}

func (gs *GatewayService) StartKafkaConsumer() {
	go func() {
		for {
			msg, err := gs.kafkaReader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Kafka consumer error: %v", err)
				continue
			}
			gs.handleKafkaMessage(msg.Value)
		}
	}()
}

func (gs *GatewayService) handleKafkaMessage(data []byte) {
	var envelope chatv1.MessageEnvelope
	if err := proto.Unmarshal(data, &envelope); err != nil {
		log.Printf("Failed to unmarshal message envelope: %v", err)
		return
	}

	for _, recipient := range envelope.Recipients {
		if conn := gs.connManager.Get(recipient); conn != nil {
			msgBytes, err := proto.Marshal(envelope.Message)

			if err != nil {
				log.Printf("Failed to parse message %v: %v", envelope.Message, err)
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, msgBytes); err != nil {
				log.Printf("Failed to send message to %s: %v", recipient, err)
			}
		}
	}
}

func (gs *GatewayService) HandleConnection(userID string, conn *websocket.Conn) {
	gs.connManager.Add(userID, conn)
	defer gs.connManager.Remove(userID)
	defer conn.Close()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Use separate channels for control and data
	done := make(chan struct{})
	sendChan := make(chan []byte, 100)

	go gs.readPump(conn, done, sendChan)
	gs.writePump(conn, done, sendChan)
}

func (gs *GatewayService) readPump(conn *websocket.Conn, done chan struct{}, sendChan chan<- []byte) {
	defer close(done)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("Read error: %v", err)
			}
			return
		}
		go gs.handleIncomingMessage(message, sendChan)
	}
}

func (gs *GatewayService) writePump(conn *websocket.Conn, done <-chan struct{}, sendChan <-chan []byte) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case msg := <-sendChan:
			conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error: %v", err)
				return
			}
		}
	}
}

func (gs *GatewayService) handleIncomingMessage(msg []byte, sendChan chan<- []byte) {
	var frame gatewayv1.StreamRequest
	if err := proto.Unmarshal(msg, &frame); err != nil {
		log.Printf("Failed to decode protobuf: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response *gatewayv1.StreamResponse
	var err error

	switch payload := frame.Payload.(type) {
	case *gatewayv1.StreamRequest_LocationUpdate:
		response, err = gs.handleLocationUpdate(ctx, &frame, payload)
	case *gatewayv1.StreamRequest_NearbyRequest:
		response, err = gs.handleNearbyRequest(ctx, &frame, payload)
	default:
		log.Printf("Unhandled message type: %T", payload)
		return
	}

	if err != nil {
		log.Printf("Message handling error: %v", err)
		return
	}

	if response != nil {
		data, err := proto.Marshal(response)
		if err != nil {
			log.Printf("Failed to marshal response: %v", err)
			return
		}
		sendChan <- data
	}
}

func (gs *GatewayService) handleLocationUpdate(ctx context.Context, frame *gatewayv1.StreamRequest, payload *gatewayv1.StreamRequest_LocationUpdate) (*gatewayv1.StreamResponse, error) {
	resp, err := gs.locationClient.UpdateLocation(ctx, payload.LocationUpdate)
	if err != nil {
		return nil, err
	}
	return &gatewayv1.StreamResponse{
		SessionId: frame.SessionId,
		Payload:   &gatewayv1.StreamResponse_LocationUpdate{LocationUpdate: resp},
	}, nil
}

func (gs *GatewayService) handleNearbyRequest(ctx context.Context, frame *gatewayv1.StreamRequest, payload *gatewayv1.StreamRequest_NearbyRequest) (*gatewayv1.StreamResponse, error) {
	resp, err := gs.locationClient.GetNearbyUsers(ctx, payload.NearbyRequest)
	if err != nil {
		return nil, err
	}
	return &gatewayv1.StreamResponse{
		SessionId: frame.SessionId,
		Payload:   &gatewayv1.StreamResponse_NearbyResponse{NearbyResponse: resp},
	}, nil
}

func (gs *GatewayService) Close() {
	gs.grpcConn.Close()
	if err := gs.kafkaReader.Close(); err != nil {
		log.Printf("Error closing Kafka reader: %v", err)
	}
}
