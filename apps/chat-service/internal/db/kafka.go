package db

import (
	"context"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	chatv1 "github.com/raghav1030/NaboServer/libs/proto/gen/go/chat/v1"
)

type KafkaManager struct {
	writer *kafka.Writer
}

func NewKafkaManager(brokers []string) *KafkaManager {
	return &KafkaManager{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (k *KafkaManager) MarshalMessage(msg *chatv1.SendMessageRequest) ([]byte, error) {
	return proto.Marshal(msg)
}

func (k *KafkaManager) PublishMessage(ctx context.Context, topic string, data []byte) error {
	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: data,
	})
}

// For completeness, you can add a consumer for delivering messages to clients via gateway
