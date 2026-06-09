package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	liveSubscriptionModel "github.com/davidsugianto/idp-core/internal/model/live_subscription"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	"github.com/davidsugianto/idp-core/internal/model/workload"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
	liveUpdateUsecase "github.com/davidsugianto/idp-core/internal/usecase/live_update"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type stubEnvironmentUsecase struct {
	getResult *environment.Environment
	getErr    error
}

func (s *stubEnvironmentUsecase) Create(ctx context.Context, teamID string, req environment.CreateEnvironmentRequest) (*environment.Environment, error) {
	return nil, nil
}

func (s *stubEnvironmentUsecase) List(ctx context.Context, teamID string) ([]environment.Environment, error) {
	return nil, nil
}

func (s *stubEnvironmentUsecase) Get(ctx context.Context, teamID, id string) (*environment.Environment, error) {
	return s.getResult, s.getErr
}

func (s *stubEnvironmentUsecase) Delete(ctx context.Context, teamID, id string) error {
	return nil
}

func (s *stubEnvironmentUsecase) GetStatus(ctx context.Context, teamID, id string) (*environment.EnvironmentStatusResponse, error) {
	return nil, nil
}

func (s *stubEnvironmentUsecase) TriggerSync(ctx context.Context, teamID, id string) error {
	return nil
}

func (s *stubEnvironmentUsecase) GetGitOpsStatus(ctx context.Context, teamID, id string) (*environment.ArgoStatus, error) {
	return nil, nil
}

func (s *stubEnvironmentUsecase) GetWorkloads(ctx context.Context, teamID, id string) (*workload.WorkloadStatusResponse, error) {
	return nil, nil
}

func (s *stubEnvironmentUsecase) GetWorkloadDetails(ctx context.Context, teamID, id, workloadName string) (*workload.WorkloadInfo, error) {
	return nil, nil
}

type stubLiveUpdateUsecase struct {
	listNotificationsResult *notificationModel.NotificationListResponse
	listNotificationsErr    error
	streamEvents            <-chan notificationModel.StreamEvent
	streamEventsErr         error
	streamLogs              <-chan notificationModel.StreamEvent
	streamLogsErr           error
}

func (s *stubLiveUpdateUsecase) Subscribe(ctx context.Context, subscription *liveSubscriptionModel.LiveSubscription) error {
	return nil
}

func (s *stubLiveUpdateUsecase) Unsubscribe(ctx context.Context, subscriptionID string) error {
	return nil
}

func (s *stubLiveUpdateUsecase) ListNotifications(ctx context.Context, req *notificationModel.ListNotificationsRequest) (*notificationModel.NotificationListResponse, error) {
	if s.listNotificationsResult != nil || s.listNotificationsErr != nil {
		return s.listNotificationsResult, s.listNotificationsErr
	}
	return &notificationModel.NotificationListResponse{}, nil
}

func (s *stubLiveUpdateUsecase) StreamEvents(ctx context.Context, subscription *liveSubscriptionModel.LiveSubscription) (<-chan notificationModel.StreamEvent, error) {
	return s.streamEvents, s.streamEventsErr
}

func (s *stubLiveUpdateUsecase) StreamLogs(ctx context.Context, req *liveSubscriptionModel.StreamLogsRequest, subscription *liveSubscriptionModel.LiveSubscription) (<-chan notificationModel.StreamEvent, error) {
	return s.streamLogs, s.streamLogsErr
}

func (s *stubLiveUpdateUsecase) PublishNotification(ctx context.Context, notification *notificationModel.Notification) error {
	return nil
}

func (s *stubLiveUpdateUsecase) PublishStatus(ctx context.Context, payload *notificationModel.StatusEventPayload) error {
	return nil
}

func (s *stubLiveUpdateUsecase) PublishProgress(ctx context.Context, payload *notificationModel.ProgressEventPayload) error {
	return nil
}

func newLiveUpdateTestHandler(envUC envUsecase.Usecase, liveUC liveUpdateUsecase.Usecase) *Handler {
	gin.SetMode(gin.TestMode)

	return New(Dependencies{
		EnvironmentUseCase: envUC,
		LiveUpdateUseCase:  liveUC,
		AuthConfig:         &config.AuthConfig{JWTSecret: "test-secret"},
		WebhookValidator:   webhook.NewValidator(),
		AllowedOrigins:     []string{"*"},
	})
}

func TestLiveUpdateHandlerExposesUS3Endpoints(t *testing.T) {
	handlerType := reflect.TypeOf(&Handler{})

	requiredMethods := []string{
		"ListNotifications",
		"StreamEnvironmentEvents",
		"StreamEnvironmentLogs",
	}

	for _, methodName := range requiredMethods {
		t.Run(methodName, func(t *testing.T) {
			_, ok := handlerType.MethodByName(methodName)
			assert.True(t, ok, "expected Handler to expose %s for Phase 3 live update contracts including notification history, SSE streams, and 401/403/410 failure handling", methodName)
		})
	}
}

func TestListNotificationsReturnsUnauthorizedWithoutUser(t *testing.T) {
	handler := newLiveUpdateTestHandler(&stubEnvironmentUsecase{}, &stubLiveUpdateUsecase{})
	router := gin.New()
	router.GET("/v1/notifications", handler.ListNotifications)

	req := httptest.NewRequest(http.MethodGet, "/v1/notifications", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStreamEnvironmentEventsReturnsForbiddenWhenEnvironmentAccessIsLost(t *testing.T) {
	handler := newLiveUpdateTestHandler(&stubEnvironmentUsecase{getErr: envUsecase.ErrEnvironmentNotFound}, &stubLiveUpdateUsecase{})
	router := gin.New()
	router.GET("/v1/environments/:id/events/stream", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("team_id", "team-1")
		handler.StreamEnvironmentEvents(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/environments/env-1/events/stream?channels=status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStreamEnvironmentEventsReturnsGoneWhenSubscriptionIsExpired(t *testing.T) {
	handler := newLiveUpdateTestHandler(
		&stubEnvironmentUsecase{getResult: &environment.Environment{ID: "env-1", TeamID: "team-1"}},
		&stubLiveUpdateUsecase{streamEventsErr: liveUpdateUsecase.ErrStreamExpired},
	)
	router := gin.New()
	router.GET("/v1/environments/:id/events/stream", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("team_id", "team-1")
		handler.StreamEnvironmentEvents(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/environments/env-1/events/stream?channels=status", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusGone, w.Code)
}

func TestStreamEnvironmentLogsReturnsBadRequestWhenWorkloadIsMissing(t *testing.T) {
	handler := newLiveUpdateTestHandler(
		&stubEnvironmentUsecase{getResult: &environment.Environment{ID: "env-1", TeamID: "team-1"}},
		&stubLiveUpdateUsecase{streamLogsErr: liveUpdateUsecase.ErrWorkloadRequired},
	)
	router := gin.New()
	router.GET("/v1/environments/:id/logs/stream", func(c *gin.Context) {
		c.Set("user_id", "user-1")
		c.Set("team_id", "team-1")
		handler.StreamEnvironmentLogs(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/environments/env-1/logs/stream", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
