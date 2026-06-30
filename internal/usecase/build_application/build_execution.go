package build_application

import (
	"context"
	"errors"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (u *usecase) TriggerBuild(ctx context.Context, teamID, actorID, applicationID string, req *buildApplicationModel.TriggerBuildRequest) (*buildApplicationModel.BuildActionResponse, error) {
	app, err := u.assertApplicationAccess(ctx, teamID, applicationID)
	if err != nil {
		return nil, err
	}

	if req.IdempotencyKey != "" {
		existing, getErr := u.repo.GetBuildByApplicationAndIdempotencyKey(ctx, applicationID, req.IdempotencyKey)
		if getErr == nil && existing != nil {
			if req.SourceRevision != "" && existing.SourceRevisionRequested != req.SourceRevision {
				return nil, ErrBuildIdempotencyConflict
			}
			return toBuildActionResponse(existing), nil
		}
		if getErr != nil && !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return nil, getErr
		}
	}

	latestSeq, err := u.repo.GetLatestBuildSequenceByApplicationID(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	build := &buildApplicationModel.Build{
		ID:                      uuid.New().String(),
		ApplicationID:           applicationID,
		TeamID:                  teamID,
		SequenceNumber:          nextSequence(latestSeq),
		Status:                  buildApplicationModel.BuildStatusQueued,
		TriggerType:             buildApplicationModel.BuildTriggerTypeManual,
		TriggeredBy:             actorID,
		IdempotencyKey:          req.IdempotencyKey,
		SourceRevisionRequested: req.SourceRevision,
		QueuedAt:                &now,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if err := u.repo.CreateBuild(ctx, build); err != nil {
		return nil, err
	}

	if err := u.createLifecycleEvent(ctx, teamID, app.ID, build.ID, buildApplicationModel.EventTypeBuildQueued, "api", "build queued"); err != nil {
		return nil, err
	}

	return toBuildActionResponse(build), nil
}

func (u *usecase) GetBuild(ctx context.Context, teamID, buildID string) (*buildApplicationModel.BuildResponse, error) {
	build, err := u.assertBuildAccess(ctx, teamID, buildID)
	if err != nil {
		return nil, err
	}

	response := buildApplicationModel.ToBuildResponse(build)

	artifact, err := u.repo.GetBuildArtifactByBuildID(ctx, buildID)
	if err == nil && artifact != nil {
		response.Artifact = &buildApplicationModel.BuildArtifactResponse{
			PublishedImageReference: artifact.PublishedImageReference,
			ImageDigest:             artifact.ImageDigest,
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	verification, err := u.repo.GetSecurityVerificationByBuildID(ctx, buildID)
	if err == nil && verification != nil {
		response.SecurityVerification = &buildApplicationModel.SecurityVerificationResponse{
			Status:           verification.Status,
			SBOMStatus:       verification.SBOMStatus,
			ScanStatus:       verification.ScanStatus,
			SigningStatus:    verification.SigningStatus,
			PolicyGateStatus: verification.PolicyGateStatus,
			PolicyGateReason: verification.PolicyGateReason,
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	update, err := u.repo.GetDeploymentUpdateByBuildID(ctx, buildID)
	if err == nil && update != nil {
		response.DeploymentUpdate = &buildApplicationModel.DeploymentUpdateResponse{
			Status:            update.Status,
			ResultingRevision: update.ResultingRevision,
			FailureReason:     update.FailureReason,
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return response, nil
}

func (u *usecase) ListBuilds(ctx context.Context, teamID, applicationID string, limit, offset int) (*buildApplicationModel.BuildHistoryResponse, error) {
	if _, err := u.assertApplicationAccess(ctx, teamID, applicationID); err != nil {
		return nil, err
	}

	builds, total, err := u.repo.ListBuildsByApplication(ctx, applicationID, teamID, limit, offset)
	if err != nil {
		return nil, err
	}

	response := buildApplicationModel.ToBuildHistoryResponse(builds, total)
	for i := range response.Builds {
		buildID := response.Builds[i].ID

		artifact, artifactErr := u.repo.GetBuildArtifactByBuildID(ctx, buildID)
		if artifactErr == nil && artifact != nil {
			response.Builds[i].Artifact = &buildApplicationModel.BuildArtifactResponse{
				PublishedImageReference: artifact.PublishedImageReference,
				ImageDigest:             artifact.ImageDigest,
			}
		} else if artifactErr != nil && !errors.Is(artifactErr, gorm.ErrRecordNotFound) {
			return nil, artifactErr
		}

		verification, verificationErr := u.repo.GetSecurityVerificationByBuildID(ctx, buildID)
		if verificationErr == nil && verification != nil {
			response.Builds[i].SecurityVerification = &buildApplicationModel.SecurityVerificationResponse{
				Status:           verification.Status,
				SBOMStatus:       verification.SBOMStatus,
				ScanStatus:       verification.ScanStatus,
				SigningStatus:    verification.SigningStatus,
				PolicyGateStatus: verification.PolicyGateStatus,
				PolicyGateReason: verification.PolicyGateReason,
			}
		} else if verificationErr != nil && !errors.Is(verificationErr, gorm.ErrRecordNotFound) {
			return nil, verificationErr
		}

		update, updateErr := u.repo.GetDeploymentUpdateByBuildID(ctx, buildID)
		if updateErr == nil && update != nil {
			response.Builds[i].DeploymentUpdate = &buildApplicationModel.DeploymentUpdateResponse{
				Status:            update.Status,
				ResultingRevision: update.ResultingRevision,
				FailureReason:     update.FailureReason,
			}
		} else if updateErr != nil && !errors.Is(updateErr, gorm.ErrRecordNotFound) {
			return nil, updateErr
		}
	}

	return response, nil
}

func (u *usecase) RetryBuild(ctx context.Context, teamID, actorID, buildID string) (*buildApplicationModel.BuildActionResponse, error) {
	build, err := u.assertBuildAccess(ctx, teamID, buildID)
	if err != nil {
		return nil, err
	}
	if !isRetryAllowed(build.Status) {
		return nil, ErrBuildRetryNotAllowed
	}

	latestSeq, err := u.repo.GetLatestBuildSequenceByApplicationID(ctx, build.ApplicationID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	retryBuild := &buildApplicationModel.Build{
		ID:                      uuid.New().String(),
		ApplicationID:           build.ApplicationID,
		TeamID:                  build.TeamID,
		SequenceNumber:          nextSequence(latestSeq),
		Status:                  buildApplicationModel.BuildStatusQueued,
		TriggerType:             buildApplicationModel.BuildTriggerTypeRetry,
		TriggeredBy:             actorID,
		SourceRevisionRequested: build.SourceRevisionResolved,
		RetryOfBuildID:          build.ID,
		QueuedAt:                &now,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if retryBuild.SourceRevisionRequested == "" {
		retryBuild.SourceRevisionRequested = build.SourceRevisionRequested
	}
	if err := u.repo.CreateBuild(ctx, retryBuild); err != nil {
		return nil, err
	}
	if err := u.createLifecycleEvent(ctx, teamID, build.ApplicationID, retryBuild.ID, buildApplicationModel.EventTypeBuildQueued, "api", "retry build queued"); err != nil {
		return nil, err
	}

	return toBuildActionResponse(retryBuild), nil
}

func (u *usecase) CancelBuild(ctx context.Context, teamID, actorID, buildID string) (*buildApplicationModel.BuildActionResponse, error) {
	build, err := u.assertBuildAccess(ctx, teamID, buildID)
	if err != nil {
		return nil, err
	}
	if isBuildTerminal(build.Status) {
		return nil, ErrBuildNotCancelable
	}

	now := time.Now()
	build.CancelRequestedBy = actorID
	build.UpdatedAt = now

	if build.Status == buildApplicationModel.BuildStatusQueued {
		build.Status = buildApplicationModel.BuildStatusCanceled
		build.FinishedAt = &now
		if err := u.repo.UpdateBuild(ctx, build); err != nil {
			return nil, err
		}
		if err := u.createLifecycleEvent(ctx, teamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeBuildCanceled, "api", "queued build canceled"); err != nil {
			return nil, err
		}
		return toBuildActionResponse(build), nil
	}

	build.Status = buildApplicationModel.BuildStatusCanceling
	if err := u.repo.UpdateBuild(ctx, build); err != nil {
		return nil, err
	}
	if err := u.createLifecycleEvent(ctx, teamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeBuildCanceled, "api", "build cancellation requested"); err != nil {
		return nil, err
	}

	return toBuildActionResponse(build), nil
}
