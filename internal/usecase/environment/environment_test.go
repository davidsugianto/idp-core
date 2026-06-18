package environment

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/davidsugianto/idp-core/internal/mocks"
	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	"github.com/davidsugianto/idp-core/internal/model/environment"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	"github.com/davidsugianto/idp-core/internal/model/workload"
	"github.com/davidsugianto/idp-core/internal/pkg/argocd"
	gitopsRepo "github.com/davidsugianto/idp-core/internal/repository/gitops"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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
		EnvironmentRepo: mockEnvRepo,
		ProvisionerRepo: mockProvRepo,
		GitopsRepo:      mockGitopsRepo,
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
				Name:                "prod-env",
				GitRepoURL:          "https://github.com/org/repo.git",
				ManifestPath:        "manifests",
				ResourceQuotaCPU:    "4",
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

func TestCreateUsesDeliveryTargetClusterServerForArgoDestination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		DeliveryTargetRepo: &fakeDeliveryTargetRepository{targets: map[string]*deliveryTargetModel.DeliveryTarget{
			"target-a": {
				ID:                "target-a",
				TeamID:            "team-123",
				ClusterName:       "cluster-a",
				ClusterServer:     "https://idp-test-control-plane:6443",
				ArgoCDServer:      "https://argo-target-a.example",
				AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
			},
		}},
		GitopsRepo: mockGitopsRepo,
	})

	mockEnvRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil)
	mockGitopsRepo.EXPECT().
		CreateApplication(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, spec argocd.ApplicationSpec) error {
			assert.Equal(t, "https://idp-test-control-plane:6443", spec.ServerURL)
			return nil
		})
	mockEnvRepo.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any(), "team-123", StatusReady, "").
		Return(nil)

	result, err := uc.Create(context.Background(), "team-123", environment.CreateEnvironmentRequest{
		Name:             "dev-env",
		GitRepoURL:       "https://github.com/org/repo.git",
		ManifestPath:     "manifests",
		DeliveryTargetID: "target-a",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "cluster-a", result.ClusterName)
	assert.Equal(t, "https://idp-test-control-plane:6443", result.ClusterServer)
}

func TestCreateUsesRequestClusterServerForArgoDestination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		DeliveryTargetRepo: &fakeDeliveryTargetRepository{targets: map[string]*deliveryTargetModel.DeliveryTarget{
			"target-a": {
				ID:                "target-a",
				TeamID:            "team-123",
				ClusterName:       "cluster-a",
				ClusterServer:     "https://idp-test-control-plane:6443",
				AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
			},
		}},
		GitopsRepo: mockGitopsRepo,
	})

	mockEnvRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil)
	mockGitopsRepo.EXPECT().
		CreateApplication(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, spec argocd.ApplicationSpec) error {
			assert.Equal(t, "https://request-supplied-server.example", spec.ServerURL)
			return nil
		})
	mockEnvRepo.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any(), "team-123", StatusReady, "").
		Return(nil)

	result, err := uc.Create(context.Background(), "team-123", environment.CreateEnvironmentRequest{
		Name:             "dev-env",
		GitRepoURL:       "https://github.com/org/repo.git",
		ManifestPath:     "manifests",
		ClusterServer:    "https://request-supplied-server.example",
		DeliveryTargetID: "target-a",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "https://request-supplied-server.example", result.ClusterServer)
}

