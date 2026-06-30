package build_application

import (
	"context"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
)

func (r *repository) CreateApplication(ctx context.Context, app *buildApplicationModel.BuildApplication) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *repository) GetApplicationByID(ctx context.Context, id string) (*buildApplicationModel.BuildApplication, error) {
	var app buildApplicationModel.BuildApplication
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *repository) GetApplicationByIDAndTeam(ctx context.Context, id, teamID string) (*buildApplicationModel.BuildApplication, error) {
	var app buildApplicationModel.BuildApplication
	err := r.db.WithContext(ctx).
		Where("id = ? AND team_id = ?", id, teamID).
		First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *repository) GetApplicationByNameAndTeam(ctx context.Context, name, teamID string) (*buildApplicationModel.BuildApplication, error) {
	var app buildApplicationModel.BuildApplication
	err := r.db.WithContext(ctx).
		Where("name = ? AND team_id = ?", name, teamID).
		First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *repository) ListApplicationsByTeam(ctx context.Context, teamID string, req *buildApplicationModel.ListBuildApplicationsRequest) ([]buildApplicationModel.BuildApplication, int64, error) {
	query := r.db.WithContext(ctx).Model(&buildApplicationModel.BuildApplication{}).Where("team_id = ?", teamID)

	if req != nil {
		if req.Status != "" {
			query = query.Where("status = ?", req.Status)
		}
		if req.RuntimeFamily != "" {
			query = query.Where("runtime_family = ?", req.RuntimeFamily)
		}
		if req.RegistryTargetID != "" {
			query = query.Where("registry_target_id = ?", req.RegistryTargetID)
		}
		if req.DeploymentAutomationEnabled != nil {
			query = query.Where("deployment_automation_enabled = ?", *req.DeploymentAutomationEnabled)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("created_at DESC")
	if req != nil {
		if req.Limit > 0 {
			query = query.Limit(req.Limit)
		}
		if req.Offset > 0 {
			query = query.Offset(req.Offset)
		}
	}

	var apps []buildApplicationModel.BuildApplication
	if err := query.Find(&apps).Error; err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

func (r *repository) UpdateApplication(ctx context.Context, app *buildApplicationModel.BuildApplication) error {
	return r.db.WithContext(ctx).Save(app).Error
}

func (r *repository) SoftDeleteApplication(ctx context.Context, id, teamID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND team_id = ?", id, teamID).
		Delete(&buildApplicationModel.BuildApplication{}).Error
}
