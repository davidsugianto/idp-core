package main

import (
	"net/http"

	"github.com/davidsugianto/go-pkgs/grace"
	extlogger "github.com/davidsugianto/go-pkgs/logger"
	httpHandler "github.com/davidsugianto/idp-core/internal/handler/http"
	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	environmentUC "github.com/davidsugianto/idp-core/internal/usecase/environment"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	*http.Server
	handler        *httpHandler.Handler
	authHandler    *httpHandler.AuthHandler
	webhookHandler *httpHandler.WebhookHandler
	config         *config.Config
	logger         *extlogger.Logger
}

type Dependencies struct {
	EnvironmentUseCase environmentUC.Usecase
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
		config:         deps.Config,
		logger:         deps.Logger,
	}
}

func (s *Server) v1Endpoint(r *gin.Engine) {
	g := r.Group("/v1")
	g.Use(gin.Recovery(), middleware.RequestID(), middleware.Logger(s.logger))

	// Swagger documentation endpoint
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// health check (public)
	g.GET("/ping", s.handler.Ping)

	// auth routes (public)
	auth := g.Group("/auth")
	auth.POST("/login", s.authHandler.Login)

	// environment routes (protected)
	envs := g.Group("/environments")
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

	// Metrics endpoint
	g.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Admission webhook endpoint
	admission := g.Group("/admission")
	admission.POST("/validate", s.webhookHandler.Validate)
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

	s.v1Endpoint(r)

	s.Addr = port
	s.Handler = r

	return grace.ServeHTTP(s.Addr, s.Handler)
}
