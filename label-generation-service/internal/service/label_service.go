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

	"github.com/johnfercher/maroto/pkg/color"
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

	// Default to the Maroto-generated PDF. For FedEx we may replace both
	// trackingNumber and pdfData below with the real FedEx response.
	pdfData, err := s.generateMarotoPDF(details, trackingNumber)
	if err != nil {
		return nil, fmt.Errorf("generate maroto pdf: %w", err)
	}

	// FedEx branch: ask carrier-integration-service to create a real FedEx
	// shipment and hand us back the official label PDF + tracking number.
	// If the integration call fails for any reason, we keep the Maroto
	// fallback so the label pipeline never breaks.
	labelSource := "maroto"
	if strings.EqualFold(carrier, "fedex") {
		if fedexPDF, fedexTracking, fedexErr := s.createFedExShipment(ctx, carrierServiceURL, details); fedexErr != nil {
			logger.Error("fedex create-shipment call failed, falling back to maroto label",
				logger.String("shipment_id", shipmentID),
				logger.String("err", fedexErr.Error()))
		} else {
			pdfData = fedexPDF
			trackingNumber = fedexTracking
			labelSource = "fedex"
			logger.Info("fedex label retrieved from carrier-integration-service",
				logger.String("shipment_id", shipmentID),
				logger.String("tracking_number", trackingNumber),
				logger.Int("label_bytes", len(pdfData)))
		}
	}

	labelData := base64.StdEncoding.EncodeToString(pdfData)

	// Default fallback URL points to API Gateway download endpoint
	gatewayURL := "http://localhost:8080"
	if s.cfg != nil && s.cfg.APIGatewayURL != "" {
		gatewayURL = strings.TrimRight(s.cfg.APIGatewayURL, "/")
	}

	label := &domain.ShippingLabel{
		ID:             utils.GenerateID(),
		ShipmentID:     shipmentID,
		Carrier:        carrier,
		TrackingNumber: trackingNumber,
		LabelData:      labelData,
		LabelURL:       fmt.Sprintf("%s/labels/%s/download", gatewayURL, shipmentID),
		Format:         "PDF",
		Status:         "generated",
		CreatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, label); err != nil {
		return nil, fmt.Errorf("save label: %w", err)
	}

	logger.Info("label generated",
		logger.String("shipment_id", shipmentID),
		logger.String("carrier", carrier),
		logger.String("source", labelSource))

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

