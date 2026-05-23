package notification

import (
	"encoding/json"
	"testing"
)

func TestKafkaNotificationSchema(t *testing.T) {
	t.Run("JSON serialization and deserialization matching", func(t *testing.T) {
		original := NotificationMessage{
			Recipient: "kavix@yahoo.com",
			Method:    "EMAIL",
			Subject:   "Shipment Generated",
			Body:      "<h1>Test Body</h1>",
		}

		bytes, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("failed to marshal notification: %v", err)
		}

		var deserialized NotificationMessage
		if err := json.Unmarshal(bytes, &deserialized); err != nil {
			t.Fatalf("failed to unmarshal notification: %v", err)
		}

		if deserialized.Recipient != original.Recipient {
			t.Errorf("expected recipient '%s', got '%s'", original.Recipient, deserialized.Recipient)
		}
		if deserialized.Method != original.Method {
			t.Errorf("expected method '%s', got '%s'", original.Method, deserialized.Method)
		}
		if deserialized.Subject != original.Subject {
			t.Errorf("expected subject '%s', got '%s'", original.Subject, deserialized.Subject)
		}
		if deserialized.Body != original.Body {
			t.Errorf("expected body '%s', got '%s'", original.Body, deserialized.Body)
		}
	})

	t.Run("empty brokers safety check", func(t *testing.T) {
		consumer := NewKafkaConsumer(nil, "topic", "group", nil)
		if consumer != nil {
			t.Error("expected consumer to be nil when brokers list is empty")
		}
	})
}
