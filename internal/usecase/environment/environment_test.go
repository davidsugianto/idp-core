package environment

import (
	"context"
	"errors"
	"testing"

	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockProvRepo := mocks.NewMockProvisionerRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo:  mockEnvRepo,
		ProvisionerRepo:  mockProvRepo,
		GitopsRepo:       mockGitopsRepo,
	})

	tests := []struct {
		name    string
		teamID  string
		req     environment.CreateEnvironmentRequest
		setup   func()
		wantErr bool
	}{
		{
			name:   "successful creation",
			teamID: "team-123",
			req: environment.CreateEnvironmentRequest{
				Name:         "dev-env",
				GitRepoURL:   "https://github.com/org/repo.git",
				ManifestPath: "manifests",
				GitRevision:  "main",
			},
			setup: func() {
				mockEnvRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				mockEnvRepo.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), "team-123", StatusReady, "").
					Return(nil)
				mockProvRepo.EXPECT().
					CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				mockProvRepo.EXPECT().
					CreateNetworkPolicy(gomock.Any(), gomock.Any(), "idp-isolation", gomock.Any()).
					Return(nil)
				mockGitopsRepo.EXPECT().
					CreateApplication(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "creation with resource quota",
			teamID: "team-123",
			req: environment.CreateEnvironmentRequest{
				Name:              "prod-env",
				GitRepoURL:        "https://github.com/org/repo.git",
				ManifestPath:      "manifests",
				ResourceQuotaCPU:  "4",
				ResourceQuotaMemory: "8Gi",
			},
			setup: func() {
				mockEnvRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				mockEnvRepo.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), "team-123", StatusReady, "").
					Return(nil)
				mockProvRepo.EXPECT().
					CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				mockProvRepo.EXPECT().
					CreateResourceQuota(gomock.Any(), gomock.Any(), "idp-quota", "4", "8Gi").
					Return(nil)
				mockProvRepo.EXPECT().
					CreateNetworkPolicy(gomock.Any(), gomock.Any(), "idp-isolation", gomock.Any()).
					Return(nil)
				mockGitopsRepo.EXPECT().
					CreateApplication(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "db creation fails",
			teamID: "team-123",
			req: environment.CreateEnvironmentRequest{
				Name:         "dev-env",
				GitRepoURL:   "https://github.com/org/repo.git",
				ManifestPath: "manifests",
			},
			setup: func() {
				mockEnvRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:   "namespace creation fails",
			teamID: "team-123",
			req: environment.CreateEnvironmentRequest{
				Name:         "dev-env",
				GitRepoURL:   "https://github.com/org/repo.git",
				ManifestPath: "manifests",
			},
			setup: func() {
				mockEnvRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				mockProvRepo.EXPECT().
					CreateNamespace(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("k8s error"))
				mockEnvRepo.EXPECT().
					UpdateStatus(gomock.Any(), gomock.Any(), "team-123", StatusFailed, gomock.Any()).
					Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.Create(context.Background(), tt.teamID, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.teamID, result.TeamID)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.NotEmpty(t, result.Namespace)
			}
		})
	}
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
	})

	tests := []struct {
		name    string
		teamID  string
		setup   func()
		wantErr bool
		wantLen int
	}{
		{
			name:   "list environments successfully",
			teamID: "team-123",
			setup: func() {
				mockEnvRepo.EXPECT().
					ListByTeam(gomock.Any(), "team-123").
					Return([]environment.Environment{
						{ID: "env-1", TeamID: "team-123", Name: "dev"},
						{ID: "env-2", TeamID: "team-123", Name: "staging"},
					}, nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:   "empty list",
			teamID: "team-empty",
			setup: func() {
				mockEnvRepo.EXPECT().
					ListByTeam(gomock.Any(), "team-empty").
					Return([]environment.Environment{}, nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:   "db error",
			teamID: "team-error",
			setup: func() {
				mockEnvRepo.EXPECT().
					ListByTeam(gomock.Any(), "team-error").
					Return(nil, errors.New("db error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.List(context.Background(), tt.teamID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLen)
			}
		})
	}
}

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
	})

	tests := []struct {
		name    string
		teamID  string
		id      string
		setup   func()
		wantErr bool
	}{
		{
			name:   "get environment successfully",
			teamID: "team-123",
			id:     "env-1",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:     "env-1",
						TeamID: "team-123",
						Name:   "dev",
					}, nil)
			},
			wantErr: false,
		},
		{
			name:   "environment not found",
			teamID: "team-123",
			id:     "nonexistent",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.Get(context.Background(), tt.teamID, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockProvRepo := mocks.NewMockProvisionerRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo:  mockEnvRepo,
		ProvisionerRepo:  mockProvRepo,
		GitopsRepo:       mockGitopsRepo,
	})

	tests := []struct {
		name    string
		teamID  string
		id      string
		setup   func()
		wantErr bool
	}{
		{
			name:   "delete successfully",
			teamID: "team-123",
			id:     "env-1",
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
					UpdateStatus(gomock.Any(), "env-1", "team-123", StatusDeleting, "").
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
			wantErr: false,
		},
		{
			name:   "environment not found",
			teamID: "team-123",
			id:     "nonexistent",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := uc.Delete(context.Background(), tt.teamID, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTriggerSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		GitopsRepo:      mockGitopsRepo,
	})

	tests := []struct {
		name    string
		teamID  string
		id      string
		setup   func()
		wantErr bool
	}{
		{
			name:   "sync successfully",
			teamID: "team-123",
			id:     "env-1",
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
			wantErr: false,
		},
		{
			name:   "no argo app name",
			teamID: "team-123",
			id:     "env-1",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:          "env-1",
						TeamID:      "team-123",
						ArgoAppName: "",
					}, nil)
			},
			wantErr: true,
		},
		{
			name:   "environment not found",
			teamID: "team-123",
			id:     "nonexistent",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := uc.TriggerSync(context.Background(), tt.teamID, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetGitOpsStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		GitopsRepo:      mockGitopsRepo,
	})

	tests := []struct {
		name    string
		teamID  string
		id      string
		setup   func()
		wantErr bool
	}{
		{
			name:   "get status successfully",
			teamID: "team-123",
			id:     "env-1",
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
			wantErr: false,
		},
		{
			name:   "environment not found",
			teamID: "team-123",
			id:     "nonexistent",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.GetGitOpsStatus(context.Background(), tt.teamID, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGetStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockProvRepo := mocks.NewMockProvisionerRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo:  mockEnvRepo,
		ProvisionerRepo:  mockProvRepo,
		GitopsRepo:       mockGitopsRepo,
	})

	tests := []struct {
		name    string
		teamID  string
		id      string
		setup   func()
		wantErr bool
	}{
		{
			name:   "get status successfully",
			teamID: "team-123",
			id:     "env-1",
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
					Return(environment.PodSummary{Total: 3, Running: 2, Pending: 1}, true)
				mockProvRepo.EXPECT().
					GetDeploymentSummary("idp-team-123-dev").
					Return(environment.DeploymentSummary{Desired: 2, Ready: 2}, true)
				mockGitopsRepo.EXPECT().
					GetApplicationStatus(gomock.Any(), "env-env-1").
					Return(&environment.ArgoStatus{
						SyncStatus:   "Synced",
						HealthStatus: "Healthy",
					}, nil)
			},
			wantErr: false,
		},
		{
			name:   "environment not found",
			teamID: "team-123",
			id:     "nonexistent",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.GetStatus(context.Background(), tt.teamID, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGenerateNamespace(t *testing.T) {
	tests := []struct {
		name     string
		teamID   string
		envName  string
		expected string
	}{
		{
			name:     "normal names",
			teamID:   "myteam",
			envName:  "dev",
			expected: "idp-myteam-dev",
		},
		{
			name:     "uppercase converted to lowercase",
			teamID:   "MyTeam",
			envName:  "Dev",
			expected: "idp-myteam-dev",
		},
		{
			name:     "special characters removed",
			teamID:   "my-team_123",
			envName:  "dev!env",
			expected: "idp-my-team-123-dev-env",
		},
		{
			name:     "long names truncated",
			teamID:   "verylongteamnameexceedinglimit",
			envName:  "verylongenvironmentnameexceedinglimit",
			expected: "idp-verylongteamnameexce-verylongenvironmentnameexceedi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateNamespace(tt.teamID, tt.envName)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), 63)
		})
	}
}

func TestGetWorkloads(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockProvRepo := mocks.NewMockProvisionerRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		ProvisionerRepo: mockProvRepo,
	})

	// Usecase without provisioner for testing "not configured" case
	ucNoProv := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		ProvisionerRepo: nil,
	})

	tests := []struct {
		name    string
		teamID  string
		id      string
		uc      Usecase
		setup   func()
		wantErr bool
	}{
		{
			name:   "get workloads successfully",
			teamID: "team-123",
			id:     "env-1",
			uc:     uc,
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
					Return([]*appsv1.Deployment{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "nginx",
								UID:  "uid-1",
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: ptrInt32(2),
								Template: corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{Image: "nginx:latest"},
										},
									},
								},
							},
							Status: appsv1.DeploymentStatus{
								Replicas:          2,
								ReadyReplicas:     2,
								UpdatedReplicas:   2,
								AvailableReplicas: 2,
							},
						},
					}, nil)
				mockProvRepo.EXPECT().
					GetPods("idp-team-123-dev").
					Return([]*corev1.Pod{}, nil)
			},
			wantErr: false,
		},
		{
			name:   "environment not found",
			teamID: "team-123",
			id:     "nonexistent",
			uc:     uc,
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:   "kubernetes not configured",
			teamID: "team-123",
			id:     "env-1",
			uc:     ucNoProv,
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:        "env-1",
						TeamID:    "team-123",
						Namespace: "idp-team-123-dev",
					}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := tt.uc.GetWorkloads(context.Background(), tt.teamID, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGetWorkloadDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockProvRepo := mocks.NewMockProvisionerRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		ProvisionerRepo: mockProvRepo,
	})

	tests := []struct {
		name         string
		teamID       string
		id           string
		workloadName string
		setup        func()
		wantErr      bool
	}{
		{
			name:         "get workload details successfully",
			teamID:       "team-123",
			id:           "env-1",
			workloadName: "nginx",
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
					Return([]*appsv1.Deployment{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "nginx",
								UID:  "uid-1",
							},
							Spec: appsv1.DeploymentSpec{
								Replicas: ptrInt32(2),
								Template: corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{Image: "nginx:latest"},
										},
									},
								},
							},
							Status: appsv1.DeploymentStatus{
								Replicas:        2,
								ReadyReplicas:   2,
								UpdatedReplicas: 2,
								AvailableReplicas: 2,
							},
						},
					}, nil)
			},
			wantErr: false,
		},
		{
			name:         "workload not found",
			teamID:       "team-123",
			id:           "env-1",
			workloadName: "nonexistent",
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
					Return([]*appsv1.Deployment{}, nil)
			},
			wantErr: true,
		},
		{
			name:         "environment not found",
			teamID:       "team-123",
			id:           "nonexistent",
			workloadName: "nginx",
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "nonexistent", "team-123").
					Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.GetWorkloadDetails(context.Background(), tt.teamID, tt.id, tt.workloadName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.workloadName, result.Name)
			}
		})
	}
}

func ptrInt32(v int32) *int32 {
	return &v
}