func TestCreateDefaultsArgoDestinationServerWhenClusterServerMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		GitopsRepo:      mockGitopsRepo,
	})

	mockEnvRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil)
	mockGitopsRepo.EXPECT().
		CreateApplication(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, spec argocd.ApplicationSpec) error {
			assert.Equal(t, defaultArgoDestinationServer, spec.ServerURL)
			return nil
		})
	mockEnvRepo.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any(), "team-123", StatusReady, "").
		Return(nil)

	result, err := uc.Create(context.Background(), "team-123", environment.CreateEnvironmentRequest{
		Name:         "dev-env",
		GitRepoURL:   "https://github.com/org/repo.git",
		ManifestPath: "manifests",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.ClusterServer)
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
		EnvironmentRepo: mockEnvRepo,
		ProvisionerRepo: mockProvRepo,
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
				mockEnvRepo.EXPECT().
					UpdateLastSync(gomock.Any(), "env-1", gomock.Any()).
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

func TestTriggerSyncUsesResolvedTargetGitOpsRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockDeliveryTargetRepo := &fakeDeliveryTargetRepository{
		targets: map[string]*deliveryTargetModel.DeliveryTarget{
			"target-a": {
				ID:                "target-a",
				TeamID:            "team-123",
				AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
				ControlPlaneName:  "cp-a",
				KubeconfigContext: "ctx-a",
				ArgoCDNamespace:   "argocd-a",
			},
		},
	}
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	providerCalls := 0
	var resolvedTarget *deliveryTargetModel.TargetControlPlane

	uc := New(Dependencies{
		EnvironmentRepo:    mockEnvRepo,
		DeliveryTargetRepo: mockDeliveryTargetRepo,
		GitopsProvider: func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (gitopsRepo.Repository, error) {
			providerCalls++
			resolvedTarget = target
			return mockGitopsRepo, nil
		},
	})

	mockEnvRepo.EXPECT().
		GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
		Return(&environment.Environment{
			ID:               "env-1",
			TeamID:           "team-123",
			DeliveryTargetID: "target-a",
			ArgoAppName:      "env-env-1",
		}, nil)
	mockGitopsRepo.EXPECT().
		SyncApplication(gomock.Any(), "env-env-1").
		Return(nil)
	mockEnvRepo.EXPECT().
		UpdateLastSync(gomock.Any(), "env-1", gomock.Any()).
		Return(nil)

	err := uc.TriggerSync(context.Background(), "team-123", "env-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, providerCalls)
	assert.NotNil(t, resolvedTarget)
	assert.Equal(t, "target-a", resolvedTarget.DeliveryTargetID)
	assert.Equal(t, "cp-a", resolvedTarget.ControlPlaneName)
	assert.Equal(t, "ctx-a", resolvedTarget.KubeconfigContext)
	assert.Equal(t, "argocd-a", resolvedTarget.ArgoCDNamespace)
}

func TestTriggerSyncReturnsTargetResolutionFailures(t *testing.T) {
	tests := []struct {
		name              string
		env               *environment.Environment
		targets           map[string]*deliveryTargetModel.DeliveryTarget
		gitopsError       error
		wantErr           error
		wantRecordFailure bool
		checkError        func(t *testing.T, err error)
	}{
		{
			name: "missing delivery target mapping",
			env: &environment.Environment{
				ID:          "env-1",
				TeamID:      "team-123",
				ArgoAppName: "env-env-1",
			},
			wantErr:           ErrDeliveryTargetNotFound,
			wantRecordFailure: true,
		},
		{
			name: "missing delivery target record",
			env: &environment.Environment{
				ID:               "env-1",
				TeamID:           "team-123",
				DeliveryTargetID: "missing-target",
				ArgoAppName:      "env-env-1",
			},
			targets:           map[string]*deliveryTargetModel.DeliveryTarget{},
			wantErr:           ErrDeliveryTargetNotFound,
			wantRecordFailure: true,
		},
		{
			name: "target outside team scope",
			env: &environment.Environment{
				ID:               "env-1",
				TeamID:           "team-123",
				DeliveryTargetID: "target-a",
				ArgoAppName:      "env-env-1",
			},
			targets: map[string]*deliveryTargetModel.DeliveryTarget{
				"target-a": {
					ID:                "target-a",
					TeamID:            "other-team",
					AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
					ControlPlaneName:  "cp-a",
				},
			},
			wantErr:           ErrTargetAccessDenied,
			wantRecordFailure: true,
		},
		{
			name: "target unavailable for placement",
			env: &environment.Environment{
				ID:               "env-1",
				TeamID:           "team-123",
				DeliveryTargetID: "target-a",
				ArgoAppName:      "env-env-1",
			},
			targets: map[string]*deliveryTargetModel.DeliveryTarget{
				"target-a": {
					ID:                "target-a",
					TeamID:            "team-123",
					AvailabilityState: deliveryTargetModel.AvailabilityDisabled,
					ControlPlaneName:  "cp-a",
				},
			},
			wantErr:           ErrTargetAccessDenied,
			wantRecordFailure: true,
		},
		{
			name: "provider resolution fails",
			env: &environment.Environment{
				ID:               "env-1",
				TeamID:           "team-123",
				DeliveryTargetID: "target-a",
				ArgoAppName:      "env-env-1",
			},
			targets: map[string]*deliveryTargetModel.DeliveryTarget{
				"target-a": {
					ID:                "target-a",
					TeamID:            "team-123",
					AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
					ControlPlaneName:  "cp-a",
				},
			},
			wantErr:           ErrSyncTargetUnavailable,
			wantRecordFailure: true,
			checkError: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "failed to resolve gitops repository")
				assert.ErrorContains(t, err, "provider boom")
			},
		},
		{
			name: "sync failure is recorded",
			env: &environment.Environment{
				ID:               "env-1",
				TeamID:           "team-123",
				DeliveryTargetID: "target-a",
				ArgoAppName:      "env-env-1",
			},
			targets: map[string]*deliveryTargetModel.DeliveryTarget{
				"target-a": {
					ID:                "target-a",
					TeamID:            "team-123",
					AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
					ControlPlaneName:  "cp-a",
				},
			},
			gitopsError:       errors.New("sync token mismatch"),
			wantRecordFailure: true,
			checkError: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "failed to trigger sync")
				assert.ErrorContains(t, err, "sync token mismatch")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
			mockEnvRepo.EXPECT().
				GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
				Return(tt.env, nil)

			mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)
			deliveryTargetRepo := &fakeDeliveryTargetRepository{targets: tt.targets}
			if tt.wantRecordFailure {
				mockEnvRepo.EXPECT().
					RecordSyncFailure(gomock.Any(), "env-1", gomock.Any()).
					Return(nil)
			}
			if tt.gitopsError != nil {
				mockGitopsRepo.EXPECT().
					SyncApplication(gomock.Any(), "env-env-1").
					Return(tt.gitopsError)
			}

			uc := New(Dependencies{
				EnvironmentRepo:    mockEnvRepo,
				DeliveryTargetRepo: deliveryTargetRepo,
				GitopsProvider: func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (gitopsRepo.Repository, error) {
					if tt.gitopsError != nil {
						return mockGitopsRepo, nil
					}
					return nil, errors.New("provider boom")
				},
			})

			err := uc.TriggerSync(context.Background(), "team-123", "env-1")
			assert.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			if tt.checkError != nil {
				tt.checkError(t, err)
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

func TestGetGitOpsStatusUsesResolvedTargetGitOpsRepository(t *testing.T) {
	tests := []struct {
		name                     string
		target                   *deliveryTargetModel.DeliveryTarget
		wantControlPlaneName     string
		wantUsesDefaultSelection bool
	}{
		{
			name: "explicit target control plane",
			target: &deliveryTargetModel.DeliveryTarget{
				ID:                "target-a",
				TeamID:            "team-123",
				AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
				ControlPlaneName:  "cp-a",
				KubeconfigContext: "ctx-a",
				ArgoCDNamespace:   "argocd-a",
			},
			wantControlPlaneName:     "cp-a",
			wantUsesDefaultSelection: false,
		},
		{
			name: "default control plane compatibility",
			target: &deliveryTargetModel.DeliveryTarget{
				ID:                "target-a",
				TeamID:            "team-123",
				AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
			},
			wantControlPlaneName:     "default-cp",
			wantUsesDefaultSelection: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
			mockDeliveryTargetRepo := &fakeDeliveryTargetRepository{
				targets: map[string]*deliveryTargetModel.DeliveryTarget{
					"target-a": tt.target,
				},
			}
			mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

			providerCalls := 0
			var resolvedTarget *deliveryTargetModel.TargetControlPlane

			uc := New(Dependencies{
				EnvironmentRepo:       mockEnvRepo,
				DeliveryTargetRepo:    mockDeliveryTargetRepo,
				DefaultTargetDefaults: deliveryTargetModel.TargetControlPlaneDefaults{ControlPlaneName: "default-cp", KubeconfigContext: "default-ctx", ArgoCDNamespace: "argocd-default"},
				GitopsProvider: func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (gitopsRepo.Repository, error) {
					providerCalls++
					resolvedTarget = target
					return mockGitopsRepo, nil
				},
			})

			mockEnvRepo.EXPECT().
				GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
				Return(&environment.Environment{
					ID:               "env-1",
					TeamID:           "team-123",
					DeliveryTargetID: "target-a",
					ArgoAppName:      "env-env-1",
				}, nil)
			mockGitopsRepo.EXPECT().
				GetApplicationStatus(gomock.Any(), "env-env-1").
				Return(&environment.ArgoStatus{SyncStatus: "Synced", HealthStatus: "Healthy"}, nil)

			status, err := uc.GetGitOpsStatus(context.Background(), "team-123", "env-1")
			assert.NoError(t, err)
			assert.NotNil(t, status)
			assert.Equal(t, 1, providerCalls)
			assert.NotNil(t, resolvedTarget)
			assert.Equal(t, "target-a", resolvedTarget.DeliveryTargetID)
			assert.Equal(t, tt.wantControlPlaneName, resolvedTarget.ControlPlaneName)
			assert.Equal(t, tt.wantUsesDefaultSelection, resolvedTarget.UsesDefaultControlPlane)
		})
	}
}

