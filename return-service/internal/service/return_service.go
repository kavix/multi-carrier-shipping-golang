package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/shipping/return-service/internal/domain"
	"github.com/shipping/return-service/internal/repository"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/utils"
)

type ReturnService struct {
	repo     *repository.ReturnRepo
	producer *kafka.Producer
}

func NewReturnService(repo *repository.ReturnRepo, producer *kafka.Producer) *ReturnService {
	return &ReturnService{repo: repo, producer: producer}
}

func (s *ReturnService) RequestReturn(ctx context.Context, userID, shipmentID, reason string) (*domain.ReturnRequest, error) {
	ret := &domain.ReturnRequest{
		ID:            utils.GenerateID(),
		ShipmentID:    shipmentID,
		UserID:        userID,
		Reason:        reason,
		Status:        string(domain.ReturnRequested),
		Carrier:       "",
		ReturnLabelID: "",
		RefundAmount:  0,
		RefundStatus:  "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, ret); err != nil {
		return nil, fmt.Errorf("create return: %w", err)
	}

	// Publish event
	event := map[string]interface{}{
		"return_id":   ret.ID,
		"shipment_id": shipmentID,
		"user_id":     userID,
		"reason":      reason,
		"status":      ret.Status,
		"event_type":  "return.created",
	}
	if err := s.producer.Publish(ctx, ret.ID, event); err != nil {
		logger.Error("failed to publish return.created", logger.String("err", err.Error()))
	}

	return ret, nil
}

func (s *ReturnService) ApproveReturn(ctx context.Context, returnID, carrier string) (*domain.ReturnRequest, error) {
	ret, err := s.repo.GetByID(ctx, returnID)
	if err != nil {
		return nil, err
	}

	if ret.Status != string(domain.ReturnRequested) {
		return nil, fmt.Errorf("return cannot be approved, current status: %s", ret.Status)
	}

	// Generate return label via label generation service
	labelServiceURL := getEnv("LABEL_SERVICE_URL", "http://label-generation-service:8084")
	labelReq := map[string]interface{}{
		"shipment_id": ret.ShipmentID,
		"carrier":     carrier,
	}
	labelJSON, _ := json.Marshal(labelReq)
	resp, err := http.Post(labelServiceURL+"/labels", "application/json", bytes.NewReader(labelJSON))
	if err != nil {
		logger.Error("failed to generate return label", logger.String("err", err.Error()))
	} else {
		defer resp.Body.Close()
		var labelResp struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&labelResp); err == nil {
			ret.ReturnLabelID = labelResp.ID
		}
	}

	ret.Carrier = carrier
	ret.Status = string(domain.ReturnApproved)
	ret.UpdatedAt = time.Now()

	if err := s.repo.UpdateStatus(ctx, returnID, ret.Status); err != nil {
		return nil, err
	}

	// Publish event
	event := map[string]interface{}{
		"return_id":   ret.ID,
		"shipment_id": ret.ShipmentID,
		"status":      ret.Status,
		"carrier":     carrier,
		"event_type":  "return.status.changed",
	}
	if err := s.producer.Publish(ctx, ret.ID, event); err != nil {
		logger.Error("failed to publish return.status.changed", logger.String("err", err.Error()))
	}

	return ret, nil
}

func (s *ReturnService) ProcessRefund(ctx context.Context, returnID string, amount float64) error {
	ret, err := s.repo.GetByID(ctx, returnID)
	if err != nil {
		return err
	}

	if ret.Status != string(domain.ReturnReceived) {
		return fmt.Errorf("return not received yet, status: %s", ret.Status)
	}

	if err := s.repo.UpdateRefund(ctx, returnID, amount, "refunded"); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, returnID, string(domain.ReturnRefunded)); err != nil {
		return err
	}

	// Publish refund event
	event := map[string]interface{}{
		"return_id":     returnID,
		"shipment_id":   ret.ShipmentID,
		"user_id":       ret.UserID,
		"refund_amount": amount,
		"status":        "refunded",
		"event_type":    "return.status.changed",
	}
	if err := s.producer.Publish(ctx, returnID, event); err != nil {
		logger.Error("failed to publish refund", logger.String("err", err.Error()))
	}

	return nil
}

func (s *ReturnService) GetReturn(ctx context.Context, id string) (*domain.ReturnRequest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ReturnService) ListReturns(ctx context.Context, shipmentID string) ([]*domain.ReturnRequest, error) {
	return s.repo.GetByShipmentID(ctx, shipmentID)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
