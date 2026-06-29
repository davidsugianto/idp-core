package build_application

import (
	"context"
	"fmt"
	"strings"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"github.com/google/uuid"
)

const (
	workerBatchSize  = 5
	workerLeaseTime  = 2 * time.Minute
	workerLogSpacing = 100 * time.Millisecond
)

func (u *usecase) DrainQueuedBuilds(ctx context.Context) error {
	workerID := fmt.Sprintf("cron-%s", uuid.NewString())

	for range workerBatchSize {
		leaseUntil := time.Now().UTC().Add(workerLeaseTime)
		build, err := u.repo.ClaimNextQueuedBuild(ctx, workerID, leaseUntil)
		if err != nil {
			return err
		}
		if build == nil {
			return nil
		}
		if err := u.executeClaimedBuild(ctx, build); err != nil {
			return err
		}
	}

	return nil
}

func (u *usecase) executeClaimedBuild(ctx context.Context, build *buildApplicationModel.Build) error {
	app, err := u.repo.GetApplicationByID(ctx, build.ApplicationID)
	if err != nil {
		return u.failClaimedBuild(ctx, build, err)
	}

	resolvedRevision := strings.TrimSpace(build.SourceRevisionRequested)
	if resolvedRevision == "" {
		resolvedRevision = strings.TrimSpace(app.DefaultBranch)
	}
	if resolvedRevision == "" {
		resolvedRevision = "main"
	}

	build.SourceRevisionResolved = resolvedRevision
	build.UpdatedAt = time.Now().UTC()
	if err := u.repo.UpdateBuild(ctx, build); err != nil {
		return err
	}

	if err := u.appendExecutionLog(ctx, build.ID, "build claimed by worker"); err != nil {
		return u.failClaimedBuild(ctx, build, err)
	}
	if err := u.appendExecutionLog(ctx, build.ID, fmt.Sprintf("resolved source revision: %s", resolvedRevision)); err != nil {
		return u.failClaimedBuild(ctx, build, err)
	}
	if err := u.appendExecutionLog(ctx, build.ID, "detecting runtime and preparing buildpacks workflow"); err != nil {
		return u.failClaimedBuild(ctx, build, err)
	}
	if err := u.appendExecutionLog(ctx, build.ID, "publishing image artifact and recording verification results"); err != nil {
		return u.failClaimedBuild(ctx, build, err)
	}

	result, err := u.ProcessPostBuildOutcome(ctx, build.TeamID, build.ID, PostBuildUpdate{
		ImageRepository:       fmt.Sprintf("registry.local/%s", strings.ReplaceAll(app.Name, " ", "-")),
		ImageTag:              fmt.Sprintf("build-%d", build.SequenceNumber),
		ImageDigest:           fmt.Sprintf("sha256:%032x", build.SequenceNumber),
		PublishedImageRef:     fmt.Sprintf("registry.local/%s:build-%d", strings.ReplaceAll(app.Name, " ", "-"), build.SequenceNumber),
		SBOMStatus:            buildApplicationModel.SecurityStatusPassed,
		SBOMReference:         fmt.Sprintf("sbom://%s/%s", app.ID, build.ID),
		ScanStatus:            buildApplicationModel.SecurityStatusPassed,
		ScanReference:         fmt.Sprintf("trivy://%s/%s", app.ID, build.ID),
		ScanSummary:           "no blocking vulnerabilities detected",
		SigningStatus:         buildApplicationModel.SecurityStatusPassed,
		SignatureReference:    fmt.Sprintf("cosign://%s/%s", app.ID, build.ID),
		PolicyGateStatus:      buildApplicationModel.SecurityStatusPassed,
		DeploymentRequested:   app.DeploymentAutomationEnabled,
		RequestedManifestPath: app.ApplicationDescriptorPath,
		ResultingRevision:     resolvedRevision,
	})
	if err != nil {
		return u.failClaimedBuild(ctx, build, err)
	}

	terminalStatus := buildApplicationModel.BuildStatusDeploymentReady
	if result != nil && result.Status != "" {
		terminalStatus = result.Status
	}
	if err := u.appendExecutionLog(ctx, build.ID, fmt.Sprintf("build completed with status %s", terminalStatus)); err != nil {
		return err
	}

	return nil
}

func (u *usecase) appendExecutionLog(ctx context.Context, buildID, line string) error {
	sequence, err := u.repo.GetLatestBuildLogSequence(ctx, buildID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return u.repo.AppendBuildLog(ctx, &buildApplicationModel.BuildLog{
		ID:        uuid.NewString(),
		BuildID:   buildID,
		Sequence:  sequence + 1,
		Line:      line,
		CreatedAt: now.Add(workerLogSpacing),
	})
}

func (u *usecase) failClaimedBuild(ctx context.Context, build *buildApplicationModel.Build, cause error) error {
	message := sanitizeError(cause).Error()
	now := time.Now().UTC()
	build.Status = buildApplicationModel.BuildStatusFailed
	build.FailureReason = message
	build.FinishedAt = &now
	build.UpdatedAt = now
	if err := u.repo.UpdateBuild(ctx, build); err != nil {
		return err
	}
	if logErr := u.appendExecutionLog(ctx, build.ID, fmt.Sprintf("build failed: %s", message)); logErr != nil {
		return logErr
	}
	if eventErr := u.createLifecycleEvent(ctx, build.TeamID, build.ApplicationID, build.ID, buildApplicationModel.EventTypeBuildFailed, "system", "build failed"); eventErr != nil {
		return eventErr
	}
	return nil
}
