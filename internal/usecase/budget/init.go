package budget

import (
	"context"

	budgetModel "github.com/davidsugianto/idp-core/internal/model/budget"
	budgetRepo "github.com/davidsugianto/idp-core/internal/repository/budget"
	costRepo "github.com/davidsugianto/idp-core/internal/repository/cost"
)

type Usecase interface {
	Create(ctx context.Context, req budgetModel.CreateBudgetRequest) (*budgetModel.BudgetResponse, error)
	Get(ctx context.Context, id string) (*budgetModel.BudgetResponse, error)
	List(ctx context.Context, teamID string) (*budgetModel.BudgetListResponse, error)
	Update(ctx context.Context, id string, req budgetModel.UpdateBudgetRequest) (*budgetModel.BudgetResponse, error)
	Delete(ctx context.Context, id string) error
	GetAlerts(ctx context.Context, budgetID string) ([]budgetModel.BudgetAlertResponse, error)
	CheckAlerts(ctx context.Context) error
}

type SlackNotifier interface {
	SendAlert(ctx context.Context, channel string, title string, fields map[string]string) error
	Channel() string
}

type usecase struct {
	budgetRepo budgetRepo.Repository
	costRepo   costRepo.Repository
	slack      SlackNotifier
}

type Dependencies struct {
	BudgetRepo    budgetRepo.Repository
	CostRepo      costRepo.Repository
	SlackNotifier SlackNotifier
}

func New(deps Dependencies) Usecase {
	return &usecase{
		budgetRepo: deps.BudgetRepo,
		costRepo:   deps.CostRepo,
		slack:      deps.SlackNotifier,
	}
}