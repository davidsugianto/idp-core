package cron

import (
	"context"

	"github.com/davidsugianto/go-pkgs/logs"
)

func (h *Handler) BudgetAlertCheck(ctx context.Context) error {
	err := h.budgetUseCase.CheckAlerts(ctx)
	if err != nil {
		logs.Errorf("BudgetAlertCheck failed: %v", err)
		return err
	}
	logs.Info("BudgetAlertCheck completed successfully")
	return nil
}