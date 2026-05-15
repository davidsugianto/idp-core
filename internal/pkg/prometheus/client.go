package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Config holds the configuration for the Prometheus client
type Config struct {
	URL string
}

// Client is an HTTP client for Prometheus queries
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

// QueryResult represents a single Prometheus query result value with its metric labels
type QueryResult struct {
	Metric   map[string]string
	Value    float64
	UnixTime int64
}

// instantResponse is the JSON response for instant queries
type instantResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"` // [timestamp, value]
		} `json:"result"`
	} `json:"data"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

// Query executes an instant PromQL query and returns all result values
func (c *Client) Query(ctx context.Context, query string) ([]QueryResult, error) {
	if c.config.URL == "" {
		return nil, fmt.Errorf("prometheus URL not configured")
	}

	endpoint := fmt.Sprintf("%s/api/v1/query", c.config.URL)
	params := url.Values{}
	params.Set("query", query)

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	var result instantResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s - %s", result.ErrorType, result.Error)
	}

	var results []QueryResult
	for _, r := range result.Data.Result {
		if len(r.Value) >= 2 {
			// Value is [timestamp, value] where value is a string
			var timestamp float64
			var value float64
			var parseErr error

			if ts, ok := r.Value[0].(float64); ok {
				timestamp = ts
			}
			if val, ok := r.Value[1].(string); ok {
				value, parseErr = strconv.ParseFloat(val, 64)
			} else if val, ok := r.Value[1].(float64); ok {
				value = val
			}

			if parseErr == nil {
				results = append(results, QueryResult{
					Metric:   r.Metric,
					Value:    value,
					UnixTime: int64(timestamp),
				})
			}
		}
	}

	return results, nil
}

// QuerySingle is a convenience method for queries expected to return a single value
func (c *Client) QuerySingle(ctx context.Context, query string) (float64, error) {
	results, err := c.Query(ctx, query)
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, fmt.Errorf("no results returned for query: %s", query)
	}
	return results[0].Value, nil
}

// QueryByMetric executes a query and returns a map of metric label values to result values
// Useful for queries that return multiple series (e.g., per-container metrics)
func (c *Client) QueryByMetric(ctx context.Context, query string, labelKey string) (map[string]float64, error) {
	results, err := c.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	metricMap := make(map[string]float64)
	for _, r := range results {
		if labelValue, ok := r.Metric[labelKey]; ok {
			metricMap[labelValue] = r.Value
		}
	}

	return metricMap, nil
}

// RangeQueryResult represents a range query result with multiple values over time
type RangeQueryResult struct {
	Metric map[string]string
	Values []struct {
		UnixTime int64
		Value    float64
	}
}

// rangeResponse is the JSON response for range queries
type rangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"` // [[timestamp, value], ...]
		} `json:"result"`
	} `json:"data"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

// QueryRange executes a range PromQL query over a time period
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]RangeQueryResult, error) {
	if c.config.URL == "" {
		return nil, fmt.Errorf("prometheus URL not configured")
	}

	endpoint := fmt.Sprintf("%s/api/v1/query_range", c.config.URL)
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", start.Unix()))
	params.Set("end", fmt.Sprintf("%d", end.Unix()))
	params.Set("step", step.String())

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute range query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus returned status %d", resp.StatusCode)
	}

	var result rangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s - %s", result.ErrorType, result.Error)
	}

	var results []RangeQueryResult
	for _, r := range result.Data.Result {
		var rr RangeQueryResult
		rr.Metric = r.Metric
		for _, v := range r.Values {
			if len(v) >= 2 {
				var timestamp float64
				var value float64
				var parseErr error

				if ts, ok := v[0].(float64); ok {
					timestamp = ts
				}
				if val, ok := v[1].(string); ok {
					value, parseErr = strconv.ParseFloat(val, 64)
				} else if val, ok := v[1].(float64); ok {
					value = val
				}

				if parseErr == nil {
					rr.Values = append(rr.Values, struct {
						UnixTime int64
						Value    float64
					}{
						UnixTime: int64(timestamp),
						Value:    value,
					})
				}
			}
		}
		results = append(results, rr)
	}

	return results, nil
}
