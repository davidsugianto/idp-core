package main

import (
	"context"
	"fmt"

	"github.com/davidsugianto/idp-core/internal/pkg/config"

	"github.com/davidsugianto/go-pkgs/db"
	"github.com/davidsugianto/go-pkgs/logger"

	envRepository "github.com/davidsugianto/idp-core/internal/repository/environment"
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
)

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

	// Repositories
	envRepo := envRepository.New(envRepository.Dependencies{
		Database: dbClient,
	})

	// UseCases
	envUC := envUsecase.New(envUsecase.Dependencies{
		EnvironmentRepo: envRepo,
	})

	server := New(Dependencies{
		EnvironmentUseCase: envUC,
		Config:             cfg,
		Logger:             logs,
	})

	logs.Info().Int("port", cfg.Server.Port).Msg("listening on port")
	if err := server.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		logs.Fatal().Err(err).Msg("server failed to run")
	}
}
