package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/davidsugianto/go-pkgs/db"
	"github.com/davidsugianto/go-pkgs/logs"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/opencost"
	"github.com/davidsugianto/idp-core/internal/pkg/prometheus"
	"github.com/davidsugianto/idp-core/internal/pkg/redislock"
	"github.com/davidsugianto/idp-core/internal/pkg/slack"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	budgetRepo "github.com/davidsugianto/idp-core/internal/repository/budget"
	costRepo "github.com/davidsugianto/idp-core/internal/repository/cost"
	budgetUsecase "github.com/davidsugianto/idp-core/internal/usecase/budget"
	costUsecase "github.com/davidsugianto/idp-core/internal/usecase/cost"
	"github.com/go-redis/redis/v8"
)

var (
	errLogPath   string
	infoLogPath  string
	debugLogPath string
)

func main() {
	ctx := context.Background()

	// Logs
	flag.StringVar(&errLogPath, "error_log", "log/idp-core-cron.error.log", "error log")
	flag.StringVar(&infoLogPath, "info_log", "log/idp-core-cron.info.log", "info log")
	flag.StringVar(&debugLogPath, "debug_log", "log/idp-core-cron.debug.log", "debug log")
	flag.Parse()

	setLog(logs.ErrorLevel, errLogPath)
	setLog(logs.InfoLevel, infoLogPath)
	setLog(logs.DebugLevel, debugLogPath)

	logs.Info("Starting IDP Core CRON server")

	// Config
	cfgPath := fmt.Sprintf("configs/config.%s.yaml", os.Getenv("APP_ENV"))
	cfg, err := config.Load(cfgPath)
	if err != nil {
		logs.Fatal(fmt.Sprintf("cannot load config from %s", cfgPath))
		panic(err)
	}

	// DB
	dbConfig := db.NewConfig(db.Postgres, cfg.Database.Host, cfg.Database.Name).
		WithPort(cfg.Database.Port).
		WithCredentials(cfg.Database.User, cfg.Database.Password)

	dbClientWrapper, err := db.New(ctx, dbConfig)
	if err != nil {
		logs.Fatal("cannot connect to database")
	}
	dbClient := dbClientWrapper.DB

	// Redis
	// Redis Lock
	drd := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:       cfg.Redis.MasterName,
		SentinelAddrs:    []string{cfg.Redis.Address},
		SentinelPassword: cfg.Redis.Password,
		Password:         cfg.Redis.Password,
		DialTimeout:      5 * time.Second,
	})
	optGoredis := redislock.RedisDriver{
		GoRedisClient: []redis.UniversalClient{
			drd,
		},
	}
	// distributed lock redis
	distlock := redislock.New(optGoredis)

	// FinOps clients
	opencostClient := opencost.NewClient(opencost.Config{
		BaseURL: cfg.FinOps.OpenCost.BaseURL,
	})
	_ = prometheus.NewClient(prometheus.Config{
		URL: cfg.FinOps.Prometheus.URL,
	})

	// Webhook validator
	webhookValidator := webhook.NewValidator()

	// Slack client
	slackClient := slack.NewClient(cfg.Slack.WebhookURL, cfg.Slack.Channel)

	// Repositories
	costRepo := costRepo.New(costRepo.Dependencies{
		Database: dbClient,
	})
	budgetRepo := budgetRepo.New(budgetRepo.Dependencies{
		Database: dbClient,
	})

	// UseCases
	costUC := costUsecase.New(costUsecase.Dependencies{
		Repo:           costRepo,
		OpenCostClient: opencostClient,
	})
	budgetUC := budgetUsecase.New(budgetUsecase.Dependencies{
		BudgetRepo:    budgetRepo,
		CostRepo:      costRepo,
		SlackNotifier: slackClient,
	})

	// Server
	server := New(Dependencies{
		Schedules:        cfg.Cron.Schedules,
		Port:             cfg.Cron.Port,
		CostUseCase:      costUC,
		BudgetUseCase:    budgetUC,
		Config:           cfg,
		Distlock:         distlock,
		WebhookValidator: webhookValidator,
	})
	server.Run(ctx, cfg.Cron.GraceTimeout*time.Second)
}

func setLog(level logs.Level, filePath string) {
	lgr, err := logs.NewLogger(&logs.Config{
		Level:   level,
		LogFile: filePath,
		Caller:  false,
		AppName: "idp-core - CRON",
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