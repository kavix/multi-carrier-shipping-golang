package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/shipping/label-generation-service/internal/config"
	"github.com/shipping/label-generation-service/internal/domain"
	"github.com/shipping/label-generation-service/internal/repository"
	"github.com/shipping/shared/pkg/kafka"
	"github.com/shipping/shared/pkg/logger"
	"github.com/shipping/shared/pkg/utils"

	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

type LabelService struct {
	repo     *repository.LabelRepo
	producer *kafka.Producer
	cfg      *config.Config
}

func NewLabelService(repo *repository.LabelRepo, producer *kafka.Producer, cfg *config.Config) *LabelService {
	return &LabelService{repo: repo, producer: producer, cfg: cfg}
}

func (s *LabelService) GenerateLabel(ctx context.Context, details map[string]interface{}) (*domain.ShippingLabel, error) {
	shipmentID, _ := details["shipment_id"].(string)
	carrier, _ := details["carrier"].(string)

	if shipmentID == "" {
		return nil, fmt.Errorf("missing shipment_id in details")
	}

	// Enrich event with a FedEx rate if this shipment uses FedEx
	if strings.EqualFold(carrier, "fedex") {
		if rateCost, rateCurrency, rateService, err := s.lookupFedExRate(ctx, details); err == nil {
			details["rate_cost"] = rateCost
			details["rate_currency"] = rateCurrency
			details["rate_service"] = rateService
		} else {
			logger.Error("failed to lookup fedex rate", logger.String("shipment_id", shipmentID), logger.String("err", err.Error()))
		}
	}

	// Call carrier integration service to generate label
	carrierServiceURL := getEnv("CARRIER_SERVICE_URL", "http://carrier-integration-service:8082")

	// In production, this calls the carrier API to generate a real label
	// For demo, we generate realistic tracking numbers based on carrier
	trackingNumber := s.generateRealisticTrackingNumber(carrier)

	// Generate PDF using Maroto
	pdfData, err := s.generateMarotoPDF(details, trackingNumber)
	if err != nil {
		return nil, fmt.Errorf("generate maroto pdf: %w", err)
	}

	labelData := base64.StdEncoding.EncodeToString(pdfData)

	label := &domain.ShippingLabel{
		ID:             utils.GenerateID(),
		ShipmentID:     shipmentID,
		Carrier:        carrier,
		TrackingNumber: trackingNumber,
		LabelData:      labelData,
		LabelURL:       fmt.Sprintf("%s/labels/download/%s", carrierServiceURL, shipmentID),
		Format:         "PDF",
		Status:         "generated",
		CreatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, label); err != nil {
		return nil, fmt.Errorf("save label: %w", err)
	}

	logger.Info("label generated initially with maroto", logger.String("shipment_id", shipmentID), logger.String("carrier", carrier))

	// If S3 bucket is configured, upload the label PDF to S3 and update LabelURL
	if s.cfg != nil && s.cfg.S3BucketARN != "" {
		// parse bucket name from ARN (arn:aws:s3:::bucket-name)
		parts := strings.Split(s.cfg.S3BucketARN, ":::")
		bucket := parts[len(parts)-1]

		data := pdfData
		if err == nil {
			// load aws config using request context so timeouts/cancel propagate
			var awsCfg aws.Config
			if s.cfg.AWSAccessKey != "" && s.cfg.AWSSecretKey != "" {
				awsCfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(s.cfg.AWSRegion), awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s.cfg.AWSAccessKey, s.cfg.AWSSecretKey, "")))
			} else {
				awsCfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(s.cfg.AWSRegion))
			}
			if err != nil {
				logger.Error("aws config load failed", logger.String("err", err.Error()))
			} else {
				key := fmt.Sprintf("labels/%s.pdf", label.ShipmentID)

				// Retry logic: up to 3 attempts with exponential backoff
				var lastErr error
				for attempt := 1; attempt <= 3; attempt++ {
					logger.Info("s3 upload attempt", logger.String("label_id", label.ID), logger.String("bucket", bucket), logger.String("region", s.cfg.AWSRegion), logger.String("attempt", fmt.Sprintf("%d", attempt)))

					// build object URL
					objectURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, s.cfg.AWSRegion, key)

					// sign request using v4 signer
					req, rErr := http.NewRequestWithContext(ctx, "PUT", objectURL, bytes.NewReader(data))
					if rErr != nil {
						lastErr = rErr
						logger.Error("failed to create http request for s3 put", logger.String("err", rErr.Error()))
						break
					}
					req.Header.Set("Content-Type", "application/pdf")

					// retrieve credentials
					creds, credErr := awsCfg.Credentials.Retrieve(ctx)
					if credErr != nil {
						lastErr = credErr
						logger.Error("failed to retrieve aws credentials", logger.String("err", credErr.Error()))
						break
					}

					// compute payload hash
					sum := sha256.Sum256(data)
					payloadHash := hex.EncodeToString(sum[:])
					req.Header.Set("x-amz-content-sha256", payloadHash)
					req.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))

					signer := v4.NewSigner()
					if signErr := signer.SignHTTP(ctx, creds, req, payloadHash, "s3", s.cfg.AWSRegion, time.Now()); signErr != nil {
						lastErr = signErr
						logger.Error("v4 signing failed", logger.String("err", signErr.Error()))
					} else {
						// do the HTTP put
						resp, httpErr := http.DefaultClient.Do(req)
						if httpErr != nil {
							lastErr = httpErr
							logger.Error("http put to s3 failed", logger.String("err", httpErr.Error()))
						} else {
							respBody, _ := io.ReadAll(resp.Body)
							resp.Body.Close()
							if resp.StatusCode >= 200 && resp.StatusCode < 300 {
								label.LabelURL = objectURL
								if updateErr := s.repo.UpdateLabelURL(ctx, label.ID, label.LabelURL); updateErr != nil {
									logger.Error("update label url failed", logger.String("err", updateErr.Error()))
								} else {
									logger.Info("label uploaded to s3", logger.String("label_id", label.ID), logger.String("url", label.LabelURL))
								}
								lastErr = nil
								break
							}
							lastErr = fmt.Errorf("s3 put returned status %d: %s", resp.StatusCode, string(respBody))
							logger.Error("s3 put returned non-2xx", logger.String("status", fmt.Sprintf("%d", resp.StatusCode)), logger.String("body", string(respBody)))
						}
					}

					// simple backoff
					backoff := time.Duration(attempt*500) * time.Millisecond
					select {
					case <-ctx.Done():
						logger.Error("context cancelled during s3 upload", logger.String("err", ctx.Err().Error()))
						return label, ctx.Err()
					case <-time.After(backoff):
					}
				}
				if lastErr != nil {
					logger.Error("s3 upload failed after retries", logger.String("err", lastErr.Error()))
				}
			}
		} else {
			logger.Error("decode label pdf failed", logger.String("err", err.Error()))
		}
	}

	// Publish event with final LabelURL (either S3 or local)
	event := map[string]interface{}{
		"label_id":        label.ID,
		"shipment_id":     shipmentID,
		"carrier":         carrier,
		"tracking_number": trackingNumber,
		"label_url":       label.LabelURL,
		"event_type":      "label.generated",
	}
	if err := s.producer.Publish(ctx, label.ID, event); err != nil {
		logger.Error("failed to publish label.generated", logger.String("err", err.Error()))
	}

	logger.Info("label generation process completed", logger.String("shipment_id", shipmentID), logger.String("label_url", label.LabelURL))
	return label, nil
}

