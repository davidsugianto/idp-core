package main

import (
	"net/http"

	"github.com/davidsugianto/go-pkgs/grace"
	httpHandler "github.com/davidsugianto/idp-core/internal/handler/http"
	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	apikeyUC "github.com/davidsugianto/idp-core/internal/usecase/apikey"
	auditlogUC "github.com/davidsugianto/idp-core/internal/usecase/auditlog"
	budgetUC "github.com/davidsugianto/idp-core/internal/usecase/budget"
	costUC "github.com/davidsugianto/idp-core/internal/usecase/cost"
	environmentUC "github.com/davidsugianto/idp-core/internal/usecase/environment"
	rightsizingUC "github.com/davidsugianto/idp-core/internal/usecase/rightsizing"
	quotaUC "github.com/davidsugianto/idp-core/internal/usecase/quota"
	serviceUC "github.com/davidsugianto/idp-core/internal/usecase/service"
	roleUC "github.com/davidsugianto/idp-core/internal/usecase/role"
	teamUC "github.com/davidsugianto/idp-core/internal/usecase/team"
	userUC "github.com/davidsugianto/idp-core/internal/usecase/user"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/davidsugianto/idp-core/docs/swagger" // swagger docs
)

type Server struct {
	*http.Server
	handler            *httpHandler.Handler
	config             *config.Config
	apiKeyUseCase      apikeyUC.Usecase
	auditLogUseCase    auditlogUC.Usecase
	budgetUseCase      budgetUC.Usecase
	costUseCase        costUC.Usecase
	rightsizingUseCase rightsizingUC.Usecase
	quotaUseCase       quotaUC.Usecase
	serviceUseCase     serviceUC.Usecase
}

type Dependencies struct {
	EnvironmentUseCase environmentUC.Usecase
	UserUseCase        userUC.Usecase
	TeamUseCase        teamUC.Usecase
	RoleUseCase        roleUC.Usecase
	ApiKeyUseCase      apikeyUC.Usecase
	AuditLogUseCase    auditlogUC.Usecase
	BudgetUseCase      budgetUC.Usecase
	CostUseCase        costUC.Usecase
	RightsizingUseCase rightsizingUC.Usecase
	QuotaUseCase       quotaUC.Usecase
	ServiceUseCase     serviceUC.Usecase
	Config             *config.Config
	WebhookValidator   *webhook.Validator
}

func New(deps Dependencies) *Server {
	return &Server{
		Server:             &http.Server{},
		apiKeyUseCase:      deps.ApiKeyUseCase,
		auditLogUseCase:    deps.AuditLogUseCase,
		budgetUseCase:      deps.BudgetUseCase,
		costUseCase:        deps.CostUseCase,
		rightsizingUseCase: deps.RightsizingUseCase,
		quotaUseCase:       deps.QuotaUseCase,
		serviceUseCase:     deps.ServiceUseCase,
		handler: httpHandler.New(httpHandler.Dependencies{
			EnvironmentUseCase: deps.EnvironmentUseCase,
			UserUseCase:        deps.UserUseCase,
			TeamUseCase:        deps.TeamUseCase,
			RoleUseCase:        deps.RoleUseCase,
			ApiKeyUseCase:      deps.ApiKeyUseCase,
			AuditLogUseCase:    deps.AuditLogUseCase,
			BudgetUseCase:      deps.BudgetUseCase,
			CostUseCase:        deps.CostUseCase,
			RightsizingUseCase: deps.RightsizingUseCase,
			QuotaUseCase:       deps.QuotaUseCase,
			ServiceUseCase:     deps.ServiceUseCase,
			AuthConfig:         &deps.Config.Auth,
			WebhookValidator:   deps.WebhookValidator,
		}),
		config: deps.Config,
	}
}

func (s *Server) setupPublicRoutes(r *gin.Engine) {
	// Swagger documentation endpoint (public)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check (public)
	r.GET("/ping", s.handler.Ping)
	r.GET("/health", s.handler.Ping)
	r.GET("/ready", s.handler.Ping)

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Auth routes (public)
	auth := r.Group("/auth")
	auth.POST("/login", s.handler.Login)

	// Admission webhook endpoint (public)
	admission := r.Group("/admission")
	admission.POST("/validate", s.handler.Validate)
}

