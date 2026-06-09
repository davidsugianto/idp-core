package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/davidsugianto/go-pkgs/response"
	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	liveSubscriptionModel "github.com/davidsugianto/idp-core/internal/model/live_subscription"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
	liveUpdateUsecase "github.com/davidsugianto/idp-core/internal/usecase/live_update"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const liveUpdateStreamTTL = 30 * time.Minute

// ListNotifications godoc
// @Summary List recent notifications
// @Description List recent notifications visible to the authenticated caller with optional environment and kind filters
// @Tags live-updates
// @Produce json
// @Param environment_id query string false "Filter by environment ID"
// @Param kind query string false "Filter by notification kind"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} notificationModel.NotificationListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/notifications [get]
// @Security ApiKeyAuth
func (h *Handler) ListNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	var req notificationModel.ListNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	if !middleware.IsAdmin(c) {
		teamID := middleware.GetTeamID(c)
		if teamID == "" {
			response.GinUnauthorized(c, errors.New("unauthorized"))
			return
		}
		req.TeamID = teamID
	}
	req.UserID = userID

	result, err := h.liveUpdateUseCase.ListNotifications(c.Request.Context(), &req)
	if err != nil {
		response.GinInternalServerError(c, err)
		return
	}

	response.GinSuccess(c, result)
}

// StreamEnvironmentEvents godoc
// @Summary Stream environment events
// @Description Open an authenticated SSE stream for environment status, progress, and notification events
// @Tags live-updates
// @Produce text/event-stream
// @Param id path string true "Environment ID"
// @Param channels query string false "Comma-separated channels: status,progress,notification"
// @Param last_event_id query string false "Last delivered SSE event ID"
// @Success 200 {string} string "SSE stream"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 410 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/environments/{id}/events/stream [get]
// @Security ApiKeyAuth
func (h *Handler) StreamEnvironmentEvents(c *gin.Context) {
	environmentID := c.Param("id")
	if environmentID == "" {
		response.GinBadRequest(c, errors.New("missing environment id"))
		return
	}

	userID := middleware.GetUserID(c)
	teamID := middleware.GetTeamID(c)
	if userID == "" || teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	var req liveSubscriptionModel.SubscribeEventsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}

	channels := parseChannels(c.Query("channels"))
	if len(channels) == 0 {
		channels = []string{liveSubscriptionModel.ChannelStatus, liveSubscriptionModel.ChannelProgress, liveSubscriptionModel.ChannelNotification}
	}

	subscription, err := h.authorizeEnvironmentStream(c, environmentID, userID, teamID)
	if err != nil {
		h.writeStreamError(c, err)
		return
	}
	subscription.Channels = channels
	subscription.LastEventID = req.LastEventID

	stream, err := h.liveUpdateUseCase.StreamEvents(c.Request.Context(), subscription)
	if err != nil {
		h.writeStreamError(c, err)
		return
	}

	h.streamSSE(c, environmentID, teamID, subscription.ID, subscription.ExpiresAt, stream)
}

// StreamEnvironmentLogs godoc
// @Summary Stream environment logs
// @Description Open an authenticated SSE stream for workload log lines within an environment
// @Tags live-updates
// @Produce text/event-stream
// @Param id path string true "Environment ID"
// @Param workload query string true "Workload name"
// @Param container query string false "Container name"
// @Param tail_lines query int false "Number of historical log lines to include"
// @Param last_event_id query string false "Last delivered SSE event ID"
// @Success 200 {string} string "SSE stream"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 410 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/environments/{id}/logs/stream [get]
// @Security ApiKeyAuth
func (h *Handler) StreamEnvironmentLogs(c *gin.Context) {
	environmentID := c.Param("id")
	if environmentID == "" {
		response.GinBadRequest(c, errors.New("missing environment id"))
		return
	}

	userID := middleware.GetUserID(c)
	teamID := middleware.GetTeamID(c)
	if userID == "" || teamID == "" {
		response.GinUnauthorized(c, errors.New("unauthorized"))
		return
	}

	var req liveSubscriptionModel.StreamLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.GinBadRequest(c, err)
		return
	}
	if req.EnvironmentID == "" {
		req.EnvironmentID = environmentID
	}

	subscription, err := h.authorizeEnvironmentStream(c, environmentID, userID, teamID)
	if err != nil {
		h.writeStreamError(c, err)
		return
	}
	subscription.Channel = liveSubscriptionModel.ChannelLog

	stream, err := h.liveUpdateUseCase.StreamLogs(c.Request.Context(), &req, subscription)
	if err != nil {
		h.writeStreamError(c, err)
		return
	}

	h.streamSSE(c, environmentID, teamID, subscription.ID, subscription.ExpiresAt, stream)
}

