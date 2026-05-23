package shipment

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type shipmentService struct {
	repo                   ShipmentRepository
	labelServiceURL        string
	authServiceURL         string
	notificationServiceURL string
	kafkaPublisher         *KafkaPublisher
	httpClient             *http.Client

	// Concurrent rate limiter tracking (5-second throttle per user)
	mu          sync.Mutex
	lastCreated map[string]time.Time
	rateLimit   time.Duration // Configurable duration for robust unit testing
}

// NewShipmentService instantiates a new ShipmentService implementation.
func NewShipmentService(repo ShipmentRepository, labelServiceURL string, authServiceURL string, notificationServiceURL string, kafkaBrokers []string) ShipmentService {
	var publisher *KafkaPublisher
	if len(kafkaBrokers) > 0 {
		slog.Info("Initializing Kafka Publisher for shipment-notifications", slog.Any("brokers", kafkaBrokers))
		publisher = NewKafkaPublisher(kafkaBrokers, "shipment-notifications")
	}
	return &shipmentService{
		repo:                   repo,
		labelServiceURL:        strings.TrimSuffix(labelServiceURL, "/"),
		authServiceURL:         strings.TrimSuffix(authServiceURL, "/"),
		notificationServiceURL: strings.TrimSuffix(notificationServiceURL, "/"),
		kafkaPublisher:         publisher,
		httpClient:             &http.Client{Timeout: 10 * time.Second},
		lastCreated:            make(map[string]time.Time),
		rateLimit:              5 * time.Second, // Default to 5s production limit
	}
}

// Helper to verify auth token with Auth Microservice
func (s *shipmentService) verifyToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", errors.New("unauthorized: missing token")
	}

	verifyURL := fmt.Sprintf("%s/api/v1/auth/verify?token=%s", s.authServiceURL, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, verifyURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create verify request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth service is currently unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("unauthorized: invalid or expired session token")
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth verification failed with status %d", resp.StatusCode)
	}

	var data struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to decode verify response: %w", err)
	}

	return data.Username, nil
}

