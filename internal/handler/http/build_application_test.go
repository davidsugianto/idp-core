package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	buildApplicationUsecase "github.com/davidsugianto/idp-core/internal/usecase/build_application"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeBuildApplicationUsecase struct {
	createApplicationFn func(teamID, actorID string, req *buildApplicationModel.CreateBuildApplicationRequest) (*buildApplicationModel.BuildApplicationResponse, error)
	getBuildFn          func(teamID, buildID string) (*buildApplicationModel.BuildResponse, error)
}

func (f *fakeBuildApplicationUsecase) CreateApplication(ctx context.Context, teamID, actorID string, req *buildApplicationModel.CreateBuildApplicationRequest) (*buildApplicationModel.BuildApplicationResponse, error) {
	if f.createApplicationFn != nil {
		return f.createApplicationFn(teamID, actorID, req)
	}
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) ListApplications(ctx context.Context, teamID string, req *buildApplicationModel.ListBuildApplicationsRequest) (*buildApplicationModel.BuildApplicationListResponse, error) {
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) GetApplication(ctx context.Context, teamID, applicationID string) (*buildApplicationModel.BuildApplicationResponse, error) {
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) UpdateApplication(ctx context.Context, teamID, actorID, applicationID string, req *buildApplicationModel.UpdateBuildApplicationRequest) (*buildApplicationModel.BuildApplicationResponse, error) {
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) DeleteApplication(ctx context.Context, teamID, actorID, applicationID string) error {
	return nil
}

func (f *fakeBuildApplicationUsecase) TriggerBuild(ctx context.Context, teamID, actorID, applicationID string, req *buildApplicationModel.TriggerBuildRequest) (*buildApplicationModel.BuildActionResponse, error) {
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) GetBuild(ctx context.Context, teamID, buildID string) (*buildApplicationModel.BuildResponse, error) {
	if f.getBuildFn != nil {
		return f.getBuildFn(teamID, buildID)
	}
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) ListBuilds(ctx context.Context, teamID, applicationID string, limit, offset int) (*buildApplicationModel.BuildHistoryResponse, error) {
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) RetryBuild(ctx context.Context, teamID, actorID, buildID string) (*buildApplicationModel.BuildActionResponse, error) {
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) CancelBuild(ctx context.Context, teamID, actorID, buildID string) (*buildApplicationModel.BuildActionResponse, error) {
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) StreamBuildLogs(ctx context.Context, teamID, buildID string, afterSequence int64, limit int) (*buildApplicationModel.BuildLogStreamResponse, error) {
	return nil, nil
}

func (f *fakeBuildApplicationUsecase) DrainQueuedBuilds(ctx context.Context) error {
	return nil
}

func setupBuildApplicationTestHandler(fakeUC buildApplicationUsecase.Usecase) *Handler {
	return New(Dependencies{
		BuildApplicationUseCase: fakeUC,
		WebhookValidator:        webhook.NewValidator(),
	})
}

func TestBuildApplicationHandlerCreateUnauthorized(t *testing.T) {
	h := setupBuildApplicationTestHandler(&fakeBuildApplicationUsecase{})
	router := gin.New()
	router.POST("/v1/build-applications", h.CreateBuildApplication)

	body := map[string]any{"name": "app", "repository_url": "https://example.com/repo.git", "application_descriptor_path": "application.yaml"}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/build-applications", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestBuildApplicationHandlerCreateSuccess(t *testing.T) {
	fakeUC := &fakeBuildApplicationUsecase{
		createApplicationFn: func(teamID, actorID string, req *buildApplicationModel.CreateBuildApplicationRequest) (*buildApplicationModel.BuildApplicationResponse, error) {
			return &buildApplicationModel.BuildApplicationResponse{ID: "app-1", TeamID: teamID, Name: req.Name, RepositoryURL: req.RepositoryURL, ApplicationDescriptorPath: req.ApplicationDescriptorPath}, nil
		},
	}
	h := setupBuildApplicationTestHandler(fakeUC)
	router := gin.New()
	router.POST("/v1/build-applications", func(c *gin.Context) {
		c.Set("team_id", "team-1")
		c.Set("user_id", "user-1")
		c.Next()
	}, h.CreateBuildApplication)

	body := map[string]any{"name": "app", "repository_url": "https://example.com/repo.git", "application_descriptor_path": "application.yaml"}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/build-applications", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestBuildApplicationHandlerGetBuildNotFound(t *testing.T) {
	fakeUC := &fakeBuildApplicationUsecase{
		getBuildFn: func(teamID, buildID string) (*buildApplicationModel.BuildResponse, error) {
			return nil, buildApplicationUsecase.ErrBuildNotFound
		},
	}
	h := setupBuildApplicationTestHandler(fakeUC)
	router := gin.New()
	router.GET("/v1/builds/:buildId", func(c *gin.Context) {
		c.Set("team_id", "team-1")
		c.Next()
	}, h.GetBuild)

	req := httptest.NewRequest(http.MethodGet, "/v1/builds/build-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
