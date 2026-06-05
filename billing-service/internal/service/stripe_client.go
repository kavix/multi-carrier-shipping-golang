package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type StripeClient struct {
	secretKey  string
	httpClient *http.Client
}

func NewStripeClient(secretKey string) *StripeClient {
	return &StripeClient{
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type StripeChargeResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (c *StripeClient) Charge(ctx context.Context, amount float64, currency string) (string, error) {
	// Stripe expects the amount in the smallest currency unit (cents for USD)
	amountInCents := int64(amount * 100)

	data := url.Values{}
	data.Set("amount", fmt.Sprintf("%d", amountInCents))
	data.Set("currency", strings.ToLower(currency))
	data.Set("payment_method", "pm_card_visa") // default Stripe test card
	data.Set("confirm", "true")
	data.Set("automatic_payment_methods[enabled]", "true")
	data.Set("automatic_payment_methods[allow_redirects]", "never")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.stripe.com/v1/payment_intents", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error.Message != "" {
			return "", fmt.Errorf("stripe api error: %s", errResp.Error.Message)
		}
		return "", fmt.Errorf("stripe api returned status %d", resp.StatusCode)
	}

	var chargeResp StripeChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResp); err != nil {
		return "", fmt.Errorf("decode stripe response: %w", err)
	}

	if chargeResp.Status != "succeeded" && chargeResp.Status != "processing" {
		return "", fmt.Errorf("stripe charge status: %s", chargeResp.Status)
	}

	return chargeResp.ID, nil
}
