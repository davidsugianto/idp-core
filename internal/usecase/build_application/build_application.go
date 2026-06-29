package build_application

import (
	"context"
	"errors"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (u *usecase) CreateApplication(ctx context.Context, teamID, actorID string, req *buildApplicationModel.CreateBuildApplicationRequest) (*buildApplicationModel.BuildApplicationResponse, error) {
	if err := validateRuntimeFamily(req.RuntimeFamily); err != nil {
		return nil, err
	}

	if err := u.validateRegistryTargetForApplication(ctx, req.RegistryTargetID); err != nil {
		return nil, err
	}

	_, err := u.repo.GetApplicationByNameAndTeam(ctx, req.Name, teamID)
	if err == nil {
		return nil, ErrApplicationAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now()
	app := &buildApplicationModel.BuildApplication{
		ID:                        uuid.New().String(),
		TeamID:                    teamID,
		Name:                      req.Name,
		Description:               req.Description,
		Status:                    buildApplicationModel.ApplicationStatusActive,
		RepositoryURL:             req.RepositoryURL,
		ApplicationDescriptorPath: req.ApplicationDescriptorPath,
		RuntimeFamily:             normalizeRuntimeFamily(req.RuntimeFamily),
		RuntimeDetectionMode:      req.RuntimeDetectionMode,
		BuilderProfileID:          req.BuilderProfileID,
		RegistryTargetID:          req.RegistryTargetID,
		GitOpsTargetID:            req.GitOpsTargetID,
		DeploymentAutomationEnabled: req.DeploymentAutomationEnabled,
		CreatedBy:                 actorID,
		UpdatedBy:                 actorID,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
	if app.RuntimeDetectionMode == "" {
		app.RuntimeDetectionMode = buildApplicationModel.RuntimeDetectionModeAuto
	}

	if err := u.repo.CreateApplication(ctx, app); err != nil {
		return nil, err
	}
	if err := u.createLifecycleEvent(ctx, teamID, app.ID, "", buildApplicationModel.EventTypeApplicationCreated, "api", "application created"); err != nil {
		return nil, err
	}

	return buildApplicationModel.ToBuildApplicationResponse(app), nil
}

func (u *usecase) ListApplications(ctx context.Context, teamID string, req *buildApplicationModel.ListBuildApplicationsRequest) (*buildApplicationModel.BuildApplicationListResponse, error) {
	if err := validateApplicationStatus(req.Status); err != nil {
		return nil, err
	}
	if err := validateRuntimeFamily(req.RuntimeFamily); err != nil {
		return nil, err
	}

	apps, total, err := u.repo.ListApplicationsByTeam(ctx, teamID, req)
	if err != nil {
		return nil, err
	}
	return buildApplicationModel.ToBuildApplicationListResponse(apps, total), nil
}

func (u *usecase) GetApplication(ctx context.Context, teamID, applicationID string) (*buildApplicationModel.BuildApplicationResponse, error) {
	app, err := u.assertApplicationAccess(ctx, teamID, applicationID)
	if err != nil {
		return nil, err
	}
	return buildApplicationModel.ToBuildApplicationResponse(app), nil
}

func (u *usecase) UpdateApplication(ctx context.Context, teamID, actorID, applicationID string, req *buildApplicationModel.UpdateBuildApplicationRequest) (*buildApplicationModel.BuildApplicationResponse, error) {
	app, err := u.assertApplicationAccess(ctx, teamID, applicationID)
	if err != nil {
		return nil, err
	}

	if req.RuntimeFamily != nil {
		if err := validateRuntimeFamily(*req.RuntimeFamily); err != nil {
			return nil, err
		}
		app.RuntimeFamily = normalizeRuntimeFamily(*req.RuntimeFamily)
	}
	if req.RegistryTargetID != nil {
		if err := u.validateRegistryTargetForApplication(ctx, *req.RegistryTargetID); err != nil {
			return nil, err
		}
		app.RegistryTargetID = *req.RegistryTargetID
	}
	if req.Status != nil {
		if err := validateApplicationStatus(*req.Status); err != nil {
			return nil, err
		}
		app.Status = *req.Status
	}
	if req.Description != nil {
		app.Description = *req.Description
	}
	if req.RuntimeDetectionMode != nil {
		app.RuntimeDetectionMode = *req.RuntimeDetectionMode
	}
	if req.BuilderProfileID != nil {
		app.BuilderProfileID = *req.BuilderProfileID
	}
	if req.DeploymentAutomationEnabled != nil {
		app.DeploymentAutomationEnabled = *req.DeploymentAutomationEnabled
	}
	if req.GitOpsTargetID != nil {
		app.GitOpsTargetID = *req.GitOpsTargetID
	}

	app.UpdatedBy = actorID
	app.UpdatedAt = time.Now()

	if err := u.repo.UpdateApplication(ctx, app); err != nil {
		return nil, err
	}
	if err := u.createLifecycleEvent(ctx, teamID, app.ID, "", buildApplicationModel.EventTypeApplicationUpdated, "api", "application updated"); err != nil {
		return nil, err
	}
	return buildApplicationModel.ToBuildApplicationResponse(app), nil
}

func (u *usecase) validateRegistryTargetForApplication(ctx context.Context, registryTargetID string) error {
	if registryTargetID == "" {
		return nil
	}
	registryType, err := u.resolveRegistryType(ctx, registryTargetID)
	if err != nil {
		return err
	}
	if registryType == "" {
		return ErrInvalidRegistryType
	}
	return validateRegistryType(registryType)
}

func (u *usecase) DeleteApplication(ctx context.Context, teamID, actorID, applicationID string) error {
	app, err := u.assertApplicationAccess(ctx, teamID, applicationID)
	if err != nil {
		return err
	}

	app.Status = buildApplicationModel.ApplicationStatusDeleted
	app.UpdatedBy = actorID
	app.UpdatedAt = time.Now()
	if err := u.repo.UpdateApplication(ctx, app); err != nil {
		return err
	}
	if err := u.repo.SoftDeleteApplication(ctx, applicationID, teamID); err != nil {
		return err
	}
	if err := u.createLifecycleEvent(ctx, teamID, applicationID, "", buildApplicationModel.EventTypeApplicationDeleted, "api", "application deleted"); err != nil {
		return err
	}
	return nil
}
