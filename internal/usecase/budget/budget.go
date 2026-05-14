package budget

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/davidsugianto/go-pkgs/logs"
	"github.com/google/uuid"

	"github.com/davidsugianto/idp-core/internal/model/budget"
	"github.com/davidsugianto/idp-core/internal/model/cost"
)

var (
	ErrBudgetNotFound   = errors.New("budget not found")
	ErrBudgetNameRequired = errors.New("budget name is required")
	ErrBudgetInvalidPeriod = errors.New("budget period must be daily, weekly, or monthly")
	ErrBudgetLimitPositive = errors.New("budget limit must be greater than zero")
)

func (u *usecase) Create(ctx context.Context, req budget.CreateBudgetRequest) (*budget.BudgetResponse, error) {
	if req.Name == "" {
		return nil, ErrBudgetNameRequired
	}
	if req.Limit <= 0 {
		return nil, ErrBudgetLimitPositive
	}
	period := req.Period
	if period == "" {
		period = budget.PeriodMonthly
	}
	if !budget.ValidPeriod(period) {
		return nil, ErrBudgetInvalidPeriod
	}

	thresholds := req.AlertThresholds
	if len(thresholds) == 0 {
		thresholds = []int{80, 90, 100}
	}
	channels := req.AlertChannels
	if len(channels) == 0 {
		channels = []string{"slack"}
	}

	b := &budget.Budget{
		ID:              uuid.New().String(),
		TeamID:          req.TeamID,
		EnvironmentID:   req.EnvironmentID,
		Name:            req.Name,
		Limit:           req.Limit,
		Period:          period,
		AlertThresholds: budget.FormatThresholds(thresholds),
		AlertChannels:   budget.FormatChannels(channels),
		Status:          budget.StatusActive,
	}

	if err := u.budgetRepo.Create(ctx, b); err != nil {
		return nil, err
	}

	return budget.ToBudgetResponse(b), nil
}

func (u *usecase) Get(ctx context.Context, id string) (*budget.BudgetResponse, error) {
	b, err := u.budgetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, ErrBudgetNotFound
	}
	return budget.ToBudgetResponse(b), nil
}

func (u *usecase) List(ctx context.Context, teamID string) (*budget.BudgetListResponse, error) {
	budgets, err := u.budgetRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return budget.ToBudgetListResponse(budgets, int64(len(budgets))), nil
}

func (u *usecase) Update(ctx context.Context, id string, req budget.UpdateBudgetRequest) (*budget.BudgetResponse, error) {
	b, err := u.budgetRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, ErrBudgetNotFound
	}

	if req.Name != nil {
		b.Name = *req.Name
	}
	if req.Limit != nil {
		if *req.Limit <= 0 {
			return nil, ErrBudgetLimitPositive
		}
		b.Limit = *req.Limit
	}
	if req.Period != nil {
		if !budget.ValidPeriod(*req.Period) {
			return nil, ErrBudgetInvalidPeriod
		}
		b.Period = *req.Period
	}
	if req.AlertThresholds != nil {
		b.AlertThresholds = budget.FormatThresholds(*req.AlertThresholds)
	}
	if req.AlertChannels != nil {
		b.AlertChannels = budget.FormatChannels(*req.AlertChannels)
	}
	if req.Status != nil {
		b.Status = *req.Status
	}

	if err := u.budgetRepo.Update(ctx, b); err != nil {
		return nil, err
	}

	return budget.ToBudgetResponse(b), nil
}

func (u *usecase) Delete(ctx context.Context, id string) error {
	b, err := u.budgetRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if b == nil {
		return ErrBudgetNotFound
	}
	return u.budgetRepo.Delete(ctx, id)
}

func (u *usecase) GetAlerts(ctx context.Context, budgetID string) ([]budget.BudgetAlertResponse, error) {
	alerts, err := u.budgetRepo.GetAlertsByBudget(ctx, budgetID)
	if err != nil {
		return nil, err
	}
	responses := make([]budget.BudgetAlertResponse, len(alerts))
	for i, a := range alerts {
		responses[i] = *budget.ToBudgetAlertResponse(&a)
	}
	return responses, nil
}

func (u *usecase) CheckAlerts(ctx context.Context) error {
	activeBudgets, err := u.budgetRepo.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to list active budgets: %w", err)
	}

	now := time.Now().UTC()

	for _, b := range activeBudgets {
		if b.Limit == 0 {
			continue
		}

		periodStart, _ := getPeriodWindow(b.Period, now)

		filter := cost.CostFilter{
			TeamID:    b.TeamID,
			StartDate: &periodStart,
			EndDate:   &now,
		}
		if b.EnvironmentID != "" {
			filter.EnvironmentID = b.EnvironmentID
		}

		records, _, err := u.costRepo.List(ctx, filter)
		if err != nil {
			logs.Errorf("BudgetAlertCheck: failed to query costs for budget %s: %v", b.ID, err)
			continue
		}

		var currentSpend float64
		for _, r := range records {
			currentSpend += r.TotalCost
		}

		percentage := (currentSpend / b.Limit) * 100

		thresholds := budget.ParseThresholds(b.AlertThresholds)
		sort.Ints(thresholds)

		for _, threshold := range thresholds {
			if percentage < float64(threshold) {
				continue
			}

			existing, err := u.budgetRepo.GetLatestAlertForThreshold(ctx, b.ID, threshold, periodStart)
			if err == nil && existing != nil {
				continue
			}

			channel := u.slack.Channel()
			title := fmt.Sprintf("[Budget Alert] *%s* has reached %d%% of its %s budget limit", b.Name, threshold, b.Period)
			fields := map[string]string{
				"Current Spend": fmt.Sprintf("$%.2f", currentSpend),
				"Budget Limit":  fmt.Sprintf("$%.2f", b.Limit),
				"Percentage":    fmt.Sprintf("%.1f%%", percentage),
				"Period":        fmt.Sprintf("%s - %s", periodStart.Format("2006-01-02"), now.Format("2006-01-02")),
			}

			alertStatus := budget.AlertStatusSent
			if err := u.slack.SendAlert(ctx, channel, title, fields); err != nil {
				logs.Errorf("BudgetAlertCheck: failed to send Slack alert for budget %s: %v", b.ID, err)
				alertStatus = budget.AlertStatusFailed
			}

			alert := &budget.BudgetAlert{
				ID:           uuid.New().String(),
				BudgetID:     b.ID,
				Timestamp:    now,
				Threshold:    threshold,
				CurrentSpend: currentSpend,
				Limit:        b.Limit,
				Percentage:   percentage,
				SentTo:       channel,
				Status:       alertStatus,
			}
			if err := u.budgetRepo.CreateAlert(ctx, alert); err != nil {
				logs.Errorf("BudgetAlertCheck: failed to record alert for budget %s: %v", b.ID, err)
			}
		}
	}

	return nil
}

func getPeriodWindow(period string, now time.Time) (time.Time, time.Time) {
	switch period {
	case budget.PeriodDaily:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return start, now
	case budget.PeriodWeekly:
		weekday := now.Weekday()
		daysSinceMonday := int(weekday) - 1
		if daysSinceMonday < 0 {
			daysSinceMonday = 6
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-daysSinceMonday, 0, 0, 0, 0, time.UTC)
		return start, now
	default:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return start, now
	}
}