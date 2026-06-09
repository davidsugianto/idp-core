package live_subscription

import "time"

const (
	ChannelStatus       = "status"
	ChannelProgress     = "progress"
	ChannelLog          = "log"
	ChannelNotification = "notification"

	SessionStatusActive      = "active"
	SessionStatusExpired     = "expired"
	SessionStatusInvalidated = "invalidated"
	SessionStatusClosed      = "closed"
)

type LiveSubscription struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id,omitempty"`
	TeamID             string     `json:"team_id,omitempty"`
	EnvironmentID      string     `json:"environment_id,omitempty"`
	Channel            string     `json:"channel"`
	Channels           []string   `json:"channels,omitempty"`
	WorkloadName       string     `json:"workload_name,omitempty"`
	ContainerName      string     `json:"container_name,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	LastEventID        string     `json:"last_event_id,omitempty"`
	Status             string     `json:"status,omitempty"`
	SubscribedAt       *time.Time `json:"subscribed_at,omitempty"`
	LastHeartbeatAt    *time.Time `json:"last_heartbeat_at,omitempty"`
	InvalidatedAt      *time.Time `json:"invalidated_at,omitempty"`
	InvalidationReason string     `json:"invalidation_reason,omitempty"`
	ClosedAt           *time.Time `json:"closed_at,omitempty"`
}

type SubscribeEventsRequest struct {
	EnvironmentID string   `form:"environment_id" json:"environment_id"`
	Channels      []string `form:"channels" json:"channels"`
	LastEventID   string   `form:"last_event_id" json:"last_event_id,omitempty"`
}

type StreamLogsRequest struct {
	EnvironmentID string `form:"environment_id" json:"environment_id"`
	WorkloadName  string `form:"workload" json:"workload"`
	ContainerName string `form:"container" json:"container,omitempty"`
	TailLines     int    `form:"tail_lines" json:"tail_lines,omitempty"`
	LastEventID   string `form:"last_event_id" json:"last_event_id,omitempty"`
}

func ValidChannel(value string) bool {
	switch value {
	case ChannelStatus, ChannelProgress, ChannelLog, ChannelNotification:
		return true
	default:
		return false
	}
}

func ValidSessionStatus(value string) bool {
	switch value {
	case SessionStatusActive, SessionStatusExpired, SessionStatusInvalidated, SessionStatusClosed:
		return true
	default:
		return false
	}
}

func FilterValidChannels(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !ValidChannel(value) {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
