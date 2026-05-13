package kubecost

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllocation(t *testing.T) {
	t.Run("successful allocation response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/model/allocation", r.URL.Path)
			assert.Equal(t, "1h", r.URL.Query().Get("window"))
			assert.Equal(t, "namespace", r.URL.Query().Get("aggregate"))

			resp := AllocationResponse{
				Code: 200,
				Data: []AllocationData{
					{
						Name:        "team-a-dev",
						CPUCost:     10.5,
						RAMCost:     5.0,
						PVCost:      1.0,
						NetworkCost: 0.5,
						TotalCost:   17.0,
						Start:       "2026-05-13T00:00:00Z",
						End:         "2026-05-13T01:00:00Z",
						Properties: &AllocationProperties{
							Namespace: "team-a-dev",
							Labels: map[string]string{
								"team": "team-a",
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(Config{BaseURL: server.URL})
		req := AllocationRequest{
			Window:    "1h",
			Aggregate: "namespace",
		}

		resp, err := client.GetAllocation(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.Code)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "team-a-dev", resp.Data[0].Name)
		assert.Equal(t, 17.0, resp.Data[0].TotalCost)
		assert.NotNil(t, resp.Data[0].Raw)
	})

	t.Run("API key is sent as Bearer token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

			resp := AllocationResponse{Code: 200, Data: []AllocationData{}}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(Config{BaseURL: server.URL, APIKey: "test-api-key"})
		req := AllocationRequest{Window: "1h"}

		_, err := client.GetAllocation(context.Background(), req)
		assert.NoError(t, err)
	})

	t.Run("non-200 status returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized"))
		}))
		defer server.Close()

		client := NewClient(Config{BaseURL: server.URL})
		req := AllocationRequest{Window: "1h"}

		resp, err := client.GetAllocation(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "401")
	})

	t.Run("invalid base URL returns error", func(t *testing.T) {
		client := NewClient(Config{BaseURL: "://invalid-url"})
		req := AllocationRequest{Window: "1h"}

		resp, err := client.GetAllocation(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("connection refused returns error", func(t *testing.T) {
		client := NewClient(Config{BaseURL: "http://127.0.0.1:1"})
		req := AllocationRequest{Window: "1h"}

		resp, err := client.GetAllocation(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
