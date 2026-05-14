package budget

import (
	"context"
	"time"

	budgetModel "github.com/davidsugianto/idp-core/internal/model/budget"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, budget *budgetModel.Budget) error
	GetByID(ctx context.Context, id string) (*budgetModel.Budget, error)
	ListByTeam(ctx context.Context, teamID string) ([]budgetModel.Budget, error)
	ListActive(ctx context.Context) ([]budgetModel.Budget, error)
	Update(ctx context.Context, budget *budgetModel.Budget) error
	Delete(ctx context.Context, id string) error

	CreateAlert(ctx context.Context, alert *budgetModel.BudgetAlert) error
	GetAlertsByBudget(ctx context.Context, budgetID string) ([]budgetModel.BudgetAlert, error)
	GetLatestAlertForThreshold(ctx context.Context, budgetID string, threshold int, periodStart time.Time) (*budgetModel.BudgetAlert, error)
}

type repository struct {
	db *gorm.DB
}

type Dependencies struct {
	Database *gorm.DB
}

func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}

func (r *repository) Create(ctx context.Context, budget *budgetModel.Budget) error {
	return r.db.WithContext(ctx).Create(budget).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*budgetModel.Budget, error) {
	var budget budgetModel.Budget
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&budget).Error
	if err != nil {
		return nil, err
	}
	return &budget, nil
}

func (r *repository) ListByTeam(ctx context.Context, teamID string) ([]budgetModel.Budget, error) {
	var budgets []budgetModel.Budget
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&budgets).Error
	return budgets, err
}

func (r *repository) ListActive(ctx context.Context) ([]budgetModel.Budget, error) {
	var budgets []budgetModel.Budget
	err := r.db.WithContext(ctx).
		Where("status = ?", budgetModel.StatusActive).
		Find(&budgets).Error
	return budgets, err
}

func (r *repository) Update(ctx context.Context, budget *budgetModel.Budget) error {
	return r.db.WithContext(ctx).Save(budget).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&budgetModel.Budget{}).Error
}

func (r *repository) CreateAlert(ctx context.Context, alert *budgetModel.BudgetAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *repository) GetAlertsByBudget(ctx context.Context, budgetID string) ([]budgetModel.BudgetAlert, error) {
	var alerts []budgetModel.BudgetAlert
	err := r.db.WithContext(ctx).
		Where("budget_id = ?", budgetID).
		Order("timestamp DESC").
		Find(&alerts).Error
	return alerts, err
}

func (r *repository) GetLatestAlertForThreshold(ctx context.Context, budgetID string, threshold int, periodStart time.Time) (*budgetModel.BudgetAlert, error) {
	var alert budgetModel.BudgetAlert
	err := r.db.WithContext(ctx).
		Where("budget_id = ? AND threshold = ? AND timestamp >= ?", budgetID, threshold, periodStart).
		Order("timestamp DESC").
		First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}