func (h *Handler) authorizeEnvironmentStream(c *gin.Context, environmentID, userID, teamID string) (*liveSubscriptionModel.LiveSubscription, error) {
	if _, err := h.environmentUseCase.Get(c.Request.Context(), teamID, environmentID); err != nil {
		if errors.Is(err, envUsecase.ErrEnvironmentNotFound) {
			return nil, errStreamForbidden
		}
		return nil, err
	}

	expiresAt := time.Now().Add(liveUpdateStreamTTL)
	return &liveSubscriptionModel.LiveSubscription{
		ID:            uuid.NewString(),
		UserID:        userID,
		TeamID:        teamID,
		EnvironmentID: environmentID,
		ExpiresAt:     &expiresAt,
		LastHeartbeatAt: func() *time.Time {
			now := time.Now()
			return &now
		}(),
	}, nil
}

var errStreamForbidden = errors.New("stream access forbidden")

func (h *Handler) streamSSE(c *gin.Context, environmentID, teamID, subscriptionID string, expiresAt *time.Time, stream <-chan notificationModel.StreamEvent) {
	defer func() { _ = h.liveUpdateUseCase.Unsubscribe(c.Request.Context(), subscriptionID) }()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.GinInternalServerError(c, errors.New("streaming unsupported"))
		return
	}

	started := false
	keepAliveTicker := time.NewTicker(15 * time.Second)
	accessTicker := time.NewTicker(5 * time.Second)
	defer keepAliveTicker.Stop()
	defer accessTicker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-accessTicker.C:
			if expiresAt != nil && time.Now().After(*expiresAt) {
				if !started {
					response.GinError(c, http.StatusGone, liveUpdateUsecase.ErrStreamExpired)
				}
				return
			}
			if _, err := h.environmentUseCase.Get(c.Request.Context(), teamID, environmentID); err != nil {
				if !started && errors.Is(err, envUsecase.ErrEnvironmentNotFound) {
					response.GinError(c, http.StatusGone, errStreamForbidden)
				}
				return
			}
		case <-keepAliveTicker.C:
			if !started {
				c.Status(http.StatusOK)
				started = true
			}
			_, _ = c.Writer.Write([]byte(": keep-alive\n\n"))
			flusher.Flush()
		case event, ok := <-stream:
			if !ok {
				return
			}
			if !started {
				c.Status(http.StatusOK)
				started = true
			}
			if err := writeSSEEvent(c, event); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeSSEEvent(c *gin.Context, event notificationModel.StreamEvent) error {
	payload, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}
	if event.ID != "" {
		if _, err := c.Writer.Write([]byte("id: " + event.ID + "\n")); err != nil {
			return err
		}
	}
	if _, err := c.Writer.Write([]byte("event: " + event.Event + "\n")); err != nil {
		return err
	}
	if _, err := c.Writer.Write([]byte("data: " + string(payload) + "\n\n")); err != nil {
		return err
	}
	return nil
}

func parseChannels(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	channels := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		channels = append(channels, trimmed)
	}
	return liveSubscriptionModel.FilterValidChannels(channels)
}

func (h *Handler) writeStreamError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, liveUpdateUsecase.ErrSubscriptionRequired),
		errors.Is(err, liveUpdateUsecase.ErrInvalidChannel),
		errors.Is(err, liveUpdateUsecase.ErrWorkloadRequired):
		response.GinBadRequest(c, err)
	case errors.Is(err, liveUpdateUsecase.ErrEnvironmentNotFound),
		errors.Is(err, liveUpdateUsecase.ErrWorkloadNotFound):
		response.GinNotFound(c, err)
	case errors.Is(err, errStreamForbidden):
		response.GinError(c, http.StatusForbidden, err)
	case errors.Is(err, liveUpdateUsecase.ErrStreamExpired):
		response.GinError(c, http.StatusGone, err)
	default:
		response.GinInternalServerError(c, err)
	}
}
