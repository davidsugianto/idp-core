package budget

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Period constants
const (
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
)

// Status constants
const (
	StatusActive = "active"
	StatusPaused = "paused"
)

// Alert status constants
const (
	AlertStatusSent   = "sent"
	AlertStatusFailed = "failed"
)

type Budget struct {
	ID              string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TeamID          string         `gorm:"type:varchar(36);not null;index" json:"team_id"`
	EnvironmentID   string         `gorm:"type:varchar(36)" json:"environment_id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	Limit           float64        `gorm:"type:numeric(12,4);not null" json:"limit"`
	Period          string         `gorm:"type:varchar(20);not null;default:monthly" json:"period"`
	AlertThresholds string         `gorm:"type:text;not null;default:'80,90,100'" json:"alert_thresholds"`
	AlertChannels   string         `gorm:"type:text;not null;default:'[\"slack\"]'" json:"alert_channels"`
	Status          string         `gorm:"type:varchar(20);not null;default:active" json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Budget) TableName() string { return "budgets" }

type BudgetAlert struct {
	ID           string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	BudgetID     string    `gorm:"type:varchar(36);not null;index" json:"budget_id"`
	Timestamp    time.Time `gorm:"not null;index" json:"timestamp"`
	Threshold    int       `gorm:"not null" json:"threshold"`
	CurrentSpend float64   `gorm:"type:numeric(12,4);not null;default:0" json:"current_spend"`
	Limit        float64   `gorm:"type:numeric(12,4);not null;default:0" json:"limit"`
	Percentage   float64   `gorm:"type:numeric(6,2);not null;default:0" json:"percentage"`
	SentTo       string    `gorm:"type:text;not null;default:'[]'" json:"sent_to"`
	Status       string    `gorm:"type:varchar(20);not null;default:sent" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func (BudgetAlert) TableName() string { return "budget_alerts" }

type CreateBudgetRequest struct {
	TeamID          string `json:"team_id" binding:"required"`
	EnvironmentID   string `json:"environment_id"`
	Name            string `json:"name" binding:"required"`
	Limit           float64 `json:"limit" binding:"required"`
	Period          string `json:"period"`
	AlertThresholds []int   `json:"alert_thresholds"`
	AlertChannels   []string `json:"alert_channels"`
}

type UpdateBudgetRequest struct {
	Name            *string   `json:"name"`
	Limit           *float64  `json:"limit"`
	Period          *string   `json:"period"`
	AlertThresholds *[]int    `json:"alert_thresholds"`
	AlertChannels   *[]string `json:"alert_channels"`
	Status          *string   `json:"status"`
}

type BudgetResponse struct {
	ID              string    `json:"id"`
	TeamID          string    `json:"team_id"`
	EnvironmentID   string    `json:"environment_id"`
	Name            string    `json:"name"`
	Limit           float64   `json:"limit"`
	Period          string    `json:"period"`
	AlertThresholds []int     `json:"alert_thresholds"`
	AlertChannels   []string  `json:"alert_channels"`
	Status          string    `json:"status"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

type BudgetListResponse struct {
	Budgets []BudgetResponse `json:"budgets"`
	Total   int64            `json:"total"`
}

type BudgetAlertResponse struct {
	ID           string  `json:"id"`
	BudgetID     string  `json:"budget_id"`
	Timestamp    string  `json:"timestamp"`
	Threshold    int     `json:"threshold"`
	CurrentSpend float64 `json:"current_spend"`
	Limit        float64 `json:"limit"`
	Percentage   float64 `json:"percentage"`
	SentTo       string  `json:"sent_to"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

func ToBudgetResponse(b *Budget) *BudgetResponse {
	return &BudgetResponse{
		ID:              b.ID,
		TeamID:          b.TeamID,
		EnvironmentID:   b.EnvironmentID,
		Name:            b.Name,
		Limit:           b.Limit,
		Period:          b.Period,
		AlertThresholds: ParseThresholds(b.AlertThresholds),
		AlertChannels:   ParseChannels(b.AlertChannels),
		Status:          b.Status,
		CreatedAt:       b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       b.UpdatedAt.Format(time.RFC3339),
	}
}

func ToBudgetListResponse(budgets []Budget, total int64) *BudgetListResponse {
	responses := make([]BudgetResponse, len(budgets))
	for i, b := range budgets {
		responses[i] = *ToBudgetResponse(&b)
	}
	return &BudgetListResponse{
		Budgets: responses,
		Total:   total,
	}
}

func ToBudgetAlertResponse(a *BudgetAlert) *BudgetAlertResponse {
	return &BudgetAlertResponse{
		ID:           a.ID,
		BudgetID:     a.BudgetID,
		Timestamp:    a.Timestamp.Format(time.RFC3339),
		Threshold:    a.Threshold,
		CurrentSpend: a.CurrentSpend,
		Limit:        a.Limit,
		Percentage:   a.Percentage,
		SentTo:       a.SentTo,
		Status:       a.Status,
		CreatedAt:    a.CreatedAt.Format(time.RFC3339),
	}
}

func ParseThresholds(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}

func FormatThresholds(thresholds []int) string {
	parts := make([]string, len(thresholds))
	for i, t := range thresholds {
		parts[i] = strconv.Itoa(t)
	}
	return strings.Join(parts, ",")
}

func ParseChannels(s string) []string {
	if s == "" {
		return nil
	}
	var channels []string
	if err := json.Unmarshal([]byte(s), &channels); err != nil {
		return nil
	}
	return channels
}

func FormatChannels(channels []string) string {
	data, _ := json.Marshal(channels)
	return string(data)
}

func ValidPeriod(period string) bool {
	switch period {
	case PeriodDaily, PeriodWeekly, PeriodMonthly:
		return true
	}
	return false
}