func formatAddressLines(addr string) []string {
	parts := strings.Split(addr, ",")
	var lines []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

func getServiceIndicator(carrier, serviceType string) (letter string, description string) {
	carrierUpper := strings.ToUpper(carrier)
	serviceLower := strings.ToLower(serviceType)

	switch carrierUpper {
	case "FEDEX":
		if strings.Contains(serviceLower, "ground") {
			return "G", "FedEx Ground"
		}
		return "E", "FedEx Express"
	case "UPS":
		if strings.Contains(serviceLower, "ground") {
			return "UG", "UPS Ground"
		}
		return "UA", "UPS Air"
	case "DHL":
		return "D", "DHL Express"
	case "USPS":
		return "P", "USPS Priority"
	default:
		return "S", "Standard Shipping"
	}
}

func (s *LabelService) generateMarotoPDF(details map[string]interface{}, trackingNumber string) ([]byte, error) {
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	// Set default margins to prevent rendering artifacts caused by changing page margins mid-document
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

	// Format Sender Address
	var senderStreet, senderCityStateZip string
	if sv, ok := details["sender_validated"].(map[string]interface{}); ok && sv["is_valid"] == true {
		street, _ := sv["street"].(string)
		city, _ := sv["city"].(string)
		state, _ := sv["state"].(string)
		zip, _ := sv["postal_code"].(string)
		senderStreet = street
		senderCityStateZip = fmt.Sprintf("%s, %s %s", city, state, zip)
	} else {
		lines := formatAddressLines(senderAddr)
		if len(lines) > 0 {
			senderStreet = lines[0]
			if len(lines) > 1 {
				senderCityStateZip = strings.Join(lines[1:], ", ")
			}
		}
	}

	// Format Receiver Address
	var receiverStreet, receiverCityStateZip string
	if rv, ok := details["receiver_validated"].(map[string]interface{}); ok && rv["is_valid"] == true {
		street, _ := rv["street"].(string)
		city, _ := rv["city"].(string)
		state, _ := rv["state"].(string)
		zip, _ := rv["postal_code"].(string)
		receiverStreet = street
		receiverCityStateZip = fmt.Sprintf("%s, %s %s", city, state, zip)
	} else {
		lines := formatAddressLines(receiverAddr)
		if len(lines) > 0 {
			receiverStreet = lines[0]
			if len(lines) > 1 {
				receiverCityStateZip = strings.Join(lines[1:], ", ")
			}
		}
	}

	// Determine carrier specific details and class letters
	serviceLetter, serviceDesc := getServiceIndicator(carrier, serviceType)

	// Determine rate text
	rateText := "N/A"
	if rateCost > 0 {
		rateText = fmt.Sprintf("%.2f %s", rateCost, rateCurrency)
		if rateServiceLabel != "" && rateServiceLabel != "FedEx Rate" {
			rateText = fmt.Sprintf("%.2f %s (%s)", rateCost, rateCurrency, rateServiceLabel)
		}
	}

	receiverZip := extractPostalCode(receiverAddr)
	if receiverZip == "" {
		receiverZip = "00000"
	}

	// Enable borders around the grid cells to create structured blocks
	m.SetBorder(true)

	// Row 1 (Header/From/Service): 35 mm
	m.Row(35, func() {
		m.Col(8, func() {
			m.Text("FROM:", props.Text{Size: 7, Style: consts.Bold, Top: 2, Left: 4})
			m.Text(senderName, props.Text{Size: 9, Style: consts.Bold, Top: 7, Left: 4})
			m.Text(senderStreet, props.Text{Size: 8, Top: 13, Left: 4})
			m.Text(senderCityStateZip, props.Text{Size: 8, Top: 19, Left: 4})
		})
		m.SetBackgroundColor(color.Color{Red: 0, Green: 0, Blue: 0})
		m.Col(4, func() {
			m.Text(serviceLetter, props.Text{
				Size:  24,
				Style: consts.Bold,
				Align: consts.Center,
				Top:   6,
				Color: color.Color{Red: 255, Green: 255, Blue: 255},
			})
			m.Text(serviceDesc, props.Text{
				Size:  7,
				Style: consts.Bold,
				Align: consts.Center,
				Top:   24,
				Color: color.Color{Red: 255, Green: 255, Blue: 255},
			})
		})
		m.SetBackgroundColor(color.Color{Red: 255, Green: 255, Blue: 255})
	})

	// Row 2 (To Address): 50 mm
	m.Row(50, func() {
		m.Col(12, func() {
			m.Text("SHIP TO:", props.Text{Size: 8, Style: consts.Bold, Top: 3, Left: 6})
			m.Text(receiverName, props.Text{Size: 14, Style: consts.Bold, Top: 10, Left: 6})
			m.Text(receiverStreet, props.Text{Size: 11, Top: 21, Left: 6})
			m.Text(receiverCityStateZip, props.Text{Size: 11, Top: 31, Left: 6})
		})
	})

	// Row 3 (Carrier Banner): 22 mm
	m.SetBackgroundColor(color.Color{Red: 0, Green: 0, Blue: 0})
	m.Row(22, func() {
		m.Col(12, func() {
			m.Text(strings.ToUpper(carrier)+" - "+strings.ToUpper(serviceType), props.Text{
				Size:  16,
				Style: consts.Bold,
				Align: consts.Center,
				Top:   7,
				Color: color.Color{Red: 255, Green: 255, Blue: 255},
			})
		})
	})
	m.SetBackgroundColor(color.Color{Red: 255, Green: 255, Blue: 255})

	// Row 4 (Routing/DataMatrix): 35 mm
	m.Row(35, func() {
		m.Col(8, func() {
			m.Text("RTP / POSTAL CODE:", props.Text{Size: 7, Style: consts.Bold, Top: 2, Left: 4})
			m.Barcode(receiverZip, props.Barcode{
				Percent: 75,
				Center:  true,
				Top:     7,
			})
			m.Text(receiverZip, props.Text{Size: 8, Align: consts.Center, Top: 24})
		})
		m.Col(4, func() {
			m.Text("DATA MATRIX", props.Text{Size: 7, Style: consts.Bold, Top: 2, Align: consts.Center})
			m.DataMatrixCode(shipmentID, props.Rect{
				Percent: 70,
				Center:  true,
				Top:     7,
			})
		})
	})

	// Row 5 (Package Details): 22 mm
	m.Row(22, func() {
		m.Col(3, func() {
			m.Text("SHIPMENT ID", props.Text{Size: 7, Style: consts.Bold, Top: 3, Left: 4})
			m.Text(shipmentID, props.Text{Size: 8, Top: 11, Left: 4})
		})
		m.Col(3, func() {
			m.Text("WEIGHT", props.Text{Size: 7, Style: consts.Bold, Top: 3, Left: 4})
			m.Text(fmt.Sprintf("%.2f KG", weight), props.Text{Size: 8, Top: 11, Left: 4})
		})
		m.Col(3, func() {
			m.Text("SERVICE TYPE", props.Text{Size: 7, Style: consts.Bold, Top: 3, Left: 4})
			m.Text(strings.ToUpper(serviceType), props.Text{Size: 8, Top: 11, Left: 4})
		})
		m.Col(3, func() {
			m.Text("POSTAGE/RATE", props.Text{Size: 7, Style: consts.Bold, Top: 3, Left: 4})
			m.Text(rateText, props.Text{Size: 8, Top: 11, Left: 4})
		})
	})

	// Row 6 (Tracking Barcode): 72 mm
	m.Row(72, func() {
		m.Col(12, func() {
			m.Text("TRACKING NUMBER:", props.Text{Size: 8, Style: consts.Bold, Top: 3, Left: 6})
			m.Barcode(trackingNumber, props.Barcode{
				Percent: 80,
				Center:  true,
				Top:     9,
			})
			m.Text(trackingNumber, props.Text{
				Size:  14,
				Style: consts.Bold,
				Align: consts.Center,
				Top:   56,
			})
		})
	})

	// Row 7 (Footer/Notes): 24 mm
	m.Row(24, func() {
		m.Col(12, func() {
			m.Text("INSTRUCTIONS / NOTES:", props.Text{Size: 6, Style: consts.Bold, Top: 2, Left: 4})
			m.Text("Deliver to recipient. Please handle with care. Return service requested.", props.Text{Size: 7, Top: 8, Left: 4})
			m.Text("System generated label. No signature required on delivery.", props.Text{Size: 7, Top: 14, Left: 4})
		})
	})

	m.SetBorder(false)

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

// fedexCreateShipmentRequest mirrors the JSON DTO accepted by
// POST /carriers/fedex/create-shipment in the carrier-integration-service.
type fedexCreateShipmentRequest struct {
	AccountNumber                 string          `json:"account_number"`
	ServiceType                   string          `json:"service_type"`
	PackagingType                 string          `json:"packaging_type"`
	Weight                        float64         `json:"weight"`
	WeightUnits                   string          `json:"weight_units"`
	Sender                        fedexAddressDTO `json:"sender"`
	SenderContact                 fedexContactDTO `json:"sender_contact"`
	Recipient                     fedexAddressDTO `json:"recipient"`
	RecipientContact              fedexContactDTO `json:"recipient_contact"`
	IsInternational               bool            `json:"is_international"`
	TotalCustomsValue             float64         `json:"total_customs_value"`
	TotalCustomsCurrency          string          `json:"total_customs_currency"`
	CommodityDescription          string          `json:"commodity_description"`
	CommodityCountryOfManufacture string          `json:"commodity_country_of_manufacture"`
	CommodityQuantity             int             `json:"commodity_quantity"`
	CommodityUnitPrice            float64         `json:"commodity_unit_price"`
}

type fedexAddressDTO struct {
	StreetLines     []string `json:"street_lines"`
	City            string   `json:"city"`
	StateOrProvince string   `json:"state_or_province_code"`
	PostalCode      string   `json:"postal_code"`
	CountryCode     string   `json:"country_code"`
}

type fedexContactDTO struct {
	PersonName  string `json:"person_name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	CompanyName string `json:"company_name"`
}

// buildFedExAddressDTO turns a free-form address string into a structured
// FedEx address by reusing the existing address-line splitter and the
// postal-code extractor.
func buildFedExAddressDTO(addr string) fedexAddressDTO {
	lines := formatAddressLines(addr)
	postal := extractPostalCode(addr)
	return fedexAddressDTO{
		StreetLines: lines,
		PostalCode:  postal,
		CountryCode: "US",
	}
}

func buildFedExContactDTO(name, email string) fedexContactDTO {
	return fedexContactDTO{
		PersonName: name,
		Email:      email,
	}
}

// fedexCreateShipmentResponse mirrors the JSON returned by the
// carrier-integration-service /carriers/fedex/create-shipment endpoint.
type fedexCreateShipmentResponse struct {
	TrackingNumber string `json:"tracking_number"`
	LabelPDFB64    string `json:"label_pdf_b64"`
	LabelFormat    string `json:"label_format"`
	ServiceType    string `json:"service_type"`
}

// createFedExShipment POSTs the FedEx create-shipment request to the
// carrier-integration-service and returns the decoded label PDF bytes plus
// the FedEx master tracking number. The caller is expected to fall back to
// the Maroto-generated label when this returns an error.
func (s *LabelService) createFedExShipment(ctx context.Context, carrierServiceURL string, details map[string]interface{}) ([]byte, string, error) {
	shipmentID, _ := details["shipment_id"].(string)
	senderName, _ := details["sender_name"].(string)
	senderAddr, _ := details["sender"].(string)
	senderEmail, _ := details["sender_email"].(string)
	receiverName, _ := details["receiver_name"].(string)
	receiverAddr, _ := details["receiver"].(string)
	receiverEmail, _ := details["receiver_email"].(string)
	serviceType, _ := details["service_type"].(string)
	weightF, _ := details["weight"].(float64)
	isInternational, _ := details["is_international"].(bool)
	customsValue, _ := details["customs_value"].(float64)
	customsCurrency, _ := details["customs_currency"].(string)
	commodityDescription, _ := details["description"].(string)

	// Default to FedEx Ground for domestic; INTERNATIONAL_PRIORITY for international
	if serviceType == "" {
		if isInternational {
			serviceType = "INTERNATIONAL_PRIORITY"
		} else {
			serviceType = "FEDEX_GROUND"
		}
	}

	if customsCurrency == "" {
		customsCurrency = "USD"
	}
	if commodityDescription == "" {
		commodityDescription = "Shipping Items"
	}

	payload := fedexCreateShipmentRequest{
		AccountNumber:  "740561073", // sandbox account, matches Python script
		ServiceType:    serviceType,
		PackagingType:  "YOUR_PACKAGING",
		Weight:         weightF,
		WeightUnits:    "LB",
		Sender:         buildFedExAddressDTO(senderAddr),
		SenderContact:  buildFedExContactDTO(senderName, senderEmail),
		Recipient:      buildFedExAddressDTO(receiverAddr),
		RecipientContact: buildFedExContactDTO(receiverName, receiverEmail),
		IsInternational:  isInternational,
		TotalCustomsValue: customsValue,
		TotalCustomsCurrency: customsCurrency,
		CommodityDescription: commodityDescription,
		CommodityQuantity: 1,
		CommodityUnitPrice: customsValue,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("marshal fedex request: %w", err)
	}

	endpoint := strings.TrimRight(carrierServiceURL, "/") + "/carriers/fedex/create-shipment"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("build fedex http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("call fedex endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("fedex endpoint returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed fedexCreateShipmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, "", fmt.Errorf("decode fedex response: %w", err)
	}
	if parsed.LabelPDFB64 == "" {
		return nil, "", fmt.Errorf("fedex response missing label_pdf_b64")
	}
	if parsed.TrackingNumber == "" {
		return nil, "", fmt.Errorf("fedex response missing tracking_number")
	}

	pdfBytes, err := base64.StdEncoding.DecodeString(parsed.LabelPDFB64)
	if err != nil {
		return nil, "", fmt.Errorf("decode fedex label pdf: %w", err)
	}

	logger.Info("fedex create-shipment succeeded",
		logger.String("shipment_id", shipmentID),
		logger.String("tracking_number", parsed.TrackingNumber),
		logger.String("label_format", parsed.LabelFormat),
		logger.Int("label_bytes", len(pdfBytes)))

	return pdfBytes, parsed.TrackingNumber, nil
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
