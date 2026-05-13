package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Config holds the configuration for the Prometheus client
type Config struct {
	URL string
}

// Client is a stub HTTP client for Prometheus queries
type Client struct {
	httpClient *http.Client
	config     Config
}

// NewClient creates a new Prometheus client
func NewClient(cfg Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		config:     cfg,
	}
}

// Query executes an instant PromQL query and returns the result value.
// This is a stub for future use (e.g., rightsizing in Week 6).
func (c *Client) Query(ctx context.Context, query string) (float64, error) {
	if c.config.URL == "" {
		return 0, fmt.Errorf("prometheus URL not configured")
	}
	return 0, fmt.Errorf("prometheus query not yet implemented")
}