package cron

import (
	"context"

	"github.com/davidsugianto/go-pkgs/logs"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	buildApplicationUsecase "github.com/davidsugianto/idp-core/internal/usecase/build_application"
	budgetUsecase "github.com/davidsugianto/idp-core/internal/usecase/budget"
	costUsecase "github.com/davidsugianto/idp-core/internal/usecase/cost"
	rightsizingUsecase "github.com/davidsugianto/idp-core/internal/usecase/rightsizing"
)

type Handler struct {
	costUseCase            costUsecase.Usecase
	budgetUseCase          budgetUsecase.Usecase
	rightsizingUseCase     rightsizingUsecase.Usecase
	buildApplicationUseCase buildApplicationUsecase.Usecase
	authConfig             *config.AuthConfig
	webhookValidator       *webhook.Validator
}

type Dependencies struct {
	CostUseCase            costUsecase.Usecase
	BudgetUseCase          budgetUsecase.Usecase
	RightsizingUseCase     rightsizingUsecase.Usecase
	BuildApplicationUseCase buildApplicationUsecase.Usecase
	AuthConfig             *config.AuthConfig
	WebhookValidator       *webhook.Validator
}

func New(deps Dependencies) *Handler {
	return &Handler{
		costUseCase:            deps.CostUseCase,
		budgetUseCase:          deps.BudgetUseCase,
		rightsizingUseCase:     deps.RightsizingUseCase,
		buildApplicationUseCase: deps.BuildApplicationUseCase,
		authConfig:             deps.AuthConfig,
		webhookValidator:       deps.WebhookValidator,
	}
}

func (h *Handler) Ping(ctx context.Context) error {
	logs.Info("ok")
	return nil
}
