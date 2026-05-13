package cost

import (
	"context"
	"errors"
	"testing"

	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/model/cost"
	"github.com/davidsugianto/idp-core/internal/pkg/opencost"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSyncCosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCostRepository(ctrl)
	mockOpenCost := mocks.NewMockOpenCostClient(ctrl)

	uc := New(Dependencies{
		Repo:           mockRepo,
		OpenCostClient: mockOpenCost,
	})

	t.Run("successful sync with data", func(t *testing.T) {
		mockOpenCost.EXPECT().
			GetAllocation(gomock.Any(), gomock.Any()).
			Return(&opencost.AllocationResponse{
				Code: 200,
				Data: []opencost.AllocationData{
					{
						Name:        "team-a-dev",
						CPUCost:     10.5,
						RAMCost:     5.0,
						PVCost:      1.0,
						NetworkCost: 0.5,
						TotalCost:   17.0,
						Start:       "2026-05-13T00:00:00Z",
						End:         "2026-05-13T01:00:00Z",
						Properties: &opencost.AllocationProperties{
							Namespace: "team-a-dev",
							Labels: map[string]string{
								"team":        "team-a",
								"environment": "env-dev",
							},
						},
					},
				},
			}, nil)

		mockRepo.EXPECT().
			BatchCreate(gomock.Any(), gomock.Any()).
			Return(nil)

		err := uc.SyncCosts(context.Background())
		assert.NoError(t, err)
	})

	t.Run("successful sync with empty data", func(t *testing.T) {
		mockOpenCost.EXPECT().
			GetAllocation(gomock.Any(), gomock.Any()).
			Return(&opencost.AllocationResponse{
				Code: 200,
				Data: []opencost.AllocationData{},
			}, nil)

		err := uc.SyncCosts(context.Background())
		assert.NoError(t, err)
	})

	t.Run("opencost API error propagates", func(t *testing.T) {
		mockOpenCost.EXPECT().
			GetAllocation(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("connection refused"))

		err := uc.SyncCosts(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("batch create error propagates", func(t *testing.T) {
		mockOpenCost.EXPECT().
			GetAllocation(gomock.Any(), gomock.Any()).
			Return(&opencost.AllocationResponse{
				Code: 200,
				Data: []opencost.AllocationData{
					{
						Name:      "ns-1",
						TotalCost: 10.0,
						Start:     "2026-05-13T00:00:00Z",
						End:       "2026-05-13T01:00:00Z",
					},
				},
			}, nil)

		mockRepo.EXPECT().
			BatchCreate(gomock.Any(), gomock.Any()).
			Return(errors.New("db write error"))

		err := uc.SyncCosts(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db write error")
	})
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCostRepository(ctrl)
	uc := New(Dependencies{Repo: mockRepo})

	t.Run("list with filters", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]cost.CostRecord{
				{ID: "cost-1", Namespace: "team-a-dev", TotalCost: 150.0},
				{ID: "cost-2", Namespace: "team-b-prod", TotalCost: 300.0},
			}, int64(2), nil)

		filter := cost.CostFilter{
			TeamID: "team-a",
			Limit:  10,
		}

		resp, err := uc.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, resp.CostRecords, 2)
		assert.Equal(t, int64(2), resp.Total)
	})

	t.Run("enforces default limit", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]cost.CostRecord{}, int64(0), nil)

		filter := cost.CostFilter{Limit: 0}
		_, err := uc.List(context.Background(), filter)
		assert.NoError(t, err)
	})

	t.Run("caps max limit", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]cost.CostRecord{}, int64(0), nil)

		filter := cost.CostFilter{Limit: 500}
		_, err := uc.List(context.Background(), filter)
		assert.NoError(t, err)
	})

	t.Run("empty result", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]cost.CostRecord{}, int64(0), nil)

		resp, err := uc.List(context.Background(), cost.CostFilter{})
		assert.NoError(t, err)
		assert.Empty(t, resp.CostRecords)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return(nil, int64(0), errors.New("db error"))

		_, err := uc.List(context.Background(), cost.CostFilter{Limit: 10})
		assert.Error(t, err)
	})
}

func TestGetTeamCosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockCostRepository(ctrl)
	uc := New(Dependencies{Repo: mockRepo})

	t.Run("returns team costs", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByTeamAndPeriod(gomock.Any(), "team-a", "", "2026-05-01", "2026-05-13").
			Return([]cost.CostRecord{
				{ID: "cost-1", TeamID: "team-a", Namespace: "team-a-dev", TotalCost: 100.0},
			}, nil)

		resp, err := uc.GetTeamCosts(context.Background(), "team-a", "", "2026-05-01", "2026-05-13")
		assert.NoError(t, err)
		assert.Len(t, resp.CostRecords, 1)
		assert.Equal(t, int64(1), resp.Total)
	})

	t.Run("empty result for unknown team", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByTeamAndPeriod(gomock.Any(), "unknown", "", "", "").
			Return([]cost.CostRecord{}, nil)

		resp, err := uc.GetTeamCosts(context.Background(), "unknown", "", "", "")
		assert.NoError(t, err)
		assert.Empty(t, resp.CostRecords)
	})
}