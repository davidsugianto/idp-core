package cron

import (
	"context"

	logs "github.com/davidsugianto/go-pkgs/logger"
)

func (h *Handler) FinopsSyncCosts(ctx context.Context) error {
	err := h.costUseCase.SyncCosts(ctx)
	if err != nil {
		logs.Error().Msgf("FinopsSyncCosts failed: %v", err)
		return err
	}
	logs.Info().Msg("FinopsSyncCosts completed successfully")
	return nil
}
