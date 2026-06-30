package cron

import (
	"context"

	"github.com/davidsugianto/go-pkgs/logs"
)

func (h *Handler) BuildExecutorDispatch(ctx context.Context) error {
	if h.buildApplicationUseCase == nil {
		return nil
	}

	err := h.buildApplicationUseCase.DrainQueuedBuilds(ctx)
	if err != nil {
		logs.Errorf("BuildExecutorDispatch failed: %v", err)
		return err
	}
	logs.Info("BuildExecutorDispatch completed successfully")
	return nil
}
