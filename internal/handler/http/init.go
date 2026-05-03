package http

import (
	"net/http"

	"github.com/davidsugianto/idp-core/internal/usecase/environment"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	environmentUseCase environment.Usecase
}

type Dependencies struct {
	EnvironmentUseCase environment.Usecase
}

func New(deps Dependencies) *Handler {
	return &Handler{
		environmentUseCase: deps.EnvironmentUseCase,
	}
}

// Ping godoc
// @Summary Health check
// @Description Check if the API is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /v1/ping [get]
func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
