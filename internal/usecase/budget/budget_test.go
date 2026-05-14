package budget

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/model/budget"
	"github.com/davidsugianto/idp-core/internal/model/cost"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBudgetRepository(ctrl)
	mockSlack := mocks.NewMockSlackNotifier(ctrl)
	uc := New(Dependencies{
		BudgetRepo:    mockRepo,
		CostRepo:      nil,
		SlackNotifier: mockSlack,
	})

	t.Run("creates budget with all fields", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := uc.Create(context.Background(), budget.CreateBudgetRequest{
			TeamID:          "team-1",
			EnvironmentID:   "env-1",
			Name:            "My Budget",
			Limit:           1000.0,
			Period:          budget.PeriodMonthly,
			AlertThresholds: []int{50, 75, 100},
			AlertChannels:   []string{"slack", "email"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "My Budget", resp.Name)
		assert.Equal(t, float64(1000.0), resp.Limit)
		assert.Equal(t, budget.PeriodMonthly, resp.Period)
		assert.Equal(t, budget.StatusActive, resp.Status)
		assert.Len(t, resp.AlertThresholds, 3)
		assert.Len(t, resp.AlertChannels, 2)
		assert.NotEmpty(t, resp.ID)
	})

	t.Run("uses defaults when optional fields are empty", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := uc.Create(context.Background(), budget.CreateBudgetRequest{
			TeamID: "team-1",
			Name:   "Default Budget",
			Limit:  500.0,
		})
		assert.NoError(t, err)
		assert.Equal(t, budget.PeriodMonthly, resp.Period)
		assert.Equal(t, []int{80, 90, 100}, resp.AlertThresholds)
		assert.Equal(t, []string{"slack"}, resp.AlertChannels)
	})

	t.Run("rejects empty name", func(t *testing.T) {
		resp, err := uc.Create(context.Background(), budget.CreateBudgetRequest{
			TeamID: "team-1",
			Limit:  100.0,
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrBudgetNameRequired)
	})

	t.Run("rejects non-positive limit", func(t *testing.T) {
		resp, err := uc.Create(context.Background(), budget.CreateBudgetRequest{
			TeamID: "team-1",
			Name:   "Bad Budget",
			Limit:  0,
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrBudgetLimitPositive)
	})

	t.Run("rejects invalid period", func(t *testing.T) {
		resp, err := uc.Create(context.Background(), budget.CreateBudgetRequest{
			TeamID: "team-1",
			Name:   "Bad Budget",
			Limit:  100.0,
			Period: "yearly",
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrBudgetInvalidPeriod)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("db error"))

		resp, err := uc.Create(context.Background(), budget.CreateBudgetRequest{
			TeamID: "team-1",
			Name:   "Budget",
			Limit:  100.0,
		})
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "db error")
	})
}

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBudgetRepository(ctrl)
	uc := New(Dependencies{BudgetRepo: mockRepo})

	t.Run("returns budget by id", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "budget-1").
			Return(&budget.Budget{
				ID:     "budget-1",
				Name:   "Test Budget",
				Limit:  500.0,
				Period: budget.PeriodDaily,
				Status: budget.StatusActive,
			}, nil)

		resp, err := uc.Get(context.Background(), "budget-1")
		assert.NoError(t, err)
		assert.Equal(t, "budget-1", resp.ID)
		assert.Equal(t, "Test Budget", resp.Name)
	})

	t.Run("returns not found for nil budget", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, nil)

		resp, err := uc.Get(context.Background(), "nonexistent")
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrBudgetNotFound)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "budget-1").
			Return(nil, errors.New("connection refused"))

		resp, err := uc.Get(context.Background(), "budget-1")
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "connection refused")
	})
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBudgetRepository(ctrl)
	uc := New(Dependencies{BudgetRepo: mockRepo})

	t.Run("returns budgets for team", func(t *testing.T) {
		mockRepo.EXPECT().
			ListByTeam(gomock.Any(), "team-1").
			Return([]budget.Budget{
				{ID: "b-1", Name: "Budget 1", Limit: 100.0, Period: budget.PeriodMonthly, Status: budget.StatusActive},
				{ID: "b-2", Name: "Budget 2", Limit: 200.0, Period: budget.PeriodDaily, Status: budget.StatusPaused},
			}, nil)

		resp, err := uc.List(context.Background(), "team-1")
		assert.NoError(t, err)
		assert.Len(t, resp.Budgets, 2)
		assert.Equal(t, int64(2), resp.Total)
	})

	t.Run("returns empty list for team with no budgets", func(t *testing.T) {
		mockRepo.EXPECT().
			ListByTeam(gomock.Any(), "team-empty").
			Return([]budget.Budget{}, nil)

		resp, err := uc.List(context.Background(), "team-empty")
		assert.NoError(t, err)
		assert.Empty(t, resp.Budgets)
		assert.Equal(t, int64(0), resp.Total)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			ListByTeam(gomock.Any(), "team-1").
			Return(nil, errors.New("db error"))

		resp, err := uc.List(context.Background(), "team-1")
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "db error")
	})
}

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBudgetRepository(ctrl)
	uc := New(Dependencies{BudgetRepo: mockRepo})

	existing := &budget.Budget{
		ID:     "budget-1",
		Name:   "Original",
		Limit:  100.0,
		Period: budget.PeriodMonthly,
		Status: budget.StatusActive,
	}

	t.Run("updates name", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "budget-1").
			Return(existing, nil)
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		newName := "Updated Name"
		resp, err := uc.Update(context.Background(), "budget-1", budget.UpdateBudgetRequest{
			Name: &newName,
		})
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", resp.Name)
	})

	t.Run("rejects non-positive limit", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "budget-1").
			Return(existing, nil)

		badLimit := 0.0
		resp, err := uc.Update(context.Background(), "budget-1", budget.UpdateBudgetRequest{
			Limit: &badLimit,
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrBudgetLimitPositive)
	})

	t.Run("rejects invalid period", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "budget-1").
			Return(existing, nil)

		badPeriod := "annually"
		resp, err := uc.Update(context.Background(), "budget-1", budget.UpdateBudgetRequest{
			Period: &badPeriod,
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrBudgetInvalidPeriod)
	})

	t.Run("returns not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, nil)

		newName := "New"
		resp, err := uc.Update(context.Background(), "nonexistent", budget.UpdateBudgetRequest{
			Name: &newName,
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrBudgetNotFound)
	})
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBudgetRepository(ctrl)
	uc := New(Dependencies{BudgetRepo: mockRepo})

	t.Run("deletes budget", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "budget-1").
			Return(&budget.Budget{ID: "budget-1"}, nil)
		mockRepo.EXPECT().
			Delete(gomock.Any(), "budget-1").
			Return(nil)

		err := uc.Delete(context.Background(), "budget-1")
		assert.NoError(t, err)
	})

	t.Run("returns not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, nil)

		err := uc.Delete(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, ErrBudgetNotFound)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "budget-1").
			Return(&budget.Budget{ID: "budget-1"}, nil)
		mockRepo.EXPECT().
			Delete(gomock.Any(), "budget-1").
			Return(errors.New("db error"))

		err := uc.Delete(context.Background(), "budget-1")
		assert.ErrorContains(t, err, "db error")
	})
}

