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

type StripeCheckoutSessionResponse struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	PaymentStatus string `json:"payment_status"` // "paid", "unpaid"
	Metadata      struct {
		InvoiceID string `json:"invoice_id"`
	} `json:"metadata"`
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

func (c *StripeClient) CreateCheckoutSession(ctx context.Context, invoiceID string, amount float64, currency string) (string, string, error) {
	amountInCents := int64(amount * 100)

	data := url.Values{}
	data.Set("mode", "payment")
	data.Set("payment_method_types[0]", "card")
	data.Set("line_items[0][price_data][currency]", strings.ToLower(currency))
	data.Set("line_items[0][price_data][product_data][name]", fmt.Sprintf("Shipping Fee (Invoice %s)", invoiceID))
	data.Set("line_items[0][price_data][unit_amount]", fmt.Sprintf("%d", amountInCents))
	data.Set("line_items[0][quantity]", "1")
	data.Set("success_url", "http://localhost:5173/?session_id={CHECKOUT_SESSION_ID}&payment_status=success")
	data.Set("cancel_url", "http://localhost:5173/?payment_status=failed")
	data.Set("metadata[invoice_id]", invoiceID)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("do request: %w", err)
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
			return "", "", fmt.Errorf("stripe api error: %s", errResp.Error.Message)
		}
		return "", "", fmt.Errorf("stripe api returned status %d", resp.StatusCode)
	}

	var sessionResp StripeCheckoutSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}

	return sessionResp.ID, sessionResp.URL, nil
}

func (c *StripeClient) RetrieveCheckoutSession(ctx context.Context, sessionID string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.stripe.com/v1/checkout/sessions/"+sessionID, nil)
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error.Message != "" {
			return "", "", fmt.Errorf("stripe api error: %s", errResp.Error.Message)
		}
		return "", "", fmt.Errorf("stripe api returned status %d", resp.StatusCode)
	}

	var sessionResp StripeCheckoutSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return "", "", fmt.Errorf("decode response: %w", err)
	}

	return sessionResp.PaymentStatus, sessionResp.Metadata.InvoiceID, nil
}