func TestGetGitOpsStatusReturnsTargetResolutionFailures(t *testing.T) {
	tests := []struct {
		name       string
		env        *environment.Environment
		targets    map[string]*deliveryTargetModel.DeliveryTarget
		provider   GitopsProvider
		wantErr    error
		checkError func(t *testing.T, err error)
	}{
		{
			name: "missing delivery target mapping",
			env: &environment.Environment{
				ID:          "env-1",
				TeamID:      "team-123",
				ArgoAppName: "env-env-1",
			},
			wantErr: ErrDeliveryTargetNotFound,
		},
		{
			name: "missing delivery target record",
			env: &environment.Environment{
				ID:               "env-1",
				TeamID:           "team-123",
				DeliveryTargetID: "missing-target",
				ArgoAppName:      "env-env-1",
			},
			targets: map[string]*deliveryTargetModel.DeliveryTarget{},
			wantErr: ErrDeliveryTargetNotFound,
		},
		{
			name: "provider resolution fails",
			env: &environment.Environment{
				ID:               "env-1",
				TeamID:           "team-123",
				DeliveryTargetID: "target-a",
				ArgoAppName:      "env-env-1",
			},
			targets: map[string]*deliveryTargetModel.DeliveryTarget{
				"target-a": {
					ID:                "target-a",
					TeamID:            "team-123",
					AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
					ControlPlaneName:  "cp-a",
				},
			},
			provider: func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (gitopsRepo.Repository, error) {
				return nil, errors.New("provider boom")
			},
			checkError: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "failed to resolve gitops repository")
				assert.ErrorContains(t, err, "provider boom")
			},
		},
		{
			name: "application lookup failure surfaces",
			env: &environment.Environment{
				ID:               "env-1",
				TeamID:           "team-123",
				DeliveryTargetID: "target-a",
				ArgoAppName:      "env-env-1",
			},
			targets: map[string]*deliveryTargetModel.DeliveryTarget{
				"target-a": {
					ID:                "target-a",
					TeamID:            "team-123",
					AvailabilityState: deliveryTargetModel.AvailabilityAvailable,
					ControlPlaneName:  "cp-a",
				},
			},
			provider: func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (gitopsRepo.Repository, error) {
				return &fakeGitopsRepoWithStatusError{err: errors.New("application not found in resolved control plane")}, nil
			},
			checkError: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "application not found in resolved control plane")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
			mockEnvRepo.EXPECT().
				GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
				Return(tt.env, nil)

			provider := tt.provider
			if provider == nil {
				provider = func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (gitopsRepo.Repository, error) {
					return nil, errors.New("provider boom")
				}
			}

			uc := New(Dependencies{
				EnvironmentRepo:    mockEnvRepo,
				DeliveryTargetRepo: &fakeDeliveryTargetRepository{targets: tt.targets},
				GitopsProvider:     provider,
			})

			status, err := uc.GetGitOpsStatus(context.Background(), "team-123", "env-1")
			assert.Nil(t, status)
			assert.Error(t, err)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
			if tt.checkError != nil {
				tt.checkError(t, err)
			}
		})
	}
}

