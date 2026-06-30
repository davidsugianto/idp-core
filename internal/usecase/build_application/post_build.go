package build_application

import (
	"context"
	"errors"
	"fmt"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostBuildUpdate struct {
	ImageRepository       string
	ImageTag              string
	ImageDigest           string
	PublishedImageRef     string
	SBOMStatus            string
	SBOMReference         string
	ScanStatus            string
	ScanReference         string
	ScanSummary           string
	SigningStatus         string
	SignatureReference    string
	PolicyGateStatus      string
	PolicyGateReason      string
	DeploymentRequested   bool
	RequestedManifestPath string
	ResultingRevision     string
}

func (u *usecase) ProcessPostBuildOutcome(ctx context.Context, teamID, buildID string, update PostBuildUpdate) (*buildApplicationModel.BuildResponse, error) {
	build, err := u.assertBuildAccess(ctx, teamID, buildID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if build.Status != buildApplicationModel.BuildStatusRunning {
		build.Status = buildApplicationModel.BuildStatusRunning
		if build.StartedAt == nil {
			build.StartedAt = &now
		}
		build.UpdatedAt = now
		if err := u.repo.UpdateBuild(ctx, build); err != nil {
			return nil, err
		}
		_ = u.createLifecycleEvent(ctx, teamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeBuildRunning, "system", "build running")
	}

	artifact, err := u.upsertArtifact(ctx, build, update, now)
	if err != nil {
		return nil, err
	}

	verification, err := u.upsertSecurityVerification(ctx, build, artifact, update, now)
	if err != nil {
		return nil, err
	}

	if verification.PolicyGateStatus == "" {
		verification.PolicyGateStatus = verification.Status
	}

	if verification.PolicyGateStatus == buildApplicationModel.SecurityStatusFailed {
		build.Status = buildApplicationModel.BuildStatusBlocked
		build.FailureReason = verification.PolicyGateReason
		build.FinishedAt = &now
		build.UpdatedAt = now
		if err := u.repo.UpdateBuild(ctx, build); err != nil {
			return nil, err
		}
		_ = u.createLifecycleEvent(ctx, teamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeSecurityBlocked, "system", "security policy blocked deployment")
		return u.GetBuild(ctx, teamID, buildID)
	}

	if update.DeploymentRequested {
		if err := u.applyGitOpsDeploymentUpdate(ctx, teamID, build, update, now); err != nil {
			build.Status = buildApplicationModel.BuildStatusFailed
			build.FailureReason = sanitizeError(err).Error()
			build.FinishedAt = &now
			build.UpdatedAt = now
			_ = u.repo.UpdateBuild(ctx, build)
			_ = u.createLifecycleEvent(ctx, teamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeDeployFailed, "system", "deployment update failed")
			_ = u.createLifecycleEvent(ctx, teamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeBuildFailed, "system", "build failed")
			return nil, err
		}
	}

	build.Status = buildApplicationModel.BuildStatusDeploymentReady
	build.FinishedAt = &now
	build.UpdatedAt = now
	if err := u.repo.UpdateBuild(ctx, build); err != nil {
		return nil, err
	}
	_ = u.createLifecycleEvent(ctx, teamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeBuildSucceeded, "system", "build completed and deployment ready")

	return u.GetBuild(ctx, teamID, buildID)
}

func (u *usecase) upsertArtifact(ctx context.Context, build *buildApplicationModel.Build, update PostBuildUpdate, now time.Time) (*buildApplicationModel.BuildArtifact, error) {
	artifact, err := u.repo.GetBuildArtifactByBuildID(ctx, build.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	isCreate := artifact == nil || errors.Is(err, gorm.ErrRecordNotFound)
	if isCreate {
		artifact = &buildApplicationModel.BuildArtifact{
			ID:            uuid.New().String(),
			BuildID:       build.ID,
			ApplicationID: build.ApplicationID,
			CreatedAt:     now,
		}
	}

	artifact.ImageRepository = update.ImageRepository
	artifact.ImageTag = update.ImageTag
	artifact.ImageDigest = update.ImageDigest
	artifact.PublishedImageReference = update.PublishedImageRef
	artifact.PublishedAt = &now
	artifact.UpdatedAt = now

	if isCreate {
		if err := u.repo.CreateBuildArtifact(ctx, artifact); err != nil {
			return nil, err
		}
	} else {
		if err := u.repo.UpdateBuildArtifact(ctx, artifact); err != nil {
			return nil, err
		}
	}
	return artifact, nil
}

func (u *usecase) upsertSecurityVerification(ctx context.Context, build *buildApplicationModel.Build, artifact *buildApplicationModel.BuildArtifact, update PostBuildUpdate, now time.Time) (*buildApplicationModel.SecurityVerification, error) {
	if err := validateSecurityStatus(update.PolicyGateStatus); err != nil && update.PolicyGateStatus != "" {
		return nil, err
	}

	verification, err := u.repo.GetSecurityVerificationByBuildID(ctx, build.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	isCreate := verification == nil || errors.Is(err, gorm.ErrRecordNotFound)
	if isCreate {
		verification = &buildApplicationModel.SecurityVerification{
			ID:        uuid.New().String(),
			BuildID:   build.ID,
			CreatedAt: now,
		}
	}

	verification.ArtifactID = artifact.ID
	verification.Status = buildApplicationModel.SecurityStatusPassed
	verification.SBOMStatus = update.SBOMStatus
	verification.SBOMReference = update.SBOMReference
	verification.ScanStatus = update.ScanStatus
	verification.ScanReference = update.ScanReference
	verification.ScanSummary = update.ScanSummary
	verification.SigningStatus = update.SigningStatus
	verification.SignatureReference = update.SignatureReference
	verification.PolicyGateStatus = update.PolicyGateStatus
	verification.PolicyGateReason = update.PolicyGateReason
	verification.CompletedAt = &now
	verification.UpdatedAt = now

	if verification.PolicyGateStatus == buildApplicationModel.SecurityStatusFailed {
		verification.Status = buildApplicationModel.SecurityStatusFailed
	}

	if isCreate {
		if err := u.repo.CreateSecurityVerification(ctx, verification); err != nil {
			return nil, err
		}
	} else {
		if err := u.repo.UpdateSecurityVerification(ctx, verification); err != nil {
			return nil, err
		}
	}

	if verification.SBOMStatus != "" {
		_ = u.createLifecycleEvent(ctx, build.TeamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeSecuritySBOM, "system", "sbom generated")
	}
	if verification.ScanStatus != "" {
		_ = u.createLifecycleEvent(ctx, build.TeamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeSecurityScan, "system", "security scan completed")
	}
	if verification.SigningStatus != "" {
		_ = u.createLifecycleEvent(ctx, build.TeamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeSecuritySigning, "system", "image signing completed")
	}

	return verification, nil
}

func (u *usecase) applyGitOpsDeploymentUpdate(ctx context.Context, teamID string, build *buildApplicationModel.Build, update PostBuildUpdate, now time.Time) error {
	app, err := u.repo.GetApplicationByIDAndTeam(ctx, build.ApplicationID, teamID)
	if err != nil {
		return err
	}
	if app == nil {
		return ErrApplicationNotFound
	}
	if app.GitOpsTargetID == "" {
		return fmt.Errorf("gitops target is not configured")
	}

	deployment, err := u.repo.GetDeploymentUpdateByBuildID(ctx, build.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	isCreate := deployment == nil || errors.Is(err, gorm.ErrRecordNotFound)
	if isCreate {
		deployment = &buildApplicationModel.DeploymentUpdate{
			ID:            uuid.New().String(),
			BuildID:       build.ID,
			ApplicationID: build.ApplicationID,
			CreatedAt:     now,
		}
	}

	deployment.Status = buildApplicationModel.DeploymentUpdateStatusInProgress
	deployment.RequestedImageReference = update.PublishedImageRef
	deployment.RequestedManifestPath = update.RequestedManifestPath
	deployment.StartedAt = &now
	deployment.UpdatedAt = now
	if isCreate {
		if err := u.repo.CreateDeploymentUpdate(ctx, deployment); err != nil {
			return err
		}
	} else {
		if err := u.repo.UpdateDeploymentUpdate(ctx, deployment); err != nil {
			return err
		}
	}
	_ = u.createLifecycleEvent(ctx, teamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeDeployStarted, "system", "deployment update started")

	gitopsRepo := u.gitopsRepo
	if u.gitopsProvider != nil {
		if u.deliveryTargetRepo == nil {
			return fmt.Errorf("delivery target repository is not configured")
		}
		controlPlane, err := u.deliveryTargetRepo.GetControlPlaneByID(ctx, app.GitOpsTargetID)
		if err != nil {
			return err
		}
		gitopsRepo, err = u.gitopsProvider(ctx, controlPlane)
		if err != nil {
			return err
		}
	}
	if gitopsRepo == nil {
		return fmt.Errorf("gitops repository not configured")
	}
	if err := gitopsRepo.SyncApplication(ctx, app.GitOpsTargetID); err != nil {
		deployment.Status = buildApplicationModel.DeploymentUpdateStatusFailed
		deployment.FailureReason = sanitizeError(err).Error()
		deployment.FinishedAt = &now
		deployment.UpdatedAt = now
		_ = u.repo.UpdateDeploymentUpdate(ctx, deployment)
		return err
	}

	deployment.Status = buildApplicationModel.DeploymentUpdateStatusSucceeded
	deployment.ResultingRevision = update.ResultingRevision
	if deployment.ResultingRevision == "" {
		deployment.ResultingRevision = build.SourceRevisionResolved
	}
	deployment.FinishedAt = &now
	deployment.UpdatedAt = now
	if err := u.repo.UpdateDeploymentUpdate(ctx, deployment); err != nil {
		return err
	}
	_ = u.createLifecycleEvent(ctx, teamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeDeploySucceeded, "system", "deployment update succeeded")

	return nil
}
