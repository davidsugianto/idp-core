package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	k8sPkg "github.com/davidsugianto/idp-core/internal/pkg/kubernetes"

	"github.com/davidsugianto/go-pkgs/db"
	"github.com/davidsugianto/go-pkgs/logs"
	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	argocdPkg "github.com/davidsugianto/idp-core/internal/pkg/argocd"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/opencost"
	"github.com/davidsugianto/idp-core/internal/pkg/prometheus"
	"github.com/davidsugianto/idp-core/internal/pkg/redislock"
	"github.com/davidsugianto/idp-core/internal/pkg/slack"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	buildApplicationRepository "github.com/davidsugianto/idp-core/internal/repository/build_application"
	budgetRepository "github.com/davidsugianto/idp-core/internal/repository/budget"
	costRepository "github.com/davidsugianto/idp-core/internal/repository/cost"
	deliveryTargetRepository "github.com/davidsugianto/idp-core/internal/repository/delivery_target"
	gitopsRepository "github.com/davidsugianto/idp-core/internal/repository/gitops"
	monitoringRepository "github.com/davidsugianto/idp-core/internal/repository/monitoring"
	notificationRepository "github.com/davidsugianto/idp-core/internal/repository/notification"
	provisionerRepository "github.com/davidsugianto/idp-core/internal/repository/provisioner"
	rightsizingRepository "github.com/davidsugianto/idp-core/internal/repository/rightsizing"
	buildApplicationUsecase "github.com/davidsugianto/idp-core/internal/usecase/build_application"
	budgetUsecase "github.com/davidsugianto/idp-core/internal/usecase/budget"
	costUsecase "github.com/davidsugianto/idp-core/internal/usecase/cost"
	notificationUsecase "github.com/davidsugianto/idp-core/internal/usecase/notification"
	rightsizingUsecase "github.com/davidsugianto/idp-core/internal/usecase/rightsizing"
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
	sentinelAddrs := strings.Split(cfg.Redis.Address, ",")
	for i := range sentinelAddrs {
		sentinelAddrs[i] = strings.TrimSpace(sentinelAddrs[i])
	}
	drd := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       cfg.Redis.MasterName,
		SentinelAddrs:    sentinelAddrs,
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

	k8sClient, err := k8sPkg.NewClient(cfg.Kubernetes.InCluster, cfg.Kubernetes.KubeconfigPath, cfg.Kubernetes.KubeconfigContext)
	if err != nil {
		logs.Fatal("cannot create k8s client")
	}

	// FinOps clients
	opencostClient := opencost.NewClient(opencost.Config{
		BaseURL: cfg.FinOps.OpenCost.BaseURL,
	})
	promClient := prometheus.NewClient(prometheus.Config{
		URL: cfg.FinOps.Prometheus.URL,
	})

	// Webhook validator
	webhookValidator := webhook.NewValidator()

	// Slack client
	slackClient := slack.NewClient(cfg.Slack.WebhookURL, cfg.Slack.Channel)

	defaultTargetControlPlane := (&deliveryTargetModel.TargetControlPlane{}).Resolve(cfg.DefaultTargetControlPlane())
	var argocdClient *argocdPkg.Client
	if defaultTargetControlPlane.ArgoCDServer != "" || defaultTargetControlPlane.ArgoCDNamespace != "" {
		argocdClient, err = argocdPkg.NewClient(defaultTargetControlPlane.InCluster, defaultTargetControlPlane.KubeconfigPath, defaultTargetControlPlane.KubeconfigContext)
		if err != nil {
			logs.Fatal("cannot create argocd client")
		}
	}

	// Repositories
	provisionerRepo := provisionerRepository.New(provisionerRepository.Dependencies{
		K8sClient: k8sClient,
	})
	costRepo := costRepository.New(costRepository.Dependencies{
		Database: dbClient,
	})
	budgetRepo := budgetRepository.New(budgetRepository.Dependencies{
		Database: dbClient,
	})
	rightsizingRepo := rightsizingRepository.New(rightsizingRepository.Dependencies{
		Database: dbClient,
	})
	monitoringRepo := monitoringRepository.New(monitoringRepository.Dependencies{
		PromClient: promClient,
	})
	notificationRepo := notificationRepository.New(notificationRepository.Dependencies{
		Database: dbClient,
	})
	buildAppRepo := buildApplicationRepository.New(buildApplicationRepository.Dependencies{
		Database: dbClient,
	})
	deliveryTargetRepo := deliveryTargetRepository.New(deliveryTargetRepository.Dependencies{
		Database: dbClient,
	})
	var gitopsRepo gitopsRepository.Repository
	if argocdClient != nil {
		gitopsRepo = gitopsRepository.New(gitopsRepository.Dependencies{
			ArgoCDClient:    argocdClient,
			ArgoCDNamespace: defaultTargetControlPlane.ArgoCDNamespace,
		})
	}
	gitopsProviderBuilder := gitopsRepository.NewProvider(gitopsRepository.ProviderDependencies{
		DefaultRepository: gitopsRepo,
		Defaults:          cfg.DefaultTargetControlPlane(),
		ClientFactory: func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (*argocdPkg.Client, error) {
			return argocdPkg.NewClient(target.InCluster, target.KubeconfigPath, target.KubeconfigContext)
		},
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
	rightsizingUC := rightsizingUsecase.New(rightsizingUsecase.Dependencies{
		RightsizingRepo: rightsizingRepo,
		ProvisionerRepo: provisionerRepo,
		MonitoringRepo:  monitoringRepo,
	})
	notificationUC := notificationUsecase.New(notificationUsecase.Dependencies{
		NotificationRepo: notificationRepo,
	})
	buildAppUC := buildApplicationUsecase.New(buildApplicationUsecase.Dependencies{
		BuildApplicationRepo: buildAppRepo,
		DeliveryTargetRepo:   deliveryTargetRepo,
		GitopsRepo:           gitopsRepo,
		GitopsProvider:       buildApplicationUsecase.GitopsProvider(gitopsProviderBuilder.ForTarget),
		NotificationUC:       notificationUC,
	})

	// Server
	server := New(Dependencies{
		Schedules:               cfg.Cron.Schedules,
		Port:                    cfg.Cron.Port,
		CostUseCase:             costUC,
		BudgetUseCase:           budgetUC,
		RightsizingUseCase:      rightsizingUC,
		BuildApplicationUseCase: buildAppUC,
		Config:                  cfg,
		Distlock:                distlock,
		WebhookValidator:        webhookValidator,
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