func (s *LabelService) generateRealisticTrackingNumber(carrier string) string {
	switch strings.ToLower(carrier) {
	case "dhl":
		// 10 digits
		return utils.RandomDigits(10)
	case "fedex":
		// 12 digits
		return utils.RandomDigits(12)
	case "ups":
		// 1Z + 16 alphanumeric
		return "1Z" + utils.RandomAlphanumeric(16)
	case "usps":
		// 22 digits
		return utils.RandomDigits(22)
	case "ems", "slpost":
		// EE + 8 digits + LK
		return "EE" + utils.RandomDigits(8) + "LK"
	default:
		return fmt.Sprintf("TRACK-%s-%d", carrier, time.Now().Unix())
	}
}

func (s *LabelService) generateMarotoPDF(details map[string]interface{}, trackingNumber string) ([]byte, error) {
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(10, 10, 10)

	shipmentID, _ := details["shipment_id"].(string)
	carrier, _ := details["carrier"].(string)
	senderName, _ := details["sender_name"].(string)
	senderAddr, _ := details["sender"].(string)
	receiverName, _ := details["receiver_name"].(string)
	receiverAddr, _ := details["receiver"].(string)
	weight, _ := details["weight"].(float64)
	serviceType, _ := details["service_type"].(string)
	rateCost, _ := details["rate_cost"].(float64)
	rateCurrency, _ := details["rate_currency"].(string)
	rateServiceLabel, _ := details["rate_service"].(string)
	pickupLocID, _ := details["pickup_location_id"].(string)
	dropLocID, _ := details["drop_location_id"].(string)

	// Use validated addresses if available
	if sv, ok := details["sender_validated"].(map[string]interface{}); ok && sv["is_valid"] == true {
		street, _ := sv["street"].(string)
		city, _ := sv["city"].(string)
		state, _ := sv["state"].(string)
		zip, _ := sv["postal_code"].(string)
		senderAddr = fmt.Sprintf("%s, %s, %s %s (VALIDATED)", street, city, state, zip)
	}
	if rv, ok := details["receiver_validated"].(map[string]interface{}); ok && rv["is_valid"] == true {
		street, _ := rv["street"].(string)
		city, _ := rv["city"].(string)
		state, _ := rv["state"].(string)
		zip, _ := rv["postal_code"].(string)
		receiverAddr = fmt.Sprintf("%s, %s, %s %s (VALIDATED)", street, city, state, zip)
	}

	m.Row(20, func() {
		m.Col(12, func() {
			m.Text("SHIPPING LABEL", props.Text{
				Size:  16,
				Align: consts.Center,
				Style: consts.Bold,
			})
		})
	})

	m.Row(10, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Carrier: %s | Service: %s", carrier, serviceType), props.Text{
				Size:  10,
				Align: consts.Center,
			})
		})
	})

	m.Line(5)

	m.Row(40, func() {
		m.Col(6, func() {
			m.Text("FROM:", props.Text{Size: 8, Style: consts.Bold})
			m.Text(senderName, props.Text{Size: 10, Top: 5})
			m.Text(senderAddr, props.Text{Size: 10, Top: 15})
		})
		m.Col(6, func() {
			m.Text("TO:", props.Text{Size: 8, Style: consts.Bold})
			m.Text(receiverName, props.Text{Size: 10, Top: 5})
			m.Text(receiverAddr, props.Text{Size: 10, Top: 15})
		})
	})

	m.Line(5)

	m.Row(20, func() {
		m.Col(6, func() {
			m.Text(fmt.Sprintf("WEIGHT: %.2f KG", weight), props.Text{Size: 10, Style: consts.Bold})
		})
		m.Col(6, func() {
			m.Text(fmt.Sprintf("SHIPMENT ID: %s", shipmentID), props.Text{Size: 8})
		})
	})

	if rateCost > 0 {
		m.Row(20, func() {
			m.Col(6, func() {
				m.Text(fmt.Sprintf("RATE: %.2f %s", rateCost, rateCurrency), props.Text{Size: 10, Style: consts.Bold})
			})
			m.Col(6, func() {
				label := "FedEx Rate"
				if rateServiceLabel != "" {
					label = rateServiceLabel
				}
				m.Text(label, props.Text{Size: 8})
			})
		})
	}

	if pickupLocID != "" || dropLocID != "" {
		m.Row(20, func() {
			if pickupLocID != "" {
				m.Col(6, func() {
					m.Text(fmt.Sprintf("PICKUP AT: %s", pickupLocID), props.Text{Size: 8, Style: consts.Bold})
				})
			}
			if dropLocID != "" {
				m.Col(6, func() {
					m.Text(fmt.Sprintf("DROP OFF AT: %s", dropLocID), props.Text{Size: 8, Style: consts.Bold})
				})
			}
		})
	}

	m.Row(30, func() {
		m.Col(12, func() {
			m.Barcode(trackingNumber, props.Barcode{
				Center:  true,
				Percent: 80,
			})
			m.Text(trackingNumber, props.Text{
				Size:  10,
				Align: consts.Center,
				Top:   22,
			})
		})
	})

	buf, err := m.Output()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *LabelService) lookupFedExRate(ctx context.Context, details map[string]interface{}) (float64, string, string, error) {
	senderAddr, _ := details["sender"].(string)
	receiverAddr, _ := details["receiver"].(string)
	weight, _ := details["weight"].(float64)
	requestedServiceType, _ := details["service_type"].(string)

	carrierServiceURL := getEnv("CARRIER_SERVICE_URL", "http://carrier-integration-service:8082")
	from := extractPostalCode(senderAddr)
	to := extractPostalCode(receiverAddr)
	if from == "" {
		from = senderAddr
	}
	if to == "" {
		to = receiverAddr
	}

	rateURL := fmt.Sprintf("%s/carriers/rates?from=%s&to=%s&weight=%.2f",
		carrierServiceURL,
		url.QueryEscape(from),
		url.QueryEscape(to),
		weight,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", rateURL, nil)
	if err != nil {
		return 0, "", "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, "", "", fmt.Errorf("rate endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	type rateResult struct {
		CarrierID     string  `json:"carrier_id"`
		CarrierName   string  `json:"carrier_name"`
		ServiceType   string  `json:"service_type"`
		Currency      string  `json:"currency"`
		Cost          float64 `json:"cost"`
		EstimatedDays int     `json:"estimated_days"`
	}
	var rates []rateResult
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return 0, "", "", err
	}

	var selected *rateResult
	for _, r := range rates {
		if !strings.EqualFold(r.CarrierName, "FedEx") && !strings.EqualFold(r.CarrierID, "fedex") {
			continue
		}
		if selected == nil || r.Cost < selected.Cost {
			selected = &r
		}
		if requestedServiceType != "" && strings.Contains(strings.ToLower(r.ServiceType), strings.ToLower(requestedServiceType)) {
			selected = &r
			break
		}
	}

	if selected == nil {
		return 0, "", "", fmt.Errorf("no FedEx rate found")
	}

	return selected.Cost, selected.Currency, selected.ServiceType, nil
}

var postalCodeRegex = regexp.MustCompile(`\b\d{5}(?:-\d{4})?\b`)

func extractPostalCode(address string) string {
	return postalCodeRegex.FindString(address)
}

func (s *LabelService) GetLabel(ctx context.Context, shipmentID string) (*domain.ShippingLabel, error) {
	return s.repo.GetByShipmentID(ctx, shipmentID)
}

func (s *LabelService) DownloadLabel(ctx context.Context, shipmentID string) ([]byte, error) {
	label, err := s.repo.GetByShipmentID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(label.LabelData)
	if err != nil {
		return nil, fmt.Errorf("decode label: %w", err)
	}
	return data, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