func TestRecordOperationOutcomeStoresAuditSafeNotification(t *testing.T) {
	notificationUC := &fakeNotificationUsecase{}
	uc := New(Dependencies{
		NotificationUC: notificationUC,
	})

	uc.(*usecase).recordOperationOutcome(context.Background(), &environment.Environment{
		ID:               "env-1",
		TeamID:           "team-123",
		DeliveryTargetID: "target-a",
	}, "gitops_status", "failed", &deliveryTargetModel.TargetControlPlane{
		DeliveryTargetID:        "target-a",
		ControlPlaneName:        "cp-a",
		UsesDefaultControlPlane: false,
	}, errors.New("Get \"https://127.0.0.1:58934/apis/argoproj.io\": dial tcp 127.0.0.1:58934: connect: connection refused: token mismatch on kubeconfig bearer auth"))

	require.Len(t, notificationUC.notifications, 1)
	notification := notificationUC.notifications[0]
	assert.Equal(t, notificationModel.KindEnvironment, notification.Kind)
	assert.Equal(t, notificationModel.SeverityError, notification.Severity)
	assert.NotContains(t, notification.Message, "token")
	assert.NotContains(t, notification.Message, "kubeconfig")
	assert.NotContains(t, notification.Message, "bearer")
	assert.NotContains(t, notification.Message, "https://")
	assert.NotContains(t, notification.Message, "127.0.0.1")
	assert.NotContains(t, notification.Message, "58934")
	assert.Contains(t, notification.Message, "[redacted]")

	var payload environment.TargetResolutionOutcome
	require.NoError(t, json.Unmarshal([]byte(notification.Payload), &payload))
	assert.Equal(t, "gitops_status", payload.Operation)
	assert.Equal(t, "failed", payload.Outcome)
	assert.Equal(t, "env-1", payload.EnvironmentID)
	assert.Equal(t, "target-a", payload.DeliveryTargetID)
	assert.Equal(t, "cp-a", payload.ControlPlaneName)
	assert.False(t, payload.UsesDefaultControlPlane)
	assert.NotContains(t, payload.Error, "token")
	assert.NotContains(t, payload.Error, "kubeconfig")
	assert.NotContains(t, payload.Error, "bearer")
	assert.NotContains(t, payload.Error, "https://")
	assert.NotContains(t, payload.Error, "127.0.0.1")
	assert.NotContains(t, payload.Error, "58934")
	assert.Contains(t, payload.Error, "[redacted]")
}

func TestGetStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	mockProvRepo := mocks.NewMockProvisionerRepository(ctrl)
	mockGitopsRepo := mocks.NewMockGitopsRepository(ctrl)

	uc := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		ProvisionerRepo: mockProvRepo,
		GitopsRepo:      mockGitopsRepo,
	})

	tests := []struct {
		name    string
		teamID  string
		id      string
		setup   func()
		assert  func(t *testing.T, result *environment.EnvironmentStatusResponse, err error)
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
			assert: func(t *testing.T, result *environment.EnvironmentStatusResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, 3, result.PodSummary.Total)
				assert.Equal(t, 2, result.DeploymentSummary.Desired)
			},
		},
		{
			name:   "falls back to live state when cached summaries are unavailable",
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
				mockProvRepo.EXPECT().GetPodSummary("idp-team-123-dev").Return(environment.PodSummary{}, false)
				mockProvRepo.EXPECT().GetDeploymentSummary("idp-team-123-dev").Return(environment.DeploymentSummary{}, false)
				mockProvRepo.EXPECT().GetWorkloads("idp-team-123-dev").Return([]*appsv1.Deployment{{
					ObjectMeta: metav1.ObjectMeta{Name: "nginx", UID: "uid-1"},
					Spec: appsv1.DeploymentSpec{
						Replicas: ptrInt32(2),
						Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Image: "nginx:latest"}}}},
					},
					Status: appsv1.DeploymentStatus{Replicas: 2, ReadyReplicas: 1, UpdatedReplicas: 2, AvailableReplicas: 1},
				}}, nil)
				mockProvRepo.EXPECT().GetPods("idp-team-123-dev").Return([]*corev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{Name: "nginx-pod-1", UID: "pod-1"},
					Status: corev1.PodStatus{Phase: corev1.PodRunning},
				}, {
					ObjectMeta: metav1.ObjectMeta{Name: "nginx-pod-2", UID: "pod-2"},
					Status: corev1.PodStatus{Phase: corev1.PodPending},
				}}, nil)
				mockGitopsRepo.EXPECT().GetApplicationStatus(gomock.Any(), "env-env-1").Return(&environment.ArgoStatus{SyncStatus: "Synced", HealthStatus: "Healthy"}, nil)
			},
			assert: func(t *testing.T, result *environment.EnvironmentStatusResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, environment.PodSummary{Total: 2, Running: 1, Pending: 1, Failed: 0}, result.PodSummary)
				assert.Equal(t, environment.DeploymentSummary{Desired: 2, Ready: 1, Updated: 2, Available: 1}, result.DeploymentSummary)
			},
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
			assert: func(t *testing.T, result *environment.EnvironmentStatusResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.GetStatus(context.Background(), tt.teamID, tt.id)
			tt.assert(t, result, err)
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

	ucNoProv := New(Dependencies{
		EnvironmentRepo: mockEnvRepo,
		ProvisionerRepo: nil,
	})

	tests := []struct {
		name   string
		teamID string
		id     string
		uc     Usecase
		setup  func()
		assert func(t *testing.T, result *workload.WorkloadStatusResponse, err error)
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
					Return([]*appsv1.Deployment{{
						ObjectMeta: metav1.ObjectMeta{Name: "nginx", UID: "uid-1"},
						Spec: appsv1.DeploymentSpec{
							Replicas: ptrInt32(2),
							Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Image: "nginx:latest"}}}},
						},
						Status: appsv1.DeploymentStatus{Replicas: 2, ReadyReplicas: 2, UpdatedReplicas: 2, AvailableReplicas: 2},
					}}, nil)
				mockProvRepo.EXPECT().
					GetPods("idp-team-123-dev").
					Return([]*corev1.Pod{}, nil)
			},
			assert: func(t *testing.T, result *workload.WorkloadStatusResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "env-1", result.EnvironmentID)
				assert.Equal(t, "idp-team-123-dev", result.Namespace)
				assert.Len(t, result.Workloads, 1)
			},
		},
		{
			name:   "empty namespace preserves environment context",
			teamID: "team-123",
			id:     "env-1",
			uc:     uc,
			setup: func() {
				mockEnvRepo.EXPECT().
					GetByIDAndTeam(gomock.Any(), "env-1", "team-123").
					Return(&environment.Environment{
						ID:        "env-1",
						TeamID:    "team-123",
						Namespace: "idp-team-123-empty",
					}, nil)
				mockProvRepo.EXPECT().GetWorkloads("idp-team-123-empty").Return([]*appsv1.Deployment{}, nil)
				mockProvRepo.EXPECT().GetPods("idp-team-123-empty").Return([]*corev1.Pod{}, nil)
			},
			assert: func(t *testing.T, result *workload.WorkloadStatusResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "env-1", result.EnvironmentID)
				assert.Equal(t, "idp-team-123-empty", result.Namespace)
				assert.NotNil(t, result.Workloads)
				assert.Len(t, result.Workloads, 0)
			},
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
			assert: func(t *testing.T, result *workload.WorkloadStatusResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
			},
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
						ID:               "env-1",
						TeamID:           "team-123",
						Namespace:        "idp-team-123-dev",
						DeliveryTargetID: "target-a",
					}, nil)
			},
			assert: func(t *testing.T, result *workload.WorkloadStatusResponse, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.ErrorIs(t, err, ErrWorkloadStateUnavailable)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := tt.uc.GetWorkloads(context.Background(), tt.teamID, tt.id)
			tt.assert(t, result, err)
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
		assert       func(t *testing.T, result *workload.WorkloadInfo, err error)
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
					Return([]*appsv1.Deployment{{
						ObjectMeta: metav1.ObjectMeta{Name: "nginx", UID: "uid-1"},
						Spec: appsv1.DeploymentSpec{
							Replicas: ptrInt32(2),
							Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Image: "nginx:latest"}}}},
						},
						Status: appsv1.DeploymentStatus{Replicas: 2, ReadyReplicas: 2, UpdatedReplicas: 2, AvailableReplicas: 2},
					}}, nil)
				mockProvRepo.EXPECT().GetPods("idp-team-123-dev").Return([]*corev1.Pod{}, nil)
			},
			assert: func(t *testing.T, result *workload.WorkloadInfo, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "nginx", result.Name)
				assert.Equal(t, "nginx:latest", result.Image)
			},
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
				mockProvRepo.EXPECT().GetWorkloads("idp-team-123-dev").Return([]*appsv1.Deployment{}, nil)
				mockProvRepo.EXPECT().GetPods("idp-team-123-dev").Return([]*corev1.Pod{}, nil)
			},
			assert: func(t *testing.T, result *workload.WorkloadInfo, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.ErrorIs(t, err, ErrWorkloadNotFound)
			},
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
			assert: func(t *testing.T, result *workload.WorkloadInfo, err error) {
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.GetWorkloadDetails(context.Background(), tt.teamID, tt.id, tt.workloadName)
			tt.assert(t, result, err)
		})
	}
}

