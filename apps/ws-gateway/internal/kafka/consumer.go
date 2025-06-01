package kafka

import (
	"context"
	"log"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/raghav1030/NaboServer/apps/ws-gateway/internal/connection"
	chatv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/chat/v1"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	reader  *kafka.Reader
	connMgr *connection.Manager
}

func NewConsumer(brokers []string, groupID string, connMgr *connection.Manager) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			GroupID:     groupID,
			StartOffset: kafka.FirstOffset,
		}),
		connMgr: connMgr,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Kafka consumer error: %v\n", err)
				continue
			}

			switch {
			case strings.HasPrefix(msg.Topic, "user-"): // Unified message format
				c.handleUserMessage(msg.Value)
			default:
				log.Printf("Unknown topic: %s\n", msg.Topic)
			}
		}
	}()
}

func (c *Consumer) handleUserMessage(data []byte) {
	var envelope chatv1.MessageEnvelope
	if err := proto.Unmarshal(data, &envelope); err != nil {
		log.Printf("Failed to unmarshal message envelope: %v\n", err)
		return
	}

	// The envelope contains all required recipient information
	for _, recipient := range envelope.Recipients {
		if conn := c.connMgr.GetConnection(recipient); conn != nil {
			msgBytes, err := proto.Marshal(envelope.Message)
			if err != nil {
				log.Printf("Failed to marshal message for %s: %v\n", recipient, err)
				continue
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, msgBytes); err != nil {
				log.Printf("Failed to send message to %s: %v\n", recipient, err)
			}
		}
	}
}

func (c *Consumer) Close() {
	c.reader.Close()
}
