package carrierstats

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type carrierStatsService struct {
	repo       CarrierStatsRepository
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewCarrierStatsService instantiates a new FreightPulse carrier stats service.
func NewCarrierStatsService(repo CarrierStatsRepository, baseURL, apiKey string) CarrierStatsService {
	return &carrierStatsService{
		repo:       repo,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *carrierStatsService) GetPortCongestion(ctx context.Context) ([]byte, error) {
	return s.fetch(ctx, "/port-congestion")
}

func (s *carrierStatsService) GetFreightRates(ctx context.Context) ([]byte, error) {
	return s.fetch(ctx, "/freight-rates")
}

func (s *carrierStatsService) GetFuelPrices(ctx context.Context) ([]byte, error) {
	return s.fetch(ctx, "/fuel-prices")
}

func (s *carrierStatsService) GetDisruptions(ctx context.Context) ([]byte, error) {
	return s.fetch(ctx, "/disruptions")
}

func (s *carrierStatsService) GetCarriers(ctx context.Context) ([]byte, error) {
	return s.fetch(ctx, "/carriers")
}

func (s *carrierStatsService) ListLogs(ctx context.Context, limit int64) ([]*CarrierStatsLog, error) {
	return s.repo.List(ctx, limit)
}

func (s *carrierStatsService) fetch(ctx context.Context, endpoint string) ([]byte, error) {
	requestURL := s.baseURL + endpoint
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create carrier stats request: %w", err)
	}
	if s.apiKey != "" {
		request.Header.Set("X-API-Key", s.apiKey)
	}

	startedAt := time.Now()
	response, err := s.httpClient.Do(request)
	if err != nil {
		s.recordLog(ctx, endpoint, requestURL, 0, time.Since(startedAt), nil, false, err)
		return nil, fmt.Errorf("freightpulse request failed: %w", err)
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		s.recordLog(ctx, endpoint, requestURL, response.StatusCode, time.Since(startedAt), nil, false, readErr)
		return nil, fmt.Errorf("failed to read freightpulse response: %w", readErr)
	}

	if response.StatusCode != http.StatusOK {
		s.recordLog(ctx, endpoint, requestURL, response.StatusCode, time.Since(startedAt), body, false, fmt.Errorf("upstream returned %d", response.StatusCode))
		return nil, fmt.Errorf("freightpulse %s returned status %d", endpoint, response.StatusCode)
	}

	s.recordLog(ctx, endpoint, requestURL, response.StatusCode, time.Since(startedAt), body, true, nil)
	return body, nil
}

func (s *carrierStatsService) recordLog(ctx context.Context, endpoint, requestURL string, statusCode int, duration time.Duration, body []byte, success bool, err error) {
	if s.repo == nil {
		return
	}

	preview := ""
	if len(body) > 0 {
		preview = string(body)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
	}

	logRecord := &CarrierStatsLog{
		Endpoint:        endpoint,
		URL:             requestURL,
		StatusCode:      statusCode,
		DurationMS:      duration.Milliseconds(),
		ResponseSize:    len(body),
		Success:         success,
		ResponsePreview: preview,
		CreatedAt:       time.Now(),
	}
	if err != nil {
		logRecord.Error = err.Error()
	}

	_ = s.repo.Create(ctx, logRecord)
}

func generateLogID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("log-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
