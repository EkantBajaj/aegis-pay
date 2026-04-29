package routing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ChargeRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	UserID   string  `json:"user_id"`
}

type ChargeResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ProviderClient struct {
	BaseURL string
	Name    string
	Breaker *ProviderBreaker
	HTTP    *http.Client
}

func NewProviderClient(name, baseURL string) *ProviderClient {
	return &ProviderClient{
		Name:    name,
		BaseURL: baseURL,
		Breaker: NewProviderBreaker(name),
		HTTP: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Charge executes a payment request wrapped in a circuit breaker
func (p *ProviderClient) Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error) {
	// Wrap the HTTP call in the breaker
	result, err := p.Breaker.Execute(func() (interface{}, error) {
		return p.doRequest(ctx, req)
	})

	if err != nil {
		return nil, err
	}

	return result.(*ChargeResponse), nil
}

func (p *ProviderClient) doRequest(ctx context.Context, req ChargeRequest) (*ChargeResponse, error) {
	body, _ := json.Marshal(req)
	
	// Create request with context for timeout support
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("provider returned error: %d", resp.StatusCode)
	}

	var chargeResp ChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chargeResp, nil
}
