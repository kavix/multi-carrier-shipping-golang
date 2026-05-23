package shipment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type NotificationMessage struct {
	Recipient string `json:"recipient"`
	Method    string `json:"method"` // "EMAIL" or "TELEGRAM"
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	if len(brokers) == 0 {
		return nil
	}
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 3 * time.Second,
			DialTimeout:  3 * time.Second,
		},
	}
}

func (p *KafkaPublisher) PublishNotification(ctx context.Context, method, recipient, subject, body string) error {
	if p == nil || p.writer == nil {
		return fmt.Errorf("kafka publisher is uninitialized or disabled")
	}

	msg := NotificationMessage{
		Recipient: recipient,
		Method:    method,
		Subject:   subject,
		Body:      body,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal notification message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: bytes,
	})
	if err != nil {
		return fmt.Errorf("failed to write kafka message: %w", err)
	}

	return nil
}

func (p *KafkaPublisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