// Helper to log user actions with Auth Microservice
func (s *shipmentService) logAction(ctx context.Context, username, action string) {
	if username == "" || s.authServiceURL == "" {
		return
	}

	payload := map[string]string{
		"username": username,
		"action":   action,
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	logURL := fmt.Sprintf("%s/api/v1/auth/logs", s.authServiceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

func (s *shipmentService) CreateShipment(
	ctx context.Context,
	token, carrier string,
	weight float64,
	origin, destination, email string,
) (*Shipment, *Label, error) {
	// 1. Authenticate & fetch username
	username, err := s.verifyToken(ctx, token)
	if err != nil {
		return nil, nil, err
	}

	// Concurrent-safe rate limit check per user
	s.mu.Lock()
	lastTime, exists := s.lastCreated[username]
	if exists && time.Since(lastTime) < s.rateLimit {
		s.mu.Unlock()
		return nil, nil, ErrRateLimitExceeded
	}
	s.lastCreated[username] = time.Now()
	s.mu.Unlock()

	// 2. Validation
	if carrier == "" {
		return nil, nil, ErrCarrierRequired
	}
	if weight <= 0 {
		return nil, nil, ErrInvalidWeight
	}

	// 3. Construct Shipment
	id, err := generateUUID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate shipment id: %w", err)
	}

	now := time.Now()
	shipment := &Shipment{
		ID:             id,
		Carrier:        carrier,
		TrackingNumber: "PENDING",
		Weight:         weight,
		Origin:         origin,
		Destination:    destination,
		Status:         "PENDING",
		Username:       username, // Record owner
		Email:          email,    // Persist recipient email
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, shipment); err != nil {
		return nil, nil, err
	}

	// 4. Request Label from Label Service
	labelReqBody := struct {
		ShipmentID  string  `json:"shipment_id"`
		Carrier     string  `json:"carrier"`
		Weight      float64 `json:"weight"`
		Origin      string  `json:"origin"`
		Destination string  `json:"destination"`
	}{
		ShipmentID:  shipment.ID,
		Carrier:     shipment.Carrier,
		Weight:      shipment.Weight,
		Origin:      shipment.Origin,
		Destination: shipment.Destination,
	}

	jsonBytes, err := json.Marshal(labelReqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal label request: %w", err)
	}

	labelURL := fmt.Sprintf("%s/api/v1/labels", s.labelServiceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, labelURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create label request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("label service is currently unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, nil, fmt.Errorf("label service returned error status %d", resp.StatusCode)
	}

	var label Label
	if err := json.NewDecoder(resp.Body).Decode(&label); err != nil {
		return nil, nil, fmt.Errorf("failed to decode label details: %w", err)
	}

	// 5. Update Shipment to CREATED
	shipment.TrackingNumber = label.TrackingNumber
	shipment.Status = "CREATED"
	shipment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, shipment); err != nil {
		return nil, nil, err
	}

	// 6. Send simulated HTML email notification
	if err := s.sendNotificationEmail(ctx, shipment, &label); err != nil {
		slog.Error("Failed to send shipment email notification", slog.String("error", err.Error()), slog.String("shipment_id", shipment.ID))
	}

	s.logAction(ctx, username, fmt.Sprintf("Create Shipment (ID: %s)", shipment.ID))

	return shipment, &label, nil
}

func (s *shipmentService) sendNotificationEmail(ctx context.Context, shipment *Shipment, label *Label) error {
	const emailTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Shipment Notification - TRK# {{.Label.TrackingNumber}}</title>
  <style>
    body {
      background-color: #0b0f19;
      color: #f1f5f9;
      font-family: 'Outfit', 'Inter', -apple-system, sans-serif;
      margin: 0;
      padding: 40px 20px;
    }
    .email-container {
      background: #1e293b;
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 20px;
      max-width: 600px;
      margin: 0 auto;
      padding: 35px;
      box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 10px 10px -5px rgba(0, 0, 0, 0.3);
    }
    .header {
      border-bottom: 1px solid rgba(255, 255, 255, 0.08);
      padding-bottom: 25px;
      margin-bottom: 30px;
      text-align: center;
    }
    .logo {
      font-size: 20px;
      font-weight: 800;
      letter-spacing: 0.15em;
      color: #6366f1;
    }
    .title {
      font-size: 24px;
      font-weight: 700;
      color: #ffffff;
      margin-top: 15px;
      margin-bottom: 5px;
    }
    .subtitle {
      font-size: 14px;
      color: #94a3b8;
      margin: 0;
    }
    .grid-2 {
      display: table;
      width: 100%;
      table-layout: fixed;
      margin-bottom: 20px;
      border-collapse: separate;
      border-spacing: 15px 0;
    }
    .grid-col {
      display: table-cell;
      width: 50%;
      vertical-align: top;
    }
    .card {
      background: #0f172a;
      border: 1px solid rgba(255, 255, 255, 0.05);
      border-radius: 12px;
      padding: 18px;
    }
    .card-title {
      font-size: 10px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #818cf8;
      margin-bottom: 8px;
    }
    .card-value {
      font-size: 14px;
      color: #e2e8f0;
      line-height: 1.4;
    }
    .label-box {
      background: #ffffff;
      color: #0f172a;
      border-radius: 16px;
      padding: 25px;
      margin-top: 30px;
      margin-bottom: 30px;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
      border: 1px solid #e2e8f0;
    }
    .label-header {
      border-bottom: 2px solid #0f172a;
      padding-bottom: 12px;
      margin-bottom: 15px;
      font-weight: 800;
      font-size: 14px;
      letter-spacing: 0.05em;
    }
    .label-address-grid {
      display: table;
      width: 100%;
      margin-bottom: 20px;
    }
    .barcode-simulated {
      height: 60px;
      background: repeating-linear-gradient(
        90deg,
        #0f172a,
        #0f172a 3px,
        #ffffff 3px,
        #ffffff 7px,
        #0f172a 7px,
        #0f172a 9px,
        #ffffff 9px,
        #ffffff 11px
      );
      margin-top: 20px;
      margin-bottom: 8px;
      border-radius: 2px;
    }
    .tracking-number {
      font-family: monospace;
      font-size: 14px;
      font-weight: 700;
      text-align: center;
      letter-spacing: 0.08em;
    }
    .footer {
      text-align: center;
      font-size: 11px;
      color: #64748b;
      border-top: 1px solid rgba(255, 255, 255, 0.08);
      padding-top: 25px;
      margin-top: 35px;
    }
    .btn {
      display: inline-block;
      background: linear-gradient(135deg, #6366f1, #8b5cf6);
      color: #ffffff !important;
      text-decoration: none;
      font-weight: 600;
      font-size: 14px;
      padding: 12px 30px;
      border-radius: 8px;
      text-align: center;
      margin-top: 10px;
    }
  </style>
</head>
<body>
  <div class="email-container">
    <div class="header">
      <div class="logo">ANTIGRAVITY SYSTEMS</div>
      <h1 class="title">Your Shipment has Dispatched</h1>
      <p class="subtitle">Notification details and printable carrier shipping label</p>
    </div>

    <div class="grid-2">
      <div class="grid-col">
        <div class="card">
          <div class="card-title">SHIPPER (ORIGIN)</div>
          <div class="card-value">{{.Shipment.Origin}}</div>
        </div>
      </div>
      <div class="grid-col">
        <div class="card">
          <div class="card-title">RECIPIENT (DESTINATION)</div>
          <div class="card-value">{{.Shipment.Destination}}</div>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <div class="grid-col">
        <div class="card">
          <div class="card-title">CARRIER & WEIGHT</div>
          <div class="card-value"><strong>{{.Shipment.Carrier}}</strong> ({{.Shipment.Weight}} lbs)</div>
        </div>
      </div>
      <div class="grid-col">
        <div class="card">
          <div class="card-title">DELIVERY STATUS</div>
          <div class="card-value" style="color: #10b981; font-weight: 700;">{{.Shipment.Status}}</div>
        </div>
      </div>
    </div>

    <div class="label-box">
      <div class="label-header">
        <span style="font-weight: 800; font-size: 14px;">{{uppercase .Shipment.Carrier}} PRIORITY MAIL</span>
      </div>
      <div class="label-address-grid">
        <div style="display: table-cell; width: 50%; font-size: 11px; padding-right: 10px; vertical-align: top;">
          <strong style="display:block; margin-bottom: 4px; font-size: 10px; color: #64748b;">FROM:</strong>
          {{.Shipment.Origin}}
        </div>
        <div style="display: table-cell; width: 50%; font-size: 11px; padding-left: 10px; vertical-align: top;">
          <strong style="display:block; margin-bottom: 4px; font-size: 10px; color: #64748b;">TO:</strong>
          {{.Shipment.Destination}}
        </div>
      </div>
      <div style="border-top: 1px solid #e2e8f0; padding-top: 8px; font-size: 11px; margin-top: 15px;">
        <strong>SHIPMENT ID:</strong> {{.Shipment.ID}}<br>
        <strong>WEIGHT:</strong> {{.Shipment.Weight}} LBS
      </div>
      <div class="barcode-simulated"></div>
      <div class="tracking-number">TRK# {{.Label.TrackingNumber}}</div>
    </div>

    <div style="text-align: center;">
      <a href="{{.Label.LabelURL}}" class="btn" target="_blank">DOWNLOAD PRINTABLE LABEL</a>
    </div>

    <div class="footer">
      This is an automated notification regarding your shipment created by user <strong>{{.Shipment.Username}}</strong>.<br>
      Recipient Notification Email: <strong>{{.Shipment.Email}}</strong>.<br>
      Antigravity Multi-Carrier Integration Service &copy; 2026. All rights reserved.
    </div>
  </div>
</body>
</html>
`

	tmpl, err := template.New("email").Funcs(template.FuncMap{
		"uppercase": strings.ToUpper,
	}).Parse(emailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Create emails directory
	dir := "./emails"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create emails directory: %w", err)
	}

	filename := filepath.Join(dir, fmt.Sprintf("shipment_%s.html", label.TrackingNumber))

	var bodyBuf bytes.Buffer
	data := struct {
		Shipment *Shipment
		Label    *Label
	}{
		Shipment: shipment,
		Label:    label,
	}

	if err := tmpl.Execute(&bodyBuf, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	htmlContent := bodyBuf.String()

	// 1. Write the executed email to a local HTML file for manual inspection
	if err := os.WriteFile(filename, bodyBuf.Bytes(), 0644); err != nil {
		slog.Error("Failed to write local preview HTML file", slog.String("error", err.Error()))
	}

	// 2. Transmit email via Decoupled Notification Microservice
	subject := fmt.Sprintf("Shipment Dispatched & Label Generated - TRK# %s", label.TrackingNumber)
	if err := s.sendRemoteNotification(ctx, "EMAIL", shipment.Email, subject, htmlContent); err != nil {
		slog.Error("Failed to send remote email notification via microservice",
			slog.String("error", err.Error()),
			slog.String("recipient", shipment.Email),
		)
		// We return the error so that the user receives proper feedback if remote request fails
		return fmt.Errorf("failed to send remote email notification: %w", err)
	}

	slog.Info("Remote email notification sent successfully!",
		slog.String("recipient", shipment.Email),
		slog.String("tracking_number", label.TrackingNumber),
		slog.String("saved_path", filename),
	)

	// 3. Dispatch simulated Telegram message via Decoupled Notification Microservice
	telegramMsg := fmt.Sprintf("Shipment TRK# %s has been generated. Carrier: %s. Origin: %s. Destination: %s.",
		label.TrackingNumber, shipment.Carrier, shipment.Origin, shipment.Destination,
	)
	if err := s.sendRemoteNotification(ctx, "TELEGRAM", "+1 (555) 019-2834", "", telegramMsg); err != nil {
		slog.Error("Failed to send remote Telegram notification via microservice",
			slog.String("error", err.Error()),
		)
	} else {
		slog.Info("Remote Telegram notification sent successfully!",
			slog.String("tracking_number", label.TrackingNumber),
		)
	}

	return nil
}

func (s *shipmentService) sendRemoteNotification(ctx context.Context, method, recipient, subject, body string) error {
	// 1. Attempt to dispatch asynchronously via Apache Kafka
	if s.kafkaPublisher != nil {
		err := s.kafkaPublisher.PublishNotification(ctx, method, recipient, subject, body)
		if err == nil {
			slog.Info("Successfully published asynchronous event to Kafka broker topic",
				slog.String("method", method),
				slog.String("recipient", recipient),
			)
			return nil
		}
		slog.Warn("Kafka event dispatch failed; gracefully falling back to synchronous REST fallback",
			slog.String("error", err.Error()),
			slog.String("method", method),
			slog.String("recipient", recipient),
		)
	}

	// 2. Fallback to direct synchronous REST trigger
	payload := map[string]string{
		"recipient": recipient,
		"method":    method,
		"subject":   subject,
		"body":      body,
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal remote notification request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/notifications", s.notificationServiceURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create remote notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notification service is currently unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errData struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return fmt.Errorf("remote notification server returned error status %d: %s", resp.StatusCode, errData.Error)
	}

	return nil
}

func (s *shipmentService) GetShipment(ctx context.Context, id string) (*Shipment, error) {
	if id == "" {
		return nil, ErrShipmentNotFound
	}
	return s.repo.GetByID(ctx, id)
}

func (s *shipmentService) ListShipments(ctx context.Context, token string) ([]*Shipment, error) {
	username, err := s.verifyToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if username == "admin" {
		return s.repo.List(ctx) // Admin can view all shipments!
	}
	return s.repo.ListByUsername(ctx, username)
}

func (s *shipmentService) UpdateShipment(
	ctx context.Context,
	token, id, carrier string,
	weight float64,
	origin, destination, status string,
) (*Shipment, error) {
	// 1. Authenticate
	username, err := s.verifyToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// 2. Fetch shipment
	shipment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Ownership enforcement (admin bypasses!)
	if username != "admin" && shipment.Username != username {
		return nil, errors.New("forbidden: you do not own this shipment")
	}

	// 4. Validation
	if carrier == "" {
		return nil, ErrCarrierRequired
	}
	if weight <= 0 {
		return nil, ErrInvalidWeight
	}

	// Status validation (must relate to shipping order transit stages)
	if status != "" {
		validStatuses := map[string]bool{
			"CREATED":          true,
			"IN_TRANSIT":       true,
			"OUT_FOR_DELIVERY": true,
			"DELIVERED":        true,
			"CANCELLED":        true,
			"RETURNED":         true,
		}
		if !validStatuses[status] {
			return nil, ErrInvalidStatus
		}
	}

	statusChanged := status != "" && shipment.Status != status
	oldStatus := shipment.Status

	// 5. Commit modifications
	shipment.Carrier = carrier
	shipment.Weight = weight
	shipment.Origin = origin
	shipment.Destination = destination
	if status != "" {
		shipment.Status = status
	}
	shipment.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, shipment); err != nil {
		return nil, err
	}

	s.logAction(ctx, username, fmt.Sprintf("Update Shipment (ID: %s)", id))

	// 6. Trigger automated email on status update
	if statusChanged {
		slog.Info("Shipment status updated. Sending alert...",
			slog.String("id", id),
			slog.String("old_status", oldStatus),
			slog.String("new_status", status),
			slog.String("email", shipment.Email),
		)
		if err := s.sendStatusEmail(ctx, shipment, oldStatus); err != nil {
			slog.Error("Failed to send status update email", slog.String("error", err.Error()), slog.String("id", id))
		}
	}

	return shipment, nil
}

func (s *shipmentService) sendStatusEmail(ctx context.Context, shipment *Shipment, oldStatus string) error {
	const statusTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Shipment Status Updated - TRK# {{.Shipment.TrackingNumber}}</title>
  <style>
    body {
      background-color: #0b0f19;
      color: #f1f5f9;
      font-family: 'Outfit', 'Inter', -apple-system, sans-serif;
      margin: 0;
      padding: 40px 20px;
    }
    .email-container {
      background: #1e293b;
      border: 1px solid rgba(255, 255, 255, 0.08);
      border-radius: 20px;
      max-width: 600px;
      margin: 0 auto;
      padding: 35px;
      box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 10px 10px -5px rgba(0, 0, 0, 0.3);
    }
    .header {
      border-bottom: 1px solid rgba(255, 255, 255, 0.08);
      padding-bottom: 25px;
      margin-bottom: 30px;
      text-align: center;
    }
    .logo {
      font-size: 20px;
      font-weight: 800;
      letter-spacing: 0.15em;
      color: #6366f1;
    }
    .title {
      font-size: 24px;
      font-weight: 700;
      color: #ffffff;
      margin-top: 15px;
      margin-bottom: 5px;
    }
    .subtitle {
      font-size: 14px;
      color: #94a3b8;
      margin: 0;
    }
    .status-alert-box {
      background: linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(139, 92, 246, 0.15));
      border: 1px solid rgba(99, 102, 241, 0.3);
      border-radius: 16px;
      padding: 25px;
      text-align: center;
      margin-top: 25px;
      margin-bottom: 30px;
    }
    .status-badge-old {
      display: inline-block;
      background: #334155;
      color: #94a3b8;
      font-size: 12px;
      font-weight: 700;
      padding: 6px 12px;
      border-radius: 20px;
      text-decoration: line-through;
    }
    .status-arrow {
      font-size: 18px;
      color: #818cf8;
      margin: 0 12px;
      vertical-align: middle;
    }
    .status-badge-new {
      display: inline-block;
      background: #10b981;
      color: #ffffff;
      font-size: 14px;
      font-weight: 800;
      padding: 8px 16px;
      border-radius: 20px;
      box-shadow: 0 0 15px rgba(16, 185, 129, 0.4);
      vertical-align: middle;
    }
    .grid-2 {
      display: table;
      width: 100%;
      table-layout: fixed;
      margin-bottom: 20px;
      border-collapse: separate;
      border-spacing: 15px 0;
    }
    .grid-col {
      display: table-cell;
      width: 50%;
      vertical-align: top;
    }
    .card {
      background: #0f172a;
      border: 1px solid rgba(255, 255, 255, 0.05);
      border-radius: 12px;
      padding: 18px;
    }
    .card-title {
      font-size: 10px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #818cf8;
      margin-bottom: 8px;
    }
    .card-value {
      font-size: 14px;
      color: #e2e8f0;
      line-height: 1.4;
    }
    .footer {
      text-align: center;
      font-size: 11px;
      color: #64748b;
      border-top: 1px solid rgba(255, 255, 255, 0.08);
      padding-top: 25px;
      margin-top: 35px;
    }
  </style>
</head>
<body>
  <div class="email-container">
    <div class="header">
      <div class="logo">ANTIGRAVITY SYSTEMS</div>
      <h1 class="title">Transit Status Updated</h1>
      <p class="subtitle">Your multi-carrier shipment has progressed to a new milestone</p>
    </div>

    <div class="status-alert-box">
      <span class="status-badge-old">{{.OldStatus}}</span>
      <span class="status-arrow">&rarr;</span>
      <span class="status-badge-new">{{.Shipment.Status}}</span>
      <div style="margin-top: 15px; font-size: 13px; color: #cbd5e1; font-weight: 500;">
        TRACKING NUMBER: <strong style="font-family: monospace; color: #ffffff;">{{.Shipment.TrackingNumber}}</strong>
      </div>
    </div>

    <div class="grid-2">
      <div class="grid-col">
        <div class="card">
          <div class="card-title">SHIPPER (ORIGIN)</div>
          <div class="card-value">{{.Shipment.Origin}}</div>
        </div>
      </div>
      <div class="grid-col">
        <div class="card">
          <div class="card-title">RECIPIENT (DESTINATION)</div>
          <div class="card-value">{{.Shipment.Destination}}</div>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <div class="grid-col">
        <div class="card">
          <div class="card-title">CARRIER & WEIGHT</div>
          <div class="card-value"><strong>{{.Shipment.Carrier}}</strong> ({{.Shipment.Weight}} lbs)</div>
        </div>
      </div>
      <div class="grid-col">
        <div class="card">
          <div class="card-title">SHIPMENT ID</div>
          <div class="card-value"><code style="font-size: 12px; font-family: monospace;">{{.Shipment.ID}}</code></div>
        </div>
      </div>
    </div>

    <div class="footer">
      This is an automated tracking notification regarding your shipment.<br>
      Recipient Notification Email: <strong>{{.Shipment.Email}}</strong>.<br>
      Antigravity Multi-Carrier Integration Service &copy; 2026. All rights reserved.
    </div>
  </div>
</body>
</html>
`

	tmpl, err := template.New("statusEmail").Parse(statusTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse status email template: %w", err)
	}

	dir := "./emails"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create emails directory: %w", err)
	}

	filename := filepath.Join(dir, fmt.Sprintf("status_update_%s.html", shipment.TrackingNumber))

	var bodyBuf bytes.Buffer
	data := struct {
		Shipment  *Shipment
		OldStatus string
	}{
		Shipment:  shipment,
		OldStatus: oldStatus,
	}

	if err := tmpl.Execute(&bodyBuf, data); err != nil {
		return fmt.Errorf("failed to execute status email template: %w", err)
	}

	htmlContent := bodyBuf.String()

	// 1. Write status email locally for verification
	if err := os.WriteFile(filename, bodyBuf.Bytes(), 0644); err != nil {
		slog.Error("Failed to write local status update HTML preview file", slog.String("error", err.Error()))
	}

	// 2. Send via Decoupled Notification Microservice
	subject := fmt.Sprintf("Shipment Status Alert - TRK# %s is now %s", shipment.TrackingNumber, shipment.Status)
	if err := s.sendRemoteNotification(ctx, "EMAIL", shipment.Email, subject, htmlContent); err != nil {
		return fmt.Errorf("failed to send remote status alert email: %w", err)
	}

	slog.Info("Remote status alert email notification sent successfully!",
		slog.String("recipient", shipment.Email),
		slog.String("tracking_number", shipment.TrackingNumber),
		slog.String("new_status", shipment.Status),
		slog.String("saved_path", filename),
	)

	// 3. Dispatch simulated Telegram status update via Decoupled Notification Microservice
	telegramMsg := fmt.Sprintf("Transit Milestone Update: Shipment TRK# %s status has progressed from %s to %s.",
		shipment.TrackingNumber, oldStatus, shipment.Status,
	)
	if err := s.sendRemoteNotification(ctx, "TELEGRAM", "+1 (555) 019-2834", "", telegramMsg); err != nil {
		slog.Error("Failed to send remote Telegram status notification via microservice",
			slog.String("error", err.Error()),
		)
	} else {
		slog.Info("Remote Telegram status notification sent successfully!",
			slog.String("tracking_number", shipment.TrackingNumber),
			slog.String("new_status", shipment.Status),
		)
	}

	return nil
}

func (s *shipmentService) DeleteShipment(ctx context.Context, token, id string) error {
	// 1. Authenticate
	username, err := s.verifyToken(ctx, token)
	if err != nil {
		return err
	}

	// 2. Fetch shipment
	shipment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 3. Ownership enforcement
	if shipment.Username != username {
		return errors.New("forbidden: you do not own this shipment")
	}

	// 4. Delete
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.logAction(ctx, username, fmt.Sprintf("Delete Shipment (ID: %s)", id))

	return nil
}

func (s *shipmentService) CancelShipmentByTracking(ctx context.Context, trackingNumber string) error {
	if trackingNumber == "" {
		return ErrShipmentNotFound
	}

	shipment, err := s.repo.GetByTracking(ctx, trackingNumber)
	if err != nil {
		return err
	}

	shipment.Status = "CANCELLED"
	shipment.UpdatedAt = time.Now()

	return s.repo.Update(ctx, shipment)
}

// generateUUID creates a standard secure random UUID
func generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant is 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
