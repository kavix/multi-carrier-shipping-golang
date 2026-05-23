package label

import (
	"context"
	"errors"
	"testing"
)

type mockLabelRepository struct {
	labels map[string]*Label
}

func newMockLabelRepository() *mockLabelRepository {
	return &mockLabelRepository{
		labels: make(map[string]*Label),
	}
}

func (m *mockLabelRepository) Create(ctx context.Context, l *Label) error {
	if _, exists := m.labels[l.ID]; exists {
		return ErrLabelAlreadyCancelled // or duplicate error
	}
	m.labels[l.ID] = l
	return nil
}

func (m *mockLabelRepository) GetByID(ctx context.Context, id string) (*Label, error) {
	l, exists := m.labels[id]
	if !exists {
		return nil, ErrLabelNotFound
	}
	return l, nil
}

func (m *mockLabelRepository) GetByTracking(ctx context.Context, trackingNum string) (*Label, error) {
	for _, l := range m.labels {
		if l.TrackingNumber == trackingNum {
			return l, nil
		}
	}
	return nil, ErrLabelNotFound
}

func (m *mockLabelRepository) Update(ctx context.Context, l *Label) error {
	if _, exists := m.labels[l.ID]; !exists {
		return ErrLabelNotFound
	}
	m.labels[l.ID] = l
	return nil
}

func (m *mockLabelRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.labels[id]; !exists {
		return ErrLabelNotFound
	}
	delete(m.labels, id)
	return nil
}

func TestCreateLabel(t *testing.T) {
	repo := newMockLabelRepository()
	svc := NewLabelService(repo, nil, "", "")
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		resp, err := svc.CreateLabel(ctx, "shipment-123", "DHL", 5.5, "London", "Paris")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		label := resp.Label
		if label.ID == "" {
			t.Errorf("expected generated ID to be populated, got empty string")
		}
		if label.ShipmentID != "shipment-123" {
			t.Errorf("expected shipmentID 'shipment-123', got %s", label.ShipmentID)
		}
		if label.TrackingNumber == "" {
			t.Errorf("expected tracking number to be auto-assigned")
		}
		if label.Status != "ACTIVE" {
			t.Errorf("expected label status to be 'ACTIVE', got %s", label.Status)
		}
	})

	t.Run("missing carrier validation error", func(t *testing.T) {
		_, err := svc.CreateLabel(ctx, "shipment-123", "", 2.0, "Origin", "Destination")
		if !errors.Is(err, ErrCarrierRequired) {
			t.Errorf("expected ErrCarrierRequired, got %v", err)
		}
	})

	t.Run("invalid weight validation error", func(t *testing.T) {
		_, err := svc.CreateLabel(ctx, "shipment-123", "FedEx", -0.5, "Origin", "Destination")
		if !errors.Is(err, ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight, got %v", err)
		}
	})
}
