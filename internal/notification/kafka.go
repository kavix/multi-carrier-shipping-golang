package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type NotificationMessage struct {
	Recipient string `json:"recipient"`
	Method    string `json:"method"` // "EMAIL" or "TELEGRAM"
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

type KafkaConsumer struct {
	reader  *kafka.Reader
	service NotificationService
}

func NewKafkaConsumer(brokers []string, topic, groupID string, service NotificationService) *KafkaConsumer {
	if len(brokers) == 0 {
		return nil
	}
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       10,
			MaxBytes:       10e6, // 10MB
			CommitInterval: time.Second,
			StartOffset:    kafka.LastOffset,
		}),
		service: service,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	if c == nil || c.reader == nil {
		slog.Warn("Kafka consumer is uninitialized or disabled")
		return
	}

	slog.Info("Starting Kafka Consumer background thread for shipment-notifications")

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				slog.Info("Kafka consumer loop stopped")
				return
			}
			slog.Error("Kafka consumer read error", slog.String("error", err.Error()))
			// Sleep a bit before retrying to prevent hot looping
			time.Sleep(2 * time.Second)
			continue
		}

		slog.Info("Kafka consumer received message event", 
			slog.Int64("offset", m.Offset),
			slog.String("key", string(m.Key)),
		)

		var msg NotificationMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			slog.Error("Failed to deserialize notification message payload", slog.String("error", err.Error()))
			continue
		}

		go func(msg NotificationMessage) {
			var err error
			switch msg.Method {
			case "EMAIL":
				_, err = c.service.SendEmailNotification(context.Background(), msg.Recipient, msg.Subject, msg.Body)
			case "TELEGRAM":
				_, err = c.service.SendTelegramNotification(context.Background(), msg.Recipient, msg.Body)
			default:
				slog.Error("Unknown notification method in event", slog.String("method", msg.Method))
				return
			}

			if err != nil {
				slog.Error("Failed to process event delivery via service layer",
					slog.String("method", msg.Method),
					slog.String("recipient", msg.Recipient),
					slog.String("error", err.Error()),
				)
			} else {
				slog.Info("Asynchronous message delivered and persisted successfully",
					slog.String("method", msg.Method),
					slog.String("recipient", msg.Recipient),
				)
			}
		}(msg)
	}
}

func (c *KafkaConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
