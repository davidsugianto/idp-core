package main

import (
	"net/http"

	"github.com/davidsugianto/go-pkgs/grace"
	extlogger "github.com/davidsugianto/go-pkgs/logger"
	httpHandler "github.com/davidsugianto/idp-core/internal/handler/http"
	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	apikeyUC "github.com/davidsugianto/idp-core/internal/usecase/apikey"
	auditlogUC "github.com/davidsugianto/idp-core/internal/usecase/auditlog"
	environmentUC "github.com/davidsugianto/idp-core/internal/usecase/environment"
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
	handler         *httpHandler.Handler
	config          *config.Config
	logger          *extlogger.Logger
	apiKeyUseCase   apikeyUC.Usecase
	auditLogUseCase auditlogUC.Usecase
}

type Dependencies struct {
	EnvironmentUseCase environmentUC.Usecase
	UserUseCase        userUC.Usecase
	TeamUseCase        teamUC.Usecase
	RoleUseCase        roleUC.Usecase
	ApiKeyUseCase      apikeyUC.Usecase
	AuditLogUseCase    auditlogUC.Usecase
	Config             *config.Config
	Logger             *extlogger.Logger
	WebhookValidator   *webhook.Validator
}

func New(deps Dependencies) *Server {
	return &Server{
		Server:          &http.Server{},
		apiKeyUseCase:   deps.ApiKeyUseCase,
		auditLogUseCase: deps.AuditLogUseCase,
		handler: httpHandler.New(httpHandler.Dependencies{
			EnvironmentUseCase: deps.EnvironmentUseCase,
			UserUseCase:        deps.UserUseCase,
			TeamUseCase:        deps.TeamUseCase,
			RoleUseCase:        deps.RoleUseCase,
			ApiKeyUseCase:      deps.ApiKeyUseCase,
			AuditLogUseCase:    deps.AuditLogUseCase,
			AuthConfig:         &deps.Config.Auth,
			WebhookValidator:   deps.WebhookValidator,
		}),
		config: deps.Config,
		logger: deps.Logger,
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
	v1.Use(gin.Recovery(), middleware.RequestID(), middleware.Logger(s.logger), middleware.AuditLog(s.auditLogUseCase))

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