func ptrInt32(v int32) *int32 {
	return &v
}

type fakeGitopsRepoWithStatusError struct {
	err error
}

func (f *fakeGitopsRepoWithStatusError) CreateApplication(ctx context.Context, spec argocd.ApplicationSpec) error {
	return nil
}

func (f *fakeGitopsRepoWithStatusError) GetApplicationStatus(ctx context.Context, name string) (*environment.ArgoStatus, error) {
	return nil, f.err
}

func (f *fakeGitopsRepoWithStatusError) SyncApplication(ctx context.Context, name string) error {
	return nil
}

func (f *fakeGitopsRepoWithStatusError) DeleteApplication(ctx context.Context, name string) error {
	return nil
}

type fakeDeliveryTargetRepository struct {
	targets map[string]*deliveryTargetModel.DeliveryTarget
}

type fakeNotificationUsecase struct {
	notifications []notificationModel.Notification
}

func (f *fakeNotificationUsecase) Create(ctx context.Context, notification *notificationModel.Notification) error {
	if notification != nil {
		f.notifications = append(f.notifications, *notification)
	}
	return nil
}

func (f *fakeNotificationUsecase) Get(ctx context.Context, id string) (*notificationModel.Notification, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeNotificationUsecase) List(ctx context.Context, req *notificationModel.ListNotificationsRequest) (*notificationModel.NotificationListResponse, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeNotificationUsecase) ListByEnvironment(ctx context.Context, environmentID string) ([]notificationModel.Notification, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeNotificationUsecase) ListByUser(ctx context.Context, userID string) ([]notificationModel.Notification, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeNotificationUsecase) Update(ctx context.Context, notification *notificationModel.Notification) error {
	return errors.New("not implemented")
}

