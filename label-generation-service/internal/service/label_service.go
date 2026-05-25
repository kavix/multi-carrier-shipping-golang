package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
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

	// Call carrier integration service to generate label
	carrierServiceURL := getEnv("CARRIER_SERVICE_URL", "http://carrier-integration-service:8082")

	// In production, this calls the carrier API to generate a real label
	// For demo, we simulate label generation using Maroto
	trackingNumber := fmt.Sprintf("TRACK-%s-%d", carrier, time.Now().Unix())

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
