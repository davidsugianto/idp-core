package build_application

import (
	"context"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
)

func (r *repository) CreateBuildArtifact(ctx context.Context, artifact *buildApplicationModel.BuildArtifact) error {
	return r.db.WithContext(ctx).Create(artifact).Error
}

func (r *repository) GetBuildArtifactByBuildID(ctx context.Context, buildID string) (*buildApplicationModel.BuildArtifact, error) {
	var artifact buildApplicationModel.BuildArtifact
	err := r.db.WithContext(ctx).Where("build_id = ?", buildID).First(&artifact).Error
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

func (r *repository) UpdateBuildArtifact(ctx context.Context, artifact *buildApplicationModel.BuildArtifact) error {
	return r.db.WithContext(ctx).Save(artifact).Error
}

func (r *repository) CreateSecurityVerification(ctx context.Context, verification *buildApplicationModel.SecurityVerification) error {
	return r.db.WithContext(ctx).Create(verification).Error
}

func (r *repository) GetSecurityVerificationByBuildID(ctx context.Context, buildID string) (*buildApplicationModel.SecurityVerification, error) {
	var verification buildApplicationModel.SecurityVerification
	err := r.db.WithContext(ctx).Where("build_id = ?", buildID).First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

func (r *repository) UpdateSecurityVerification(ctx context.Context, verification *buildApplicationModel.SecurityVerification) error {
	return r.db.WithContext(ctx).Save(verification).Error
}

func (r *repository) CreateDeploymentUpdate(ctx context.Context, update *buildApplicationModel.DeploymentUpdate) error {
	return r.db.WithContext(ctx).Create(update).Error
}

func (r *repository) GetDeploymentUpdateByBuildID(ctx context.Context, buildID string) (*buildApplicationModel.DeploymentUpdate, error) {
	var update buildApplicationModel.DeploymentUpdate
	err := r.db.WithContext(ctx).Where("build_id = ?", buildID).First(&update).Error
	if err != nil {
		return nil, err
	}
	return &update, nil
}

func (r *repository) UpdateDeploymentUpdate(ctx context.Context, update *buildApplicationModel.DeploymentUpdate) error {
	return r.db.WithContext(ctx).Save(update).Error
}

func (r *repository) CreateLifecycleEvent(ctx context.Context, event *buildApplicationModel.LifecycleEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *repository) ListLifecycleEventsByBuildID(ctx context.Context, buildID string) ([]buildApplicationModel.LifecycleEvent, error) {
	var events []buildApplicationModel.LifecycleEvent
	err := r.db.WithContext(ctx).
		Where("build_id = ?", buildID).
		Order("occurred_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *repository) ListLifecycleEventsByApplicationID(ctx context.Context, applicationID string) ([]buildApplicationModel.LifecycleEvent, error) {
	var events []buildApplicationModel.LifecycleEvent
	err := r.db.WithContext(ctx).
		Where("application_id = ?", applicationID).
		Order("occurred_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}