func (f *fakeDeliveryTargetRepository) Create(ctx context.Context, target *deliveryTargetModel.DeliveryTarget) error {
	return errors.New("not implemented")
}

func (f *fakeDeliveryTargetRepository) GetByID(ctx context.Context, id string) (*deliveryTargetModel.DeliveryTarget, error) {
	if target, ok := f.targets[id]; ok {
		return target, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeDeliveryTargetRepository) GetControlPlaneByID(ctx context.Context, id string) (*deliveryTargetModel.TargetControlPlane, error) {
	target, err := f.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return target.ControlPlane(), nil
}

func (f *fakeDeliveryTargetRepository) Update(ctx context.Context, target *deliveryTargetModel.DeliveryTarget) error {
	return errors.New("not implemented")
}

func (f *fakeDeliveryTargetRepository) UpdateAvailability(ctx context.Context, id, availabilityState, healthState, capacitySummary string) error {
	return errors.New("not implemented")
}

func (f *fakeDeliveryTargetRepository) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (f *fakeDeliveryTargetRepository) List(ctx context.Context, req *deliveryTargetModel.ListDeliveryTargetsRequest) ([]deliveryTargetModel.DeliveryTarget, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (f *fakeDeliveryTargetRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return false, errors.New("not implemented")
}

func (f *fakeDeliveryTargetRepository) ExistsBySlugExcludingID(ctx context.Context, slug, id string) (bool, error) {
	return false, errors.New("not implemented")
}
