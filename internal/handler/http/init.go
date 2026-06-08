package http

import (
	"github.com/davidsugianto/go-pkgs/response"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	oidcPkg "github.com/davidsugianto/idp-core/internal/pkg/oidc"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	apikeyUsecase "github.com/davidsugianto/idp-core/internal/usecase/apikey"
	auditlogUsecase "github.com/davidsugianto/idp-core/internal/usecase/auditlog"
	budgetUsecase "github.com/davidsugianto/idp-core/internal/usecase/budget"
	costUsecase "github.com/davidsugianto/idp-core/internal/usecase/cost"
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
	quotaUsecase "github.com/davidsugianto/idp-core/internal/usecase/quota"
	rightsizingUsecase "github.com/davidsugianto/idp-core/internal/usecase/rightsizing"
	roleUsecase "github.com/davidsugianto/idp-core/internal/usecase/role"
	serviceUsecase "github.com/davidsugianto/idp-core/internal/usecase/service"
	templateUsecase "github.com/davidsugianto/idp-core/internal/usecase/template"
	teamUsecase "github.com/davidsugianto/idp-core/internal/usecase/team"
	userUsecase "github.com/davidsugianto/idp-core/internal/usecase/user"
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
	budgetUseCase      budgetUsecase.Usecase
	rightsizingUseCase rightsizingUsecase.Usecase
	quotaUseCase       quotaUsecase.Usecase
	serviceUseCase     serviceUsecase.Usecase
	templateUseCase    templateUsecase.Usecase
	authConfig         *config.AuthConfig
	webhookValidator   *webhook.Validator
	oidcClient         *oidcPkg.Client
	oidcVerifier       *oidcPkg.Verifier
	oidcEndSessionURL  string
	allowedOrigins     []string
}

type Dependencies struct {
	EnvironmentUseCase envUsecase.Usecase
	UserUseCase        userUsecase.Usecase
	TeamUseCase        teamUsecase.Usecase
	RoleUseCase        roleUsecase.Usecase
	ApiKeyUseCase      apikeyUsecase.Usecase
	AuditLogUseCase    auditlogUsecase.Usecase
	CostUseCase        costUsecase.Usecase
	BudgetUseCase      budgetUsecase.Usecase
	RightsizingUseCase rightsizingUsecase.Usecase
	QuotaUseCase       quotaUsecase.Usecase
	ServiceUseCase     serviceUsecase.Usecase
	TemplateUseCase    templateUsecase.Usecase
	AuthConfig         *config.AuthConfig
	WebhookValidator   *webhook.Validator
	OIDCClient         *oidcPkg.Client
	OIDCVerifier       *oidcPkg.Verifier
	OIDCEndSessionURL  string
	AllowedOrigins     []string
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
		budgetUseCase:      deps.BudgetUseCase,
		rightsizingUseCase: deps.RightsizingUseCase,
		quotaUseCase:       deps.QuotaUseCase,
		serviceUseCase:     deps.ServiceUseCase,
		templateUseCase:    deps.TemplateUseCase,
		authConfig:         deps.AuthConfig,
		webhookValidator:   deps.WebhookValidator,
		oidcClient:         deps.OIDCClient,
		oidcVerifier:       deps.OIDCVerifier,
		oidcEndSessionURL:  deps.OIDCEndSessionURL,
		allowedOrigins:     deps.AllowedOrigins,
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
