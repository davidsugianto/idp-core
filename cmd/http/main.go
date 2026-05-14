package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/davidsugianto/idp-core/internal/model/cost"
	"github.com/davidsugianto/idp-core/internal/model/team"
	"github.com/davidsugianto/idp-core/internal/model/user"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/opencost"
	"github.com/davidsugianto/idp-core/internal/pkg/prometheus"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"

	"github.com/davidsugianto/go-pkgs/db"

	apikeyRepo "github.com/davidsugianto/idp-core/internal/repository/apikey"
	auditlogRepo "github.com/davidsugianto/idp-core/internal/repository/auditlog"
	costRepo "github.com/davidsugianto/idp-core/internal/repository/cost"
	envRepository "github.com/davidsugianto/idp-core/internal/repository/environment"
	permissionRepo "github.com/davidsugianto/idp-core/internal/repository/permission"
	roleRepo "github.com/davidsugianto/idp-core/internal/repository/role"
	teamRepository "github.com/davidsugianto/idp-core/internal/repository/team"
	userRepository "github.com/davidsugianto/idp-core/internal/repository/user"
	apikeyUsecase "github.com/davidsugianto/idp-core/internal/usecase/apikey"
	auditlogUsecase "github.com/davidsugianto/idp-core/internal/usecase/auditlog"
	costUsecase "github.com/davidsugianto/idp-core/internal/usecase/cost"
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
	roleUsecase "github.com/davidsugianto/idp-core/internal/usecase/role"
	teamUsecase "github.com/davidsugianto/idp-core/internal/usecase/team"
	userUsecase "github.com/davidsugianto/idp-core/internal/usecase/user"

	"github.com/davidsugianto/go-pkgs/logs"
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

var (
	errLogPath   string
	infoLogPath  string
	debugLogPath string
)

func main() {
	// Logs
	flag.StringVar(&errLogPath, "error_log", "log/idp-core.error.log", "error log")
	flag.StringVar(&infoLogPath, "info_log", "log/idp-core.info.log", "info log")
	flag.StringVar(&debugLogPath, "debug_log", "log/idp-core.debug.log", "debug log")
	flag.Parse()

	setLog(logs.ErrorLevel, errLogPath)
	setLog(logs.InfoLevel, infoLogPath)
	setLog(logs.DebugLevel, debugLogPath)

	logs.Info("Starting IDP Core API server")

	// Config
	cfgPath := fmt.Sprintf("configs/config.%s.yaml", os.Getenv("APP_ENV"))
	cfg, err := config.Load(cfgPath)
	if err != nil {
		logs.Fatalf("cannot load config from %s", cfgPath)
		panic(err)
	}

	// DB
	ctx := context.Background()
	dbConfig := db.NewConfig(db.Postgres, cfg.Database.Host, cfg.Database.Name).
		WithPort(cfg.Database.Port).
		WithCredentials(cfg.Database.User, cfg.Database.Password)

	dbClientWrapper, err := db.New(ctx, dbConfig)
	if err != nil {
		logs.Fatal("cannot connect to database")
	}
	dbClient := dbClientWrapper.DB

	// Auto-migrate Phase 2 tables
	if err := dbClient.AutoMigrate(&user.User{}, &team.Team{}, &team.TeamMember{}, &cost.CostRecord{}); err != nil {
		logs.Fatal("cannot migrate database")
	}

	// FinOps clients
	opencostClient := opencost.NewClient(opencost.Config{
		BaseURL: cfg.FinOps.OpenCost.BaseURL,
	})
	_ = prometheus.NewClient(prometheus.Config{
		URL: cfg.FinOps.Prometheus.URL,
	})

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
	costRepo := costRepo.New(costRepo.Dependencies{
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

	costUC := costUsecase.New(costUsecase.Dependencies{
		Repo:           costRepo,
		OpenCostClient: opencostClient,
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
		CostUseCase:        costUC,
		Config:             cfg,
		WebhookValidator:   webhookValidator,
	})

	logs.Info("listening on port")
	if err := server.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		logs.Fatal("server failed to run")
	}
}

func setLog(level logs.Level, filePath string) {
	lgr, err := logs.NewLogger(&logs.Config{
		Level:   level,
		LogFile: filePath,
		Caller:  false,
		AppName: "idp-core - API",
		UseJSON: false,
	})
	if err != nil {
		logs.Fatalln(err)
	}

	err = logs.SetLogger(level, lgr)
	if err != nil {
		logs.Fatalln(err)
	}
}