func TestGetAlerts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBudgetRepository(ctrl)
	uc := New(Dependencies{BudgetRepo: mockRepo})

	t.Run("returns alerts for budget", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetAlertsByBudget(gomock.Any(), "budget-1").
			Return([]budget.BudgetAlert{
				{ID: "alert-1", BudgetID: "budget-1", Threshold: 80, Status: budget.AlertStatusSent, Timestamp: now},
				{ID: "alert-2", BudgetID: "budget-1", Threshold: 90, Status: budget.AlertStatusSent, Timestamp: now},
			}, nil)

		resp, err := uc.GetAlerts(context.Background(), "budget-1")
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
	})

	t.Run("returns empty list", func(t *testing.T) {
		mockRepo.EXPECT().
			GetAlertsByBudget(gomock.Any(), "budget-1").
			Return([]budget.BudgetAlert{}, nil)

		resp, err := uc.GetAlerts(context.Background(), "budget-1")
		assert.NoError(t, err)
		assert.Empty(t, resp)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetAlertsByBudget(gomock.Any(), "budget-1").
			Return(nil, errors.New("db error"))

		resp, err := uc.GetAlerts(context.Background(), "budget-1")
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "db error")
	})
}

func TestCheckAlerts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBudgetRepo := mocks.NewMockBudgetRepository(ctrl)
	mockCostRepo := mocks.NewMockCostRepository(ctrl)
	mockSlack := mocks.NewMockSlackNotifier(ctrl)
	uc := New(Dependencies{
		BudgetRepo:    mockBudgetRepo,
		CostRepo:      mockCostRepo,
		SlackNotifier: mockSlack,
	})

	t.Run("no active budgets", func(t *testing.T) {
		mockBudgetRepo.EXPECT().
			ListActive(gomock.Any()).
			Return([]budget.Budget{}, nil)

		err := uc.CheckAlerts(context.Background())
		assert.NoError(t, err)
	})

	t.Run("active budget with spend below all thresholds", func(t *testing.T) {
		mockBudgetRepo.EXPECT().
			ListActive(gomock.Any()).
			Return([]budget.Budget{
				{
					ID:              "budget-1",
					TeamID:          "team-1",
					Name:            "Low Spend Budget",
					Limit:           1000.0,
					Period:          budget.PeriodMonthly,
					AlertThresholds: "80,90,100",
					AlertChannels:   `["slack"]`,
					Status:          budget.StatusActive,
				},
			}, nil)

		mockCostRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]cost.CostRecord{
				{ID: "cost-1", TotalCost: 100.0},
			}, int64(1), nil)

		err := uc.CheckAlerts(context.Background())
		assert.NoError(t, err)
	})

	t.Run("sends alert when spend exceeds threshold", func(t *testing.T) {
		mockBudgetRepo.EXPECT().
			ListActive(gomock.Any()).
			Return([]budget.Budget{
				{
					ID:              "budget-2",
					TeamID:          "team-1",
					Name:            "High Spend Budget",
					Limit:           1000.0,
					Period:          budget.PeriodMonthly,
					AlertThresholds: "80,90,100",
					AlertChannels:   `["slack"]`,
					Status:          budget.StatusActive,
				},
			}, nil)

		mockCostRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]cost.CostRecord{
				{ID: "cost-1", TotalCost: 850.0},
			}, int64(1), nil)

		mockBudgetRepo.EXPECT().
			GetLatestAlertForThreshold(gomock.Any(), "budget-2", 80, gomock.Any()).
			Return(nil, errors.New("not found"))

		mockSlack.EXPECT().
			Channel().
			Return("#alerts")

		mockSlack.EXPECT().
			SendAlert(gomock.Any(), "#alerts", gomock.Any(), gomock.Any()).
			Return(nil)

		mockBudgetRepo.EXPECT().
			CreateAlert(gomock.Any(), gomock.Any()).
			Return(nil)

		err := uc.CheckAlerts(context.Background())
		assert.NoError(t, err)
	})

	t.Run("skips alert when already sent for threshold in period", func(t *testing.T) {
		mockBudgetRepo.EXPECT().
			ListActive(gomock.Any()).
			Return([]budget.Budget{
				{
					ID:              "budget-3",
					TeamID:          "team-1",
					Name:            "Already Alerted",
					Limit:           1000.0,
					Period:          budget.PeriodMonthly,
					AlertThresholds: "80,90,100",
					Status:          budget.StatusActive,
				},
			}, nil)

		mockCostRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]cost.CostRecord{
				{ID: "cost-1", TotalCost: 850.0},
			}, int64(1), nil)

		mockBudgetRepo.EXPECT().
			GetLatestAlertForThreshold(gomock.Any(), "budget-3", 80, gomock.Any()).
			Return(&budget.BudgetAlert{ID: "existing-alert"}, nil)

		err := uc.CheckAlerts(context.Background())
		assert.NoError(t, err)
	})

	t.Run("skips budget with zero limit", func(t *testing.T) {
		mockBudgetRepo.EXPECT().
			ListActive(gomock.Any()).
			Return([]budget.Budget{
				{
					ID:              "budget-4",
					TeamID:          "team-1",
					Name:            "Zero Limit",
					Limit:           0,
					Period:          budget.PeriodMonthly,
					AlertThresholds: "80,90,100",
					Status:          budget.StatusActive,
				},
			}, nil)

		err := uc.CheckAlerts(context.Background())
		assert.NoError(t, err)
	})

	t.Run("continues on cost query error", func(t *testing.T) {
		mockBudgetRepo.EXPECT().
			ListActive(gomock.Any()).
			Return([]budget.Budget{
				{
					ID:              "budget-5",
					TeamID:          "team-1",
					Name:            "Error Budget",
					Limit:           1000.0,
					Period:          budget.PeriodMonthly,
					AlertThresholds: "80,90,100",
					Status:          budget.StatusActive,
				},
			}, nil)

		mockCostRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return(nil, int64(0), errors.New("query failed"))

		err := uc.CheckAlerts(context.Background())
		assert.NoError(t, err)
	})

	t.Run("propagates ListActive error", func(t *testing.T) {
		mockBudgetRepo.EXPECT().
			ListActive(gomock.Any()).
			Return(nil, errors.New("db error"))

		err := uc.CheckAlerts(context.Background())
		assert.ErrorContains(t, err, "failed to list active budgets")
	})
}

func TestGetPeriodWindow(t *testing.T) {
	t.Run("daily period", func(t *testing.T) {
		now := time.Date(2026, 5, 14, 15, 30, 0, 0, time.UTC)
		start, end := getPeriodWindow(budget.PeriodDaily, now)
		assert.Equal(t, time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC), start)
		assert.Equal(t, now, end)
	})

	t.Run("weekly period - Thursday", func(t *testing.T) {
		now := time.Date(2026, 5, 14, 15, 30, 0, 0, time.UTC)
		start, end := getPeriodWindow(budget.PeriodWeekly, now)
		assert.Equal(t, time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC), start)
		assert.Equal(t, now, end)
	})

	t.Run("weekly period - Sunday", func(t *testing.T) {
		now := time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)
		start, end := getPeriodWindow(budget.PeriodWeekly, now)
		assert.Equal(t, time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC), start)
		assert.Equal(t, now, end)
	})

	t.Run("monthly period", func(t *testing.T) {
		now := time.Date(2026, 5, 14, 15, 30, 0, 0, time.UTC)
		start, end := getPeriodWindow(budget.PeriodMonthly, now)
		assert.Equal(t, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), start)
		assert.Equal(t, now, end)
	})
}