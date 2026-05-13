package http

import (
	"github.com/davidsugianto/go-pkgs/response"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
	roleUsecase "github.com/davidsugianto/idp-core/internal/usecase/role"
	teamUsecase "github.com/davidsugianto/idp-core/internal/usecase/team"
	userUsecase "github.com/davidsugianto/idp-core/internal/usecase/user"
	apikeyUsecase "github.com/davidsugianto/idp-core/internal/usecase/apikey"
	auditlogUsecase "github.com/davidsugianto/idp-core/internal/usecase/auditlog"
	costUsecase "github.com/davidsugianto/idp-core/internal/usecase/cost"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	environmentUseCase envUsecase.Usecase
	userUseCase        userUsecase.Usecase
	teamUseCase        teamUsecase.Usecase
	roleUseCase        roleUsecase.Usecase
	apiKeyUseCase      apikeyUsecase.Usecase
	auditLogUseCase    auditlogUsecase.Usecase
	costUseCase        costUsecase.Usecase
	authConfig         *config.AuthConfig
	webhookValidator   *webhook.Validator
}

type Dependencies struct {
	EnvironmentUseCase envUsecase.Usecase
	UserUseCase        userUsecase.Usecase
	TeamUseCase        teamUsecase.Usecase
	RoleUseCase        roleUsecase.Usecase
	ApiKeyUseCase      apikeyUsecase.Usecase
	AuditLogUseCase    auditlogUsecase.Usecase
	CostUseCase        costUsecase.Usecase
	AuthConfig         *config.AuthConfig
	WebhookValidator   *webhook.Validator
}

func New(deps Dependencies) *Handler {
	return &Handler{
		environmentUseCase: deps.EnvironmentUseCase,
		userUseCase:        deps.UserUseCase,
		teamUseCase:        deps.TeamUseCase,
		roleUseCase:        deps.RoleUseCase,
		apiKeyUseCase:      deps.ApiKeyUseCase,
		auditLogUseCase:    deps.AuditLogUseCase,
		costUseCase:        deps.CostUseCase,
		authConfig:         deps.AuthConfig,
		webhookValidator:   deps.WebhookValidator,
	}
}

// Ping godoc
// @Summary Health check
// @Description Check if the API is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ping [get]
func (h *Handler) Ping(c *gin.Context) {
	response.GinSuccess(c, gin.H{"status": "ok"})
}
