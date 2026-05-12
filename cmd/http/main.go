package main

import (
	"context"
	"fmt"

	"github.com/davidsugianto/idp-core/internal/model/team"
	"github.com/davidsugianto/idp-core/internal/model/user"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"

	"github.com/davidsugianto/go-pkgs/db"
	"github.com/davidsugianto/go-pkgs/logger"

	apikeyRepo "github.com/davidsugianto/idp-core/internal/repository/apikey"
	auditlogRepo "github.com/davidsugianto/idp-core/internal/repository/auditlog"
	envRepository "github.com/davidsugianto/idp-core/internal/repository/environment"
	permissionRepo "github.com/davidsugianto/idp-core/internal/repository/permission"
	roleRepo "github.com/davidsugianto/idp-core/internal/repository/role"
	teamRepository "github.com/davidsugianto/idp-core/internal/repository/team"
	userRepository "github.com/davidsugianto/idp-core/internal/repository/user"
	apikeyUsecase "github.com/davidsugianto/idp-core/internal/usecase/apikey"
	auditlogUsecase "github.com/davidsugianto/idp-core/internal/usecase/auditlog"
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
	roleUsecase "github.com/davidsugianto/idp-core/internal/usecase/role"
	teamUsecase "github.com/davidsugianto/idp-core/internal/usecase/team"
	userUsecase "github.com/davidsugianto/idp-core/internal/usecase/user"
)

// @title IDP Core API
// @version 1.0
// @description Internal Developer Platform API for self-provisioning Kubernetes environments
// @termsOfService http://swagger.io/terms/

// @contact.name Platform Engineering Team
// @contact.email platform@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8989
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// Logger
	logs := logger.NewWithConfig(logger.Config{
		ServiceName: "idp-core",
		Environment: "development",
		Format:      logger.FormatJSON,
	})
	logs.Info().Msg("Starting IDP API server")

	// Config
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		logs.Fatal().Err(err).Msg("cannot load config")
		panic(err)
	}

	// DB
	ctx := context.Background()
	dbConfig := db.NewConfig(db.Postgres, cfg.Database.Host, cfg.Database.Name).
		WithPort(cfg.Database.Port).
		WithCredentials(cfg.Database.User, cfg.Database.Password)

	dbClientWrapper, err := db.New(ctx, dbConfig)
	if err != nil {
		logs.Fatal().Err(err).Msg("cannot connect to database")
	}
	dbClient := dbClientWrapper.DB

	// Auto-migrate Phase 2 tables
	if err := dbClient.AutoMigrate(&user.User{}, &team.Team{}, &team.TeamMember{}); err != nil {
		logs.Fatal().Err(err).Msg("cannot migrate database")
	}

	// Repositories
	envRepo := envRepository.New(envRepository.Dependencies{
		Database: dbClient,
	})
	userRepo := userRepository.New(userRepository.Dependencies{
		Database: dbClient,
	})
	teamRepo := teamRepository.New(teamRepository.Dependencies{
		Database: dbClient,
	})
	roleRepo := roleRepo.New(roleRepo.Dependencies{
		Database: dbClient,
	})
	permRepo := permissionRepo.New(permissionRepo.Dependencies{
		Database: dbClient,
	})
	apiKeyRepo := apikeyRepo.New(apikeyRepo.Dependencies{
		Database: dbClient,
	})
	auditLogRepo := auditlogRepo.New(auditlogRepo.Dependencies{
		Database: dbClient,
	})

	// UseCases
	envUC := envUsecase.New(envUsecase.Dependencies{
		EnvironmentRepo: envRepo,
	})
	userUC := userUsecase.New(userUsecase.Dependencies{
		UserRepo: userRepo,
	})
	teamUC := teamUsecase.New(teamUsecase.Dependencies{
		TeamRepo: teamRepo,
		UserRepo: userRepo,
	})
	roleUC := roleUsecase.New(roleUsecase.Dependencies{
		RoleRepo:       roleRepo,
		PermissionRepo: permRepo,
	})
	apiKeyUC := apikeyUsecase.New(apikeyUsecase.Dependencies{
		APIKeyRepo: apiKeyRepo,
	})
	auditLogUC := auditlogUsecase.New(auditlogUsecase.Dependencies{
		AuditLogRepo: auditLogRepo,
	})

	// Webhook validator
	webhookValidator := webhook.NewValidator()

	server := New(Dependencies{
		EnvironmentUseCase: envUC,
		UserUseCase:        userUC,
		TeamUseCase:        teamUC,
		RoleUseCase:        roleUC,
		ApiKeyUseCase:      apiKeyUC,
		AuditLogUseCase:    auditLogUC,
		Config:             cfg,
		Logger:             logs,
		WebhookValidator:   webhookValidator,
	})

	logs.Info().Int("port", cfg.Server.Port).Msg("listening on port")
	if err := server.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		logs.Fatal().Err(err).Msg("server failed to run")
	}
}
