package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/utils"
	"github.com/shipping/shipment-service/internal/domain"
	"github.com/shipping/shipment-service/internal/repository"
)

type ShipmentService struct {
	repo            *repository.ShipmentRepo
	createdProducer *kafka.Producer
	updatedProducer *kafka.Producer
	statusProducer  *kafka.Producer
	deletedProducer *kafka.Producer
}

func NewShipmentService(
	repo *repository.ShipmentRepo,
	createdProducer *kafka.Producer,
	updatedProducer *kafka.Producer,
	statusProducer *kafka.Producer,
	deletedProducer *kafka.Producer,
) *ShipmentService {
	return &ShipmentService{
		repo:            repo,
		createdProducer: createdProducer,
		updatedProducer: updatedProducer,
		statusProducer:  statusProducer,
		deletedProducer: deletedProducer,
	}
}

func (s *ShipmentService) CreateShipment(ctx context.Context, userID string, req *CreateShipmentRequest) (*domain.Shipment, error) {
	shipment := &domain.Shipment{
		ID:               utils.GenerateID(),
		UserID:           userID,
		SenderName:       req.SenderName,
		SenderAddress:    req.SenderAddress,
		SenderEmail:      req.SenderEmail,
		ReceiverName:     req.ReceiverName,
		ReceiverAddress:  req.ReceiverAddress,
		ReceiverEmail:    req.ReceiverEmail,
		Weight:           req.Weight,
		Dimensions:       req.Dimensions,
		Carrier:          req.Carrier,
		ServiceType:      req.ServiceType,
		PickupLocationID: req.PickupLocationID,
		DropLocationID:   req.DropLocationID,
		Status:           string(domain.StatusPending),
		TrackingNumber:   "",
		Cost:             0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.Create(ctx, shipment); err != nil {
		return nil, fmt.Errorf("create shipment: %w", err)
	}

	// Publish event to Kafka
	event := map[string]interface{}{
		"shipment_id":        shipment.ID,
		"user_id":            shipment.UserID,
		"carrier":            shipment.Carrier,
		"service_type":       shipment.ServiceType,
		"status":             shipment.Status,
		"sender_name":        shipment.SenderName,
		"sender":             shipment.SenderAddress,
		"receiver_name":      shipment.ReceiverName,
		"receiver":           shipment.ReceiverAddress,
		"sender_email":       shipment.SenderEmail,
		"receiver_email":     shipment.ReceiverEmail,
		"weight":             shipment.Weight,
		"pickup_location_id": shipment.PickupLocationID,
		"drop_location_id":   shipment.DropLocationID,
		"event_type":         "shipment.created",
	}
	if err := s.createdProducer.Publish(ctx, shipment.ID, event); err != nil {
		logger.Error("failed to publish shipment.created", logger.String("err", err.Error()))
	}

	logger.Info("shipment created", logger.String("id", shipment.ID))
	return shipment, nil
}

func (s *ShipmentService) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ShipmentService) ListUserShipments(ctx context.Context, userID string) ([]*domain.Shipment, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *ShipmentService) UpdateShipment(ctx context.Context, id string, req *UpdateShipmentRequest) (*domain.Shipment, error) {
	shipment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.SenderName != "" {
		shipment.SenderName = req.SenderName
	}
	if req.SenderAddress != "" {
		shipment.SenderAddress = req.SenderAddress
	}
	if req.SenderEmail != "" {
		shipment.SenderEmail = req.SenderEmail
	}
	if req.ReceiverName != "" {
		shipment.ReceiverName = req.ReceiverName
	}
	if req.ReceiverAddress != "" {
		shipment.ReceiverAddress = req.ReceiverAddress
	}
	if req.ReceiverEmail != "" {
		shipment.ReceiverEmail = req.ReceiverEmail
	}
	if req.Weight > 0 {
		shipment.Weight = req.Weight
	}
	if req.Dimensions != "" {
		shipment.Dimensions = req.Dimensions
	}
	if req.Carrier != "" {
		shipment.Carrier = req.Carrier
	}
	if req.ServiceType != "" {
		shipment.ServiceType = req.ServiceType
	}
	shipment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, shipment); err != nil {
		return nil, fmt.Errorf("update shipment: %w", err)
	}

	// Publish update event
	event := map[string]interface{}{
		"shipment_id":    shipment.ID,
		"user_id":        shipment.UserID,
		"status":         shipment.Status,
		"sender_email":   shipment.SenderEmail,
		"receiver_email": shipment.ReceiverEmail,
		"event_type":     "shipment.updated",
	}
	if err := s.updatedProducer.Publish(ctx, shipment.ID, event); err != nil {
		logger.Error("failed to publish shipment.updated", logger.String("err", err.Error()))
	}

	return shipment, nil
}

func (s *ShipmentService) UpdateStatus(ctx context.Context, id, status string) error {
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	shipment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to fetch shipment for status change event", logger.String("id", id), logger.String("err", err.Error()))
		return err
	}

	// Publish status change event
	event := map[string]interface{}{
		"shipment_id":    shipment.ID,
		"user_id":        shipment.UserID,
		"carrier":        shipment.Carrier,
		"status":         shipment.Status,
		"sender":         shipment.SenderAddress,
		"receiver":       shipment.ReceiverAddress,
		"sender_email":   shipment.SenderEmail,
		"receiver_email": shipment.ReceiverEmail,
		"event_type":     "shipment.status.changed",
		"timestamp":      time.Now(),
	}
	if err := s.statusProducer.Publish(ctx, id, event); err != nil {
		logger.Error("failed to publish status change", logger.String("err", err.Error()))
	}

	logger.Info("shipment status updated", logger.String("id", id), logger.String("status", status))
	return nil
}

func (s *ShipmentService) DeleteShipment(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish delete event
	event := map[string]interface{}{
		"shipment_id": id,
		"event_type":  "shipment.deleted",
	}
	if err := s.deletedProducer.Publish(ctx, id, event); err != nil {
		logger.Error("failed to publish shipment.deleted", logger.String("err", err.Error()))
	}

	return nil
}

type CreateShipmentRequest struct {
	SenderName       string  `json:"sender_name" binding:"required"`
	SenderAddress    string  `json:"sender_address" binding:"required"`
	SenderEmail      string  `json:"sender_email"`
	ReceiverName     string  `json:"receiver_name" binding:"required"`
	ReceiverAddress  string  `json:"receiver_address" binding:"required"`
	ReceiverEmail    string  `json:"receiver_email"`
	Weight           float64 `json:"weight" binding:"required,gt=0"`
	Dimensions       string  `json:"dimensions"`
	Carrier          string  `json:"carrier" binding:"required"`
	ServiceType      string  `json:"service_type" binding:"required"`
	PickupLocationID string  `json:"pickup_location_id"`
	DropLocationID   string  `json:"drop_location_id"`
}

type UpdateShipmentRequest struct {
	SenderName      string  `json:"sender_name"`
	SenderAddress   string  `json:"sender_address"`
	SenderEmail     string  `json:"sender_email"`
	ReceiverName    string  `json:"receiver_name"`
	ReceiverAddress string  `json:"receiver_address"`
	ReceiverEmail   string  `json:"receiver_email"`
	Weight          float64 `json:"weight"`
	Dimensions      string  `json:"dimensions"`
	Carrier         string  `json:"carrier"`
	ServiceType     string  `json:"service_type"`
}
