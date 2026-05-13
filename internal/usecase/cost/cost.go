package cost

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/cost"
	"github.com/davidsugianto/idp-core/internal/pkg/opencost"
	"github.com/google/uuid"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

// SyncCosts fetches cost data from OpenCost and persists it
func (u *usecase) SyncCosts(ctx context.Context) error {
	window := "1h"
	req := opencost.AllocationRequest{
		Window:    window,
		Aggregate: "namespace",
	}

	resp, err := u.opencostClient.GetAllocation(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to fetch opencost allocation: %w", err)
	}

	var records []cost.CostRecord
	for _, alloc := range resp.Data {
		start, _ := time.Parse(time.RFC3339, alloc.Start)
		end, _ := time.Parse(time.RFC3339, alloc.End)

		namespace := alloc.Name
		teamID := ""
		environmentID := ""
		if alloc.Properties != nil {
			if alloc.Properties.Namespace != "" {
				namespace = alloc.Properties.Namespace
			}
			if v, ok := alloc.Properties.Labels["team"]; ok {
				teamID = v
			}
			if v, ok := alloc.Properties.Labels["environment"]; ok {
				environmentID = v
			}
		}

		rawData, _ := json.Marshal(alloc.Raw)

		records = append(records, cost.CostRecord{
			ID:            uuid.New().String(),
			TeamID:        teamID,
			EnvironmentID: environmentID,
			Namespace:     namespace,
			PeriodStart:   start,
			PeriodEnd:     end,
			CPUCost:       alloc.CPUCost,
			RAMCost:       alloc.RAMCost,
			PVCost:        alloc.PVCost,
			NetworkCost:   alloc.NetworkCost,
			TotalCost:     alloc.TotalCost,
			RawData:       string(rawData),
		})
	}

	if len(records) == 0 {
		return nil
	}

	return u.repo.BatchCreate(ctx, records)
}

// List returns cost records with filtering and pagination
func (u *usecase) List(ctx context.Context, filter cost.CostFilter) (*cost.CostListResponse, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultLimit
	}
	if filter.Limit > maxLimit {
		filter.Limit = maxLimit
	}

	records, total, err := u.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return cost.ToCostListResponse(records, total), nil
}

// GetTeamCosts returns cost records for a specific team within a time range
func (u *usecase) GetTeamCosts(ctx context.Context, teamID, namespace, start, end string) (*cost.CostListResponse, error) {
	records, err := u.repo.GetByTeamAndPeriod(ctx, teamID, namespace, start, end)
	if err != nil {
		return nil, err
	}

	return cost.ToCostListResponse(records, int64(len(records))), nil
}