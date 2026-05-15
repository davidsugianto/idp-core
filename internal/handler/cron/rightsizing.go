package cron

import (
	"context"

	"github.com/davidsugianto/go-pkgs/logs"
)

func (h *Handler) RightsizingGenerate(ctx context.Context) error {
	err := h.rightsizingUseCase.GenerateRecommendations(ctx)
	if err != nil {
		logs.Errorf("RightsizingGenerate failed: %v", err)
		return err
	}
	logs.Info("RightsizingGenerate completed successfully")
	return nil
}
