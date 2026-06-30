package build_application

import (
	"context"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
)

func (u *usecase) StreamBuildLogs(ctx context.Context, teamID, buildID string, afterSequence int64, limit int) (*buildApplicationModel.BuildLogStreamResponse, error) {
	build, err := u.assertBuildAccess(ctx, teamID, buildID)
	if err != nil {
		return nil, err
	}

	logs, err := u.repo.ListBuildLogs(ctx, buildID, afterSequence, limit)
	if err != nil {
		return nil, err
	}

	lines := make([]string, len(logs))
	lastSequence := afterSequence
	for i := range logs {
		lines[i] = logs[i].Line
		if logs[i].Sequence > lastSequence {
			lastSequence = logs[i].Sequence
		}
	}

	streamState := "live"
	terminalSummary := ""
	if isBuildTerminal(build.Status) {
		streamState = "completed"
		summary, err := u.repo.GetBuildLogSummary(ctx, buildID)
		if err == nil {
			terminalSummary = summary
		}
	}

	return &buildApplicationModel.BuildLogStreamResponse{
		BuildID:         buildID,
		StreamState:     streamState,
		LastSequence:    lastSequence,
		TerminalSummary: terminalSummary,
		Lines:           lines,
	}, nil
}
