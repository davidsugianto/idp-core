package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/davidsugianto/go-pkgs/db"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/migration"
	"gorm.io/gorm"
)

var (
	direction  = flag.String("direction", "up", "Migration direction: up or down")
	configPath = flag.String("config", fmt.Sprintf("configs/config.%s.yaml", os.Getenv("APP_ENV")), "Path to config file")
)

func main() {
	flag.Parse()

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	dbConfig := db.NewConfig(db.Postgres, cfg.Database.Host, cfg.Database.Name).
		WithPort(cfg.Database.Port).
		WithCredentials(cfg.Database.User, cfg.Database.Password)

	dbClient, err := db.New(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	database := dbClient.DB

	switch *direction {
	case "up":
		if err := runMigrations(database); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✅ Migrations completed successfully")
	case "down":
		if err := rollbackMigration(database); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("✅ Rollback completed successfully")
	default:
		log.Fatalf("Unknown direction: %s. Use 'up' or 'down'", *direction)
	}
}

func runMigrations(db *gorm.DB) error {
	return migration.Migrate(db)
}

func rollbackMigration(db *gorm.DB) error {
	// For now, we only support AutoMigrate which doesn't support rollback
	// In production, use a proper migration tool like golang-migrate
	return fmt.Errorf("rollback not supported with AutoMigrate. Use SQL migrations for production")
}
