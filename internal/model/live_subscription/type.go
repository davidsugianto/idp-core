package live_subscription

import "time"

const (
	ChannelStatus       = "status"
	ChannelProgress     = "progress"
	ChannelLog          = "log"
	ChannelNotification = "notification"
)

type LiveSubscription struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id,omitempty"`
	TeamID        string     `json:"team_id,omitempty"`
	EnvironmentID string     `json:"environment_id,omitempty"`
	Channel       string     `json:"channel"`
	WorkloadName  string     `json:"workload_name,omitempty"`
	ContainerName string     `json:"container_name,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	LastEventID   string     `json:"last_event_id,omitempty"`
}

func ValidChannel(value string) bool {
	switch value {
	case ChannelStatus, ChannelProgress, ChannelLog, ChannelNotification:
		return true
	default:
		return false
	}
}