func (s *Server) setupAPIRoutes(r *gin.Engine) {
	// API v1 routes (protected)
	v1 := r.Group("/v1")
	v1.Use(gin.Recovery(), middleware.RequestID(), middleware.AuditLog(s.auditLogUseCase))

	// Environment routes (protected with JWT)
	envs := v1.Group("/environments")
	envs.Use(middleware.JWT(&s.config.Auth))
	envs.GET("", s.handler.ListEnvironments)
	envs.GET("/:id", s.handler.GetEnvironment)
	envs.GET("/:id/status", s.handler.GetEnvironmentStatus)
	envs.GET("/:id/gitops/status", s.handler.GetGitOpsStatus)
	envs.GET("/:id/workloads", s.handler.GetWorkloads)
	envs.GET("/:id/workloads/:name", s.handler.GetWorkloadDetails)
	envs.POST("", s.handler.CreateEnvironment)
	envs.POST("/:id/sync", s.handler.SyncEnvironment)
	envs.DELETE("/:id", s.handler.DeleteEnvironment)

	// User routes (protected with JWT)
	users := v1.Group("/users")
	users.Use(middleware.JWT(&s.config.Auth))
	users.GET("", s.handler.ListUsers)
	users.POST("", s.handler.CreateUser)
	users.GET("/:id", s.handler.GetUser)
	users.PATCH("/:id", s.handler.UpdateUser)
	users.DELETE("/:id", s.handler.DeleteUser)
	users.PUT("/:id/status", s.handler.UpdateUserStatus)

	// Team routes (protected with JWT)
	teams := v1.Group("/teams")
	teams.Use(middleware.JWT(&s.config.Auth))
	teams.GET("", s.handler.ListTeams)
	teams.POST("", s.handler.CreateTeam)
	teams.GET("/:id", s.handler.GetTeam)
	teams.PATCH("/:id", s.handler.UpdateTeam)
	teams.DELETE("/:id", s.handler.DeleteTeam)
	teams.GET("/:id/members", s.handler.ListTeamMembers)
	teams.POST("/:id/members", s.handler.AddTeamMember)
	teams.PATCH("/:id/members/:userId", s.handler.UpdateTeamMember)
	teams.DELETE("/:id/members/:userId", s.handler.RemoveTeamMember)

	// Role routes (protected with JWT)
	roles := v1.Group("/roles")
	roles.Use(middleware.JWT(&s.config.Auth))
	roles.GET("", s.handler.ListRoles)
	roles.POST("", s.handler.CreateRole)
	roles.GET("/:id", s.handler.GetRole)
	roles.PATCH("/:id", s.handler.UpdateRole)
	roles.DELETE("/:id", s.handler.DeleteRole)
	roles.POST("/assign", s.handler.AssignRole)
	roles.POST("/revoke", s.handler.RevokeRole)

	// User roles routes (add to existing users group)
	users.GET("/:id/roles", s.handler.GetUserRoles)

	// Team member roles routes (add to existing teams group)
	teams.GET("/:id/members/:userId/roles", s.handler.GetUserTeamRoles)

	// API key routes (protected with JWT)
	apiKeys := v1.Group("/api-keys")
	apiKeys.Use(middleware.JWT(&s.config.Auth))
	apiKeys.GET("", s.handler.ListAPIKeys)
	apiKeys.POST("", s.handler.CreateAPIKey)
	apiKeys.GET("/:id", s.handler.GetAPIKey)
	apiKeys.PATCH("/:id", s.handler.UpdateAPIKey)
	apiKeys.DELETE("/:id", s.handler.DeleteAPIKey)

	// Audit log routes (protected with JWT)
	auditLogs := v1.Group("/audit-logs")
	auditLogs.Use(middleware.JWT(&s.config.Auth))
	auditLogs.GET("", s.handler.ListAuditLogs)
	auditLogs.GET("/:id", s.handler.GetAuditLog)

	// Cost routes (protected with JWT)
	costs := v1.Group("/costs")
	costs.Use(middleware.JWT(&s.config.Auth))
	costs.GET("", s.handler.ListCosts)
	costs.GET("/team", s.handler.GetTeamCosts)

	// Budget routes (protected with JWT)
	budgets := v1.Group("/budgets")
	budgets.Use(middleware.JWT(&s.config.Auth))
	budgets.GET("", s.handler.ListBudgets)
	budgets.POST("", s.handler.CreateBudget)
	budgets.GET("/:id", s.handler.GetBudget)
	budgets.PATCH("/:id", s.handler.UpdateBudget)
	budgets.DELETE("/:id", s.handler.DeleteBudget)
	budgets.GET("/:id/alerts", s.handler.ListBudgetAlerts)

	// Rightsizing routes (protected with JWT)
	rightsizing := v1.Group("/rightsizing")
	rightsizing.Use(middleware.JWT(&s.config.Auth))
	rightsizing.GET("/recommendations", s.handler.ListRightsizingRecommendations)
	rightsizing.GET("/recommendations/:id", s.handler.GetRightsizingRecommendation)
	rightsizing.POST("/recommendations/:id/apply", s.handler.ApplyRightsizingRecommendation)
	rightsizing.POST("/recommendations/:id/rollback", s.handler.RollbackRightsizingRecommendation)
	rightsizing.POST("/recommendations/:id/dismiss", s.handler.DismissRightsizingRecommendation)

	// Quota routes (protected with JWT)
	quotas := v1.Group("/quotas")
	quotas.Use(middleware.JWT(&s.config.Auth))
	quotas.GET("", s.handler.ListResourceQuotas)
	quotas.POST("", s.handler.CreateResourceQuota)
	quotas.GET("/:id", s.handler.GetResourceQuota)
	quotas.PATCH("/:id", s.handler.UpdateResourceQuota)
	quotas.DELETE("/:id", s.handler.DeleteResourceQuota)
	quotas.GET("/namespace/:namespace", s.handler.GetResourceQuotaByNamespace)
	quotas.GET("/namespace/:namespace/usage", s.handler.GetNamespaceUsage)
	quotas.POST("/namespace/:namespace/usage/refresh", s.handler.RefreshNamespaceUsage)
	quotas.POST("/check", s.handler.CheckQuota)

	// Service routes (protected with JWT)
	services := v1.Group("/services")
	services.Use(middleware.JWT(&s.config.Auth))
	services.GET("", s.handler.ListServices)
	services.POST("", s.handler.CreateService)
	services.GET("/discover", s.handler.DiscoverServices)
	services.GET("/:id", s.handler.GetService)
	services.PATCH("/:id", s.handler.UpdateService)
	services.DELETE("/:id", s.handler.DeleteService)
	services.GET("/:id/versions", s.handler.ListServiceVersions)
	services.POST("/:id/versions", s.handler.CreateServiceVersion)
	services.GET("/:id/versions/:versionId", s.handler.GetServiceVersion)
	services.PATCH("/:id/versions/:versionId", s.handler.UpdateServiceVersion)
	services.GET("/:id/versions/:versionId/endpoints", s.handler.ListServiceEndpoints)
	services.POST("/:id/versions/:versionId/endpoints", s.handler.CreateServiceEndpoint)
	services.GET("/:id/versions/:versionId/endpoints/:endpointId", s.handler.GetServiceEndpoint)
	services.PATCH("/:id/versions/:versionId/endpoints/:endpointId", s.handler.UpdateServiceEndpoint)
	services.DELETE("/:id/versions/:versionId/endpoints/:endpointId", s.handler.DeleteServiceEndpoint)
}

func (s *Server) Run(port string) error {
	r := gin.Default()

	corsConfig := cors.Config{
		AllowOrigins:     s.config.CORS.AllowedOrigins,
		AllowMethods:     s.config.CORS.AllowedMethods,
		AllowHeaders:     s.config.CORS.AllowedHeaders,
		AllowCredentials: s.config.CORS.AllowCredentials,
	}
	r.Use(cors.New(corsConfig))

	// Setup routes
	s.setupPublicRoutes(r)
	s.setupAPIRoutes(r)

	s.Addr = port
	s.Handler = r

	return grace.ServeHTTP(s.Addr, s.Handler)
}
