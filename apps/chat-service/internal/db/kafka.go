package db

import (
	"context"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
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

func (k *KafkaManager) Publish(ctx context.Context, topic string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return k.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: data,
	})
}
