package kubecost

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Config holds the configuration for the Kubecost client
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// Client is an HTTP client for the Kubecost API
type Client struct {
	httpClient *http.Client
	config     Config
}

// NewClient creates a new Kubecost API client
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		config:     cfg,
	}
}

// GetAllocation fetches cost allocation data from the Kubecost API
func (c *Client) GetAllocation(ctx context.Context, req AllocationRequest) (*AllocationResponse, error) {
	u, err := url.Parse(c.config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	u = u.JoinPath("model", "allocation")

	q := u.Query()
	q.Set("window", req.Window)
	if req.Aggregate != "" {
		q.Set("aggregate", req.Aggregate)
	}
	if req.Step != "" {
		q.Set("step", req.Step)
	}
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("kubecost API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kubecost API returned status %d: %s", resp.StatusCode, string(body))
	}

	var allocResp AllocationResponse
	if err := json.Unmarshal(body, &allocResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Preserve raw data for each allocation
	for i := range allocResp.Data {
		var raw map[string]any
		if err := json.Unmarshal(body, &raw); err == nil {
			if data, ok := raw["data"].([]any); ok && i < len(data) {
				if item, ok := data[i].(map[string]any); ok {
					allocResp.Data[i].Raw = item
				}
			}
		}
	}

	return &allocResp, nil
}
