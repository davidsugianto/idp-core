package migration

import (
	"github.com/davidsugianto/idp-core/internal/model/apikey"
	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/model/workload"
	"gorm.io/gorm"
)

// Migrate runs all auto-migrations for Phase 1 models
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&environment.Environment{},
		&workload.WorkloadStatus{},
		&workload.PodStatus{},
		&apikey.APIKey{},
		&buildApplicationModel.BuildApplication{},
		&buildApplicationModel.Build{},
		&buildApplicationModel.BuildArtifact{},
		&buildApplicationModel.SecurityVerification{},
		&buildApplicationModel.DeploymentUpdate{},
		&buildApplicationModel.LifecycleEvent{},
		&buildApplicationModel.BuildLog{},
	)
}
