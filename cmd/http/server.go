package main

import (
	"net/http"

	"github.com/davidsugianto/go-pkgs/grace"
	extlogger "github.com/davidsugianto/go-pkgs/logger"
	httpHandler "github.com/davidsugianto/idp-core/internal/handler/http"
	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	environmentUC "github.com/davidsugianto/idp-core/internal/usecase/environment"
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
	handler        *httpHandler.Handler
	authHandler    *httpHandler.AuthHandler
	webhookHandler *httpHandler.WebhookHandler
	userHandler    *httpHandler.UserHandler
	teamHandler    *httpHandler.TeamHandler
	config         *config.Config
	logger         *extlogger.Logger
}

type Dependencies struct {
	EnvironmentUseCase environmentUC.Usecase
	UserUseCase        userUC.Usecase
	TeamUseCase        teamUC.Usecase
	Config             *config.Config
	Logger             *extlogger.Logger
}

func New(deps Dependencies) *Server {
	return &Server{
		Server: &http.Server{},
		handler: httpHandler.New(httpHandler.Dependencies{
			EnvironmentUseCase: deps.EnvironmentUseCase,
		}),
		authHandler:    httpHandler.NewAuthHandler(&deps.Config.Auth),
		webhookHandler: httpHandler.NewWebhookHandler(),
		userHandler:    httpHandler.NewUserHandler(deps.UserUseCase),
		teamHandler:    httpHandler.NewTeamHandler(deps.TeamUseCase),
		config:         deps.Config,
		logger:         deps.Logger,
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
	auth.POST("/login", s.authHandler.Login)

	// Admission webhook endpoint (public)
	admission := r.Group("/admission")
	admission.POST("/validate", s.webhookHandler.Validate)
}

func (s *Server) setupAPIRoutes(r *gin.Engine) {
	// API v1 routes (protected)
	v1 := r.Group("/v1")
	v1.Use(gin.Recovery(), middleware.RequestID(), middleware.Logger(s.logger))

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
	users.GET("", s.userHandler.ListUsers)
	users.POST("", s.userHandler.CreateUser)
	users.GET("/:id", s.userHandler.GetUser)
	users.PATCH("/:id", s.userHandler.UpdateUser)
	users.DELETE("/:id", s.userHandler.DeleteUser)
	users.PUT("/:id/status", s.userHandler.UpdateUserStatus)

	// Team routes (protected with JWT)
	teams := v1.Group("/teams")
	teams.Use(middleware.JWT(&s.config.Auth))
	teams.GET("", s.teamHandler.ListTeams)
	teams.POST("", s.teamHandler.CreateTeam)
	teams.GET("/:id", s.teamHandler.GetTeam)
	teams.PATCH("/:id", s.teamHandler.UpdateTeam)
	teams.DELETE("/:id", s.teamHandler.DeleteTeam)
	teams.GET("/:id/members", s.teamHandler.ListTeamMembers)
	teams.POST("/:id/members", s.teamHandler.AddTeamMember)
	teams.PATCH("/:id/members/:userId", s.teamHandler.UpdateTeamMember)
	teams.DELETE("/:id/members/:userId", s.teamHandler.RemoveTeamMember)
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
