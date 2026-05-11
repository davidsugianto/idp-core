package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	envUsecase "github.com/davidsugianto/idp-core/internal/usecase/environment"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestHandler(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository, *mocks.MockProvisionerRepository, *mocks.MockGitopsRepository) {
	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockProvRepo := mocks.NewMockProvisionerRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := envUsecase.New(envUsecase.Dependencies{
		EnvironmentRepo:  mockEnvRepo,
		ProvisionerRepo:  mockProvRepo,
		GitopsRepo:       mockGitopsRepo,
	})

	handler := New(Dependencies{
		EnvironmentUseCase: uc,
		AuthConfig:         &config.AuthConfig{JWTSecret: "test-secret"},
		WebhookValidator:   webhook.NewValidator(),
	})

	return handler, mockEnvRepo, mockProvRepo, mockGitopsRepo
}

func setupTestRouter() *gin.Engine {
	return gin.New()
}

func withTeamID(c *gin.Context, teamID string) {
	c.Set("team_id", teamID)
}

func TestCreateEnvironment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockEnvRepo, mockProvRepo, mockGitopsRepo := setupTestHandler(ctrl)

	tests := []struct {
		name       string
		teamID     string
		body       interface{}
		setup      func()
		wantStatus int
	}{
		{
			name:   "successful creation",
			teamID: "team-123",
			body: environment.CreateEnvironmentRequest{
				Name:         "dev",
				GitRepoURL:   "https://github.com/org/repo.git",
				ManifestPath: "manifests",
			},
			setup: func() {
				mockEnvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				mockEnvRepo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), "team-123", "ready", "").Return(nil)
				mockProvRepo.EXPECT().CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockProvRepo.EXPECT().CreateNetworkPolicy(gomock.Any(), gomock.Any(), "idp-isolation", gomock.Any()).Return(nil)
				mockGitopsRepo.EXPECT().CreateApplication(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing team id",
			teamID:     "",
			body:       environment.CreateEnvironmentRequest{Name: "dev"},
			setup:      func() {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid request body",
			teamID:     "team-123",
			body:       map[string]interface{}{"name": 123}, // invalid type
			setup:      func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "missing required fields",
			teamID: "team-123",
			body:   environment.CreateEnvironmentRequest{Name: ""}, // missing required fields
			setup:  func() {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			router := setupTestRouter()
			router.POST("/environments", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.CreateEnvironment)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/environments", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestListEnvironments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)

	tests := []struct {
		name       string
		teamID     string
		setup      func()
		wantStatus int
		wantLen    int
	}{
		{
			name:   "list successfully",
			teamID: "team-123",
			setup: func() {
				mockEnvRepo.EXPECT().
					ListByTeam(gomock.Any(), "team-123").
					Return([]environment.Environment{
						{ID: "env-1", TeamID: "team-123", Name: "dev"},
						{ID: "env-2", TeamID: "team-123", Name: "staging"},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:       "missing team id",
			teamID:     "",
			setup:      func() {},
			wantStatus: http.StatusUnauthorized,
			wantLen:    0,
		},
		{
			name:   "empty list",
			teamID: "team-empty",
			setup: func() {
				mockEnvRepo.EXPECT().
					ListByTeam(gomock.Any(), "team-empty").
					Return([]environment.Environment{}, nil)
			},
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			router := setupTestRouter()
			router.GET("/environments", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.ListEnvironments)

			req := httptest.NewRequest(http.MethodGet, "/environments", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				data, ok := response["data"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, data, tt.wantLen)
			}
		})
	}
}

func TestGetEnvironment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)

	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func()
		wantStatus int
	}{
		{
			name:   "get successfully",
			teamID: "team-123",
			envID:  "env-1",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:     "env-1",
						TeamID: "team-123",
						Name:   "dev",
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing team id",
			teamID:     "",
			envID:      "env-1",
			setup:      func() {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "environment not found",
			teamID: "team-123",
			envID:  "nonexistent",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			router := setupTestRouter()
			router.GET("/environments/:id", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.GetEnvironment)

			req := httptest.NewRequest(http.MethodGet, "/environments/"+tt.envID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDeleteEnvironment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockEnvRepo, mockProvRepo, mockGitopsRepo := setupTestHandler(ctrl)

	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func()
		wantStatus int
	}{
		{
			name:   "delete successfully",
			teamID: "team-123",
			envID:  "env-1",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:          "env-1",
						TeamID:      "team-123",
						Namespace:   "idp-team-123-dev",
						ArgoAppName: "env-env-1",
					}, nil)
				mockEnvRepo.EXPECT().
					UpdateStatus(gomock.Any(), "env-1", "team-123", "deleting", "").
					Return(nil)
				mockGitopsRepo.EXPECT().
					DeleteApplication(gomock.Any(), "env-env-1").
					Return(nil)
				mockProvRepo.EXPECT().
					DeleteNamespace(gomock.Any(), "idp-team-123-dev").
					Return(nil)
				mockEnvRepo.EXPECT().
					SoftDelete(gomock.Any(), "env-1", "team-123").
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing team id",
			teamID:     "",
			envID:      "env-1",
			setup:      func() {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "environment not found",
			teamID: "team-123",
			envID:  "nonexistent",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			router := setupTestRouter()
			router.DELETE("/environments/:id", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.DeleteEnvironment)

			req := httptest.NewRequest(http.MethodDelete, "/environments/"+tt.envID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestSyncEnvironment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockEnvRepo, _, mockGitopsRepo := setupTestHandler(ctrl)

	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func()
		wantStatus int
	}{
		{
			name:   "sync successfully",
			teamID: "team-123",
			envID:  "env-1",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:          "env-1",
						TeamID:      "team-123",
						ArgoAppName: "env-env-1",
					}, nil)
				mockGitopsRepo.EXPECT().
					SyncApplication(gomock.Any(), "env-env-1").
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing team id",
			teamID:     "",
			envID:      "env-1",
			setup:      func() {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "environment not found",
			teamID: "team-123",
			envID:  "nonexistent",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			router := setupTestRouter()
			router.POST("/environments/:id/sync", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.SyncEnvironment)

			req := httptest.NewRequest(http.MethodPost, "/environments/"+tt.envID+"/sync", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetEnvironmentStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockEnvRepo, mockProvRepo, mockGitopsRepo := setupTestHandler(ctrl)

	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func()
		wantStatus int
	}{
		{
			name:   "get status successfully",
			teamID: "team-123",
			envID:  "env-1",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:          "env-1",
						TeamID:      "team-123",
						Namespace:   "idp-team-123-dev",
						ArgoAppName: "env-env-1",
					}, nil)
				mockProvRepo.EXPECT().
					GetPodSummary("idp-team-123-dev").
					Return(environment.PodSummary{Total: 2, Running: 2}, true)
				mockProvRepo.EXPECT().
					GetDeploymentSummary("idp-team-123-dev").
					Return(environment.DeploymentSummary{Desired: 1, Ready: 1}, true)
				mockGitopsRepo.EXPECT().
					GetApplicationStatus(gomock.Any(), "env-env-1").
					Return(&environment.ArgoStatus{SyncStatus: "Synced"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing team id",
			teamID:     "",
			envID:      "env-1",
			setup:      func() {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			router := setupTestRouter()
			router.GET("/environments/:id/status", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.GetEnvironmentStatus)

			req := httptest.NewRequest(http.MethodGet, "/environments/"+tt.envID+"/status", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetGitOpsStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockEnvRepo, _, mockGitopsRepo := setupTestHandler(ctrl)

	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func()
		wantStatus int
	}{
		{
			name:   "get gitops status successfully",
			teamID: "team-123",
			envID:  "env-1",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:          "env-1",
						TeamID:      "team-123",
						ArgoAppName: "env-env-1",
					}, nil)
				mockGitopsRepo.EXPECT().
					GetApplicationStatus(gomock.Any(), "env-env-1").
					Return(&environment.ArgoStatus{
						SyncStatus:   "Synced",
						HealthStatus: "Healthy",
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing team id",
			teamID:     "",
			envID:      "env-1",
			setup:      func() {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			router := setupTestRouter()
			router.GET("/environments/:id/gitops/status", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.GetGitOpsStatus)

			req := httptest.NewRequest(http.MethodGet, "/environments/"+tt.envID+"/gitops/status", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetWorkloads(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockEnvRepo, mockProvRepo, _ := setupTestHandler(ctrl)

	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func()
		wantStatus int
	}{
		{
			name:   "get workloads successfully",
			teamID: "team-123",
			envID:  "env-1",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:        "env-1",
						TeamID:    "team-123",
						Namespace: "idp-team-123-dev",
					}, nil)
				mockProvRepo.EXPECT().
					GetWorkloads("idp-team-123-dev").
					Return(nil, nil)
				mockProvRepo.EXPECT().
					GetPods("idp-team-123-dev").
					Return(nil, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing team id",
			teamID:     "",
			envID:      "env-1",
			setup:      func() {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			router := setupTestRouter()
			router.GET("/environments/:id/workloads", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.GetWorkloads)

			req := httptest.NewRequest(http.MethodGet, "/environments/"+tt.envID+"/workloads", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _, _, _ := setupTestHandler(ctrl)

	router := setupTestRouter()
	router.GET("/ping", handler.Ping)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "ok", data["status"])
}

func TestGetWorkloadDetails(t *testing.T) {
	tests := []struct {
		name       string
		teamID     string
		envID      string
		workload   string
		setup      func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository, *mocks.MockProvisionerRepository, *mocks.MockGitopsRepository)
		wantStatus int
	}{
		{
			name:     "get workload details successfully",
			teamID:   "team-123",
			envID:    "env-1",
			workload: "my-deployment",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository, *mocks.MockProvisionerRepository, *mocks.MockGitopsRepository) {
				handler, mockEnvRepo, mockProvRepo, mockGitopsRepo := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:        "env-1",
						TeamID:    "team-123",
						Namespace: "idp-team-123-dev",
					}, nil)
				mockProvRepo.EXPECT().
					GetWorkloads("idp-team-123-dev").
					Return([]*appsv1.Deployment{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "my-deployment"},
							Spec: appsv1.DeploymentSpec{
								Replicas: ptrInt32(1),
								Template: corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{Name: "app", Image: "nginx:latest"},
										},
									},
								},
							},
							Status: appsv1.DeploymentStatus{
								ReadyReplicas: 1,
							},
						},
					}, nil)
				return handler, mockEnvRepo, mockProvRepo, mockGitopsRepo
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "missing team id",
			teamID:   "",
			envID:    "env-1",
			workload: "my-deployment",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository, *mocks.MockProvisionerRepository, *mocks.MockGitopsRepository) {
				return setupTestHandler(ctrl)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "environment not found",
			teamID:   "team-123",
			envID:    "nonexistent",
			workload: "my-deployment",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository, *mocks.MockProvisionerRepository, *mocks.MockGitopsRepository) {
				handler, mockEnvRepo, mockProvRepo, mockGitopsRepo := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
				return handler, mockEnvRepo, mockProvRepo, mockGitopsRepo
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "workload not found",
			teamID:   "team-123",
			envID:    "env-1",
			workload: "nonexistent-workload",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository, *mocks.MockProvisionerRepository, *mocks.MockGitopsRepository) {
				handler, mockEnvRepo, mockProvRepo, mockGitopsRepo := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:        "env-1",
						TeamID:    "team-123",
						Namespace: "idp-team-123-dev",
					}, nil)
				mockProvRepo.EXPECT().
					GetWorkloads("idp-team-123-dev").
					Return([]*appsv1.Deployment{}, nil)
				return handler, mockEnvRepo, mockProvRepo, mockGitopsRepo
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, _, _, _ := tt.setup(ctrl)

			router := setupTestRouter()
			router.GET("/environments/:id/workloads/:name", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.GetWorkloadDetails)

			req := httptest.NewRequest(http.MethodGet, "/environments/"+tt.envID+"/workloads/"+tt.workload, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func ptrInt32(i int32) *int32 {
	return &i
}

func TestGetEnvironmentStatus_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository)
		wantStatus int
	}{
		{
			name:   "environment not found",
			teamID: "team-123",
			envID:  "nonexistent",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository) {
				handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
				return handler, mockEnvRepo
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "repository error",
			teamID: "team-123",
			envID:  "env-1",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository) {
				handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(nil, assert.AnError)
				return handler, mockEnvRepo
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, _ := tt.setup(ctrl)

			router := setupTestRouter()
			router.GET("/environments/:id/status", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.GetEnvironmentStatus)

			req := httptest.NewRequest(http.MethodGet, "/environments/"+tt.envID+"/status", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetGitOpsStatus_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository)
		wantStatus int
	}{
		{
			name:   "environment not found",
			teamID: "team-123",
			envID:  "nonexistent",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository) {
				handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
				return handler, mockEnvRepo
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "repository error",
			teamID: "team-123",
			envID:  "env-1",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository) {
				handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(nil, assert.AnError)
				return handler, mockEnvRepo
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, _ := tt.setup(ctrl)

			router := setupTestRouter()
			router.GET("/environments/:id/gitops/status", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.GetGitOpsStatus)

			req := httptest.NewRequest(http.MethodGet, "/environments/"+tt.envID+"/gitops/status", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetWorkloads_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository)
		wantStatus int
	}{
		{
			name:   "environment not found",
			teamID: "team-123",
			envID:  "nonexistent",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository) {
				handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
				return handler, mockEnvRepo
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "repository error",
			teamID: "team-123",
			envID:  "env-1",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository) {
				handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(nil, assert.AnError)
				return handler, mockEnvRepo
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, _ := tt.setup(ctrl)

			router := setupTestRouter()
			router.GET("/environments/:id/workloads", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.GetWorkloads)

			req := httptest.NewRequest(http.MethodGet, "/environments/"+tt.envID+"/workloads", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestDeleteEnvironment_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository)
		wantStatus int
	}{
		{
			name:   "repository error",
			teamID: "team-123",
			envID:  "env-1",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository) {
				handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(nil, assert.AnError)
				return handler, mockEnvRepo
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, _ := tt.setup(ctrl)

			router := setupTestRouter()
			router.DELETE("/environments/:id", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.DeleteEnvironment)

			req := httptest.NewRequest(http.MethodDelete, "/environments/"+tt.envID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestSyncEnvironment_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		teamID     string
		envID      string
		setup      func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository)
		wantStatus int
	}{
		{
			name:   "repository error",
			teamID: "team-123",
			envID:  "env-1",
			setup: func(ctrl *gomock.Controller) (*Handler, *mocks.MockEnvironmentRepository) {
				handler, mockEnvRepo, _, _ := setupTestHandler(ctrl)
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(nil, assert.AnError)
				return handler, mockEnvRepo
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, _ := tt.setup(ctrl)

			router := setupTestRouter()
			router.POST("/environments/:id/sync", func(c *gin.Context) {
				if tt.teamID != "" {
					withTeamID(c, tt.teamID)
				}
				c.Next()
			}, handler.SyncEnvironment)

			req := httptest.NewRequest(http.MethodPost, "/environments/"+tt.envID+"/sync", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
