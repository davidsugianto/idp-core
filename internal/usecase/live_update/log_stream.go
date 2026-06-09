package live_update

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"
	"time"

	liveSubscriptionModel "github.com/davidsugianto/idp-core/internal/model/live_subscription"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
)

func (u *usecase) StreamLogs(ctx context.Context, req *liveSubscriptionModel.StreamLogsRequest, subscription *liveSubscriptionModel.LiveSubscription) (<-chan notificationModel.StreamEvent, error) {
	if req == nil || req.WorkloadName == "" {
		return nil, ErrWorkloadRequired
	}
	if subscription == nil {
		return nil, ErrSubscriptionRequired
	}
	if u.environmentRepo == nil || u.provisionerRepo == nil {
		return nil, ErrLogStreamingDisabled
	}

	environmentID := subscription.EnvironmentID
	if environmentID == "" {
		environmentID = req.EnvironmentID
	}
	if environmentID == "" {
		return nil, ErrEnvironmentNotFound
	}

	env, err := u.environmentRepo.GetByID(ctx, environmentID)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}

	pod, err := u.provisionerRepo.ResolvePodForWorkload(env.Namespace, req.WorkloadName)
	if err != nil {
		return nil, err
	}
	if pod == nil {
		return nil, ErrWorkloadNotFound
	}

	if subscription.Channel == "" {
		subscription.Channel = liveSubscriptionModel.ChannelLog
	}
	subscription.EnvironmentID = env.ID
	subscription.WorkloadName = req.WorkloadName
	subscription.ContainerName = req.ContainerName
	if req.LastEventID != "" {
		subscription.LastEventID = req.LastEventID
	}

	events, err := u.StreamEvents(ctx, subscription)
	if err != nil {
		return nil, err
	}

	stream, err := u.provisionerRepo.StreamPodLogs(ctx, env.Namespace, pod.Name, req.ContainerName, int64(req.TailLines))
	if err != nil {
		_ = u.Unsubscribe(context.Background(), subscription.ID)
		return nil, err
	}

	go u.forwardLogStream(ctx, subscription.ID, env.ID, req.WorkloadName, req.ContainerName, stream)
	return events, nil
}

func (u *usecase) forwardLogStream(ctx context.Context, subscriptionID, environmentID, workloadName, containerName string, stream io.ReadCloser) {
	defer stream.Close()
	defer u.Unsubscribe(context.Background(), subscriptionID)

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		u.publishLogEvent(subscriptionID, &notificationModel.LogEventPayload{
			EnvironmentID: environmentID,
			Workload:      workloadName,
			Container:     containerName,
			Line:          line,
			Timestamp:     time.Now(),
		})

		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return
	}
}

func (u *usecase) publishLogEvent(subscriptionID string, payload *notificationModel.LogEventPayload) {
	if payload == nil {
		return
	}

	event := notificationModel.StreamEvent{
		Event:     notificationModel.EventLog,
		Data:      payload,
		Timestamp: time.Now(),
	}

	u.mu.RLock()
	state, exists := u.subscriptions[subscriptionID]
	u.mu.RUnlock()
	if !exists {
		return
	}

	state.events <- event
}
