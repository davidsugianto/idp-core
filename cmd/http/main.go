package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/davidsugianto/idp-core/internal/model/apikey"
	"github.com/davidsugianto/idp-core/internal/model/auditlog"
	"github.com/davidsugianto/idp-core/internal/model/budget"
	"github.com/davidsugianto/idp-core/internal/model/cost"
	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	environmentMovementModel "github.com/davidsugianto/idp-core/internal/model/environment_movement"
	"github.com/davidsugianto/idp-core/internal/model/permission"
	"github.com/davidsugianto/idp-core/internal/model/resourcequota"
	"github.com/davidsugianto/idp-core/internal/model/rightsizing"
	"github.com/davidsugianto/idp-core/internal/model/role"
	"github.com/davidsugianto/idp-core/internal/model/service"
	"github.com/davidsugianto/idp-core/internal/model/service_dependency"
	"github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	"github.com/davidsugianto/idp-core/internal/model/service_environment"
	"github.com/davidsugianto/idp-core/internal/model/service_version"
	"github.com/davidsugianto/idp-core/internal/model/team"
	"github.com/davidsugianto/idp-core/internal/model/user"
	argocdPkg "github.com/davidsugianto/idp-core/internal/pkg/argocd"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	oidcPkg "github.com/davidsugianto/idp-core/internal/pkg/oidc"
	"github.com/davidsugianto/idp-core/internal/pkg/opencost"
	"github.com/davidsugianto/idp-core/internal/pkg/prometheus"
	"github.com/davidsugianto/idp-core/internal/pkg/slack"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"

	"github.com/davidsugianto/go-pkgs/db"
	k8sPkg "github.com/davidsugianto/idp-core/internal/pkg/kubernetes"

	apikeyRepository "github.com/davidsugianto/idp-core/internal/repository/apikey"
	auditlogRepository "github.com/davidsugianto/idp-core/internal/repository/auditlog"
	budgetRepository "github.com/davidsugianto/idp-core/internal/repository/budget"
	costRepository "github.com/davidsugianto/idp-core/internal/repository/cost"
	deliveryTargetRepository "github.com/davidsugianto/idp-core/internal/repository/delivery_target"
	envRepository "github.com/davidsugianto/idp-core/internal/repository/environment"
	environmentMovementRepository "github.com/davidsugianto/idp-core/internal/repository/environment_movement"
	gitopsRepository "github.com/davidsugianto/idp-core/internal/repository/gitops"
	monitoringRepository "github.com/davidsugianto/idp-core/internal/repository/monitoring"
	notificationRepository "github.com/davidsugianto/idp-core/internal/repository/notification"
	permissionRepository "github.com/davidsugianto/idp-core/internal/repository/permission"
	provisionerRepository "github.com/davidsugianto/idp-core/internal/repository/provisioner"
	quotaRepository "github.com/davidsugianto/idp-core/internal/repository/quota"
	rightsizingRepository "github.com/davidsugianto/idp-core/internal/repository/rightsizing"
	roleRepository "github.com/davidsugianto/idp-core/internal/repository/role"
	serviceRepository "github.com/davidsugianto/idp-core/internal/repository/service"
	teamRepository "github.com/davidsugianto/idp-core/internal/repository/team"
	templateRepository "github.com/davidsugianto/idp-core/internal/repository/template"
	userRepository "github.com/davidsugianto/idp-core/internal/repository/user"
	apikeyUsecase "github.com/davidsugianto/idp-core/internal/usecase/apikey"
	auditlogUsecase "github.com/davidsugianto/idp-core/internal/usecase/auditlog"
	budgetUsecase "github.com/davidsugianto/idp-core/internal/usecase/budget"
	costUsecase "github.com/davidsugianto/idp-core/internal/usecase/cost"
	deliveryTargetUsecase "github.com/davidsugianto/idp-core/internal/usecase/delivery_target"
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
	environmentMovementUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment_movement"
	liveUpdateUsecase "github.com/davidsugianto/idp-core/internal/usecase/live_update"
	notificationUsecase "github.com/davidsugianto/idp-core/internal/usecase/notification"
	quotaUsecase "github.com/davidsugianto/idp-core/internal/usecase/quota"
	rightsizingUsecase "github.com/davidsugianto/idp-core/internal/usecase/rightsizing"
	roleUsecase "github.com/davidsugianto/idp-core/internal/usecase/role"
	serviceUsecase "github.com/davidsugianto/idp-core/internal/usecase/service"
	teamUsecase "github.com/davidsugianto/idp-core/internal/usecase/team"
	templateUsecase "github.com/davidsugianto/idp-core/internal/usecase/template"
	userUsecase "github.com/davidsugianto/idp-core/internal/usecase/user"

	"github.com/davidsugianto/go-pkgs/logs"
	"github.com/davidsugianto/idp-core/internal/seed"
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
		logs.Fatalf("cannot connect to database: %v", err)
	}
	dbClient := dbClientWrapper.DB

	// Auto-migrate shared baseline tables.
	if err := dbClient.AutoMigrate(&user.User{}, &team.Team{}, &team.TeamMember{}, &role.Role{}, &permission.Permission{}, &role.UserRole{}, &apikey.APIKey{}, &auditlog.AuditLog{}, &cost.CostRecord{}, &budget.Budget{}, &budget.BudgetAlert{}, &rightsizing.RightsizingRecommendation{}, &resourcequota.ResourceQuota{}, &service.Service{}, &service_version.ServiceVersion{}, &service_endpoint.ServiceEndpoint{}, &service_dependency.ServiceDependency{}, &service_environment.ServiceEnvironment{}, &deliveryTargetModel.DeliveryTarget{}, &environmentMovementModel.EnvironmentMovement{}); err != nil {
		logs.Fatalf("cannot migrate database: %v", err)
	}

	// K8s client
	k8sClient, err := k8sPkg.NewClient(cfg.Kubernetes.InCluster, cfg.Kubernetes.KubeconfigPath)
	if err != nil {
		logs.Fatalf("cannot create k8s client: %v", err)
	}

	// FinOps clients
	opencostClient := opencost.NewClient(opencost.Config{
		BaseURL: cfg.FinOps.OpenCost.BaseURL,
	})
	promClient := prometheus.NewClient(prometheus.Config{
		URL: cfg.FinOps.Prometheus.URL,
	})

	var argocdClient *argocdPkg.Client
	if cfg.ArgoCD.BaseURL != "" || cfg.ArgoCD.Namespace != "" {
		argocdClient, err = argocdPkg.NewClient(cfg.Kubernetes.InCluster, cfg.Kubernetes.KubeconfigPath)
		if err != nil {
			logs.Fatalf("cannot create argocd client: %v", err)
		}
	}

	// Slack client
	slackClient := slack.NewClient(cfg.Slack.WebhookURL, cfg.Slack.Channel)

	// Repositories
	provisionerRepo := provisionerRepository.New(provisionerRepository.Dependencies{
		K8sClient: k8sClient,
	})
	envRepo := envRepository.New(envRepository.Dependencies{
		Database: dbClient,
	})
	var gitopsRepo gitopsRepository.Repository
	if argocdClient != nil {
		gitopsRepo = gitopsRepository.New(gitopsRepository.Dependencies{
			ArgoCDClient:    argocdClient,
			ArgoCDNamespace: cfg.ArgoCD.Namespace,
		})
	}
	notificationRepo := notificationRepository.New(notificationRepository.Dependencies{
		Database: dbClient,
	})
	userRepo := userRepository.New(userRepository.Dependencies{
		Database: dbClient,
	})
	teamRepo := teamRepository.New(teamRepository.Dependencies{
		Database: dbClient,
	})
	roleRepo := roleRepository.New(roleRepository.Dependencies{
		Database: dbClient,
	})
	permRepo := permissionRepository.New(permissionRepository.Dependencies{
		Database: dbClient,
	})
	apiKeyRepo := apikeyRepository.New(apikeyRepository.Dependencies{
		Database: dbClient,
	})
	auditLogRepo := auditlogRepository.New(auditlogRepository.Dependencies{
		Database: dbClient,
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
	quotaRepo := quotaRepository.New(quotaRepository.Dependencies{
		Database: dbClient,
	})
	serviceRepo := serviceRepository.New(serviceRepository.Dependencies{
		Database: dbClient,
	})
	templateRepo := templateRepository.New(templateRepository.Dependencies{
		Database: dbClient,
	})
	deliveryTargetRepo := deliveryTargetRepository.New(deliveryTargetRepository.Dependencies{
		Database: dbClient,
	})
	environmentMovementRepo := environmentMovementRepository.New(environmentMovementRepository.Dependencies{
		Database: dbClient,
	})

	// UseCases
	notificationUC := notificationUsecase.New(notificationUsecase.Dependencies{
		NotificationRepo: notificationRepo,
	})
	liveUpdateUC := liveUpdateUsecase.New(liveUpdateUsecase.Dependencies{})
	deliveryTargetUC := deliveryTargetUsecase.New(deliveryTargetUsecase.Dependencies{
		DeliveryTargetRepo:      deliveryTargetRepo,
		EnvironmentRepo:         envRepo,
		EnvironmentMovementRepo: environmentMovementRepo,
	})
	environmentMovementUC := environmentMovementUsecase.New(environmentMovementUsecase.Dependencies{
		EnvironmentRepo:         envRepo,
		DeliveryTargetRepo:      deliveryTargetRepo,
		EnvironmentMovementRepo: environmentMovementRepo,
	})
	envUC := envUsecase.New(envUsecase.Dependencies{
		EnvironmentRepo:    envRepo,
		DeliveryTargetRepo: deliveryTargetRepo,
		ProvisionerRepo:    provisionerRepo,
		GitopsRepo:         gitopsRepo,
		TemplateRepo:       templateRepo,
		NotificationUC:     notificationUC,
		LiveUpdateUC:       liveUpdateUC,
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
	quotaUC := quotaUsecase.New(quotaUsecase.Dependencies{
		QuotaRepo:       quotaRepo,
		ProvisionerRepo: provisionerRepo,
	})
	serviceUC := serviceUsecase.New(serviceUsecase.Dependencies{
		ServiceRepo:     serviceRepo,
		EnvironmentRepo: envRepo,
	})
	templateUC := templateUsecase.New(templateUsecase.Dependencies{
		TemplateRepo: templateRepo,
	})

	// Seed default data (roles, permissions, users, default team)
	seeder := seed.NewSeeder(roleRepo, permRepo, userRepo, teamRepo)
	if err := seeder.SeedAll(ctx); err != nil {
		logs.Fatalf("cannot seed database: %v", err)
	}

	// Webhook validator
	webhookValidator := webhook.NewValidatorWithQuota(quotaUC)

	// OIDC client initialization
	var oidcClient *oidcPkg.Client
	var oidcVerifier *oidcPkg.Verifier
	if cfg.OIDC.Enabled {
		oidcClient, err = oidcPkg.NewClient(ctx, &oidcPkg.Config{
			IssuerURL:          cfg.OIDC.IssuerURL,
			DiscoveryURL:       cfg.OIDC.DiscoveryURL,
			ClientID:           cfg.OIDC.ClientID,
			ClientSecret:       cfg.OIDC.ClientSecret,
			RedirectURL:        cfg.OIDC.RedirectURL,
			Scopes:             cfg.OIDC.Scopes,
			InsecureIssuerURLs: cfg.OIDC.InsecureIssuerURLs,
		})
		if err != nil {
			logs.Fatalf("cannot initialize OIDC client: %v", err)
		}

		oidcVerifier = oidcPkg.NewVerifier(oidcClient, &oidcPkg.VerifierConfig{
			GroupsClaim: cfg.OIDC.GroupsClaim,
			AdminGroup:  cfg.OIDC.AdminGroup,
		})

		logs.Info("OIDC authentication enabled")
	}

	server := New(Dependencies{
		EnvironmentUseCase:         envUC,
		UserUseCase:                userUC,
		TeamUseCase:                teamUC,
		RoleUseCase:                roleUC,
		ApiKeyUseCase:              apiKeyUC,
		AuditLogUseCase:            auditLogUC,
		BudgetUseCase:              budgetUC,
		CostUseCase:                costUC,
		RightsizingUseCase:         rightsizingUC,
		QuotaUseCase:               quotaUC,
		ServiceUseCase:             serviceUC,
		TemplateUseCase:            templateUC,
		DeliveryTargetUseCase:      deliveryTargetUC,
		EnvironmentMovementUseCase: environmentMovementUC,
		NotificationUseCase:        notificationUC,
		LiveUpdateUseCase:          liveUpdateUC,
		Config:                     cfg,
		WebhookValidator:           webhookValidator,
		OIDCClient:                 oidcClient,
		OIDCVerifier:               oidcVerifier,
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
