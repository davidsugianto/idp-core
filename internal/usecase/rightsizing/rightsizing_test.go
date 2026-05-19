package rightsizing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/model/rightsizing"
	prometheusPkg "github.com/davidsugianto/idp-core/internal/pkg/prometheus"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListRecommendations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRightsizingRepository(ctrl)
	uc := New(Dependencies{
		RightsizingRepo: mockRepo,
	})

	t.Run("returns recommendations", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]rightsizing.RightsizingRecommendation{
				{
					ID:                "rec-1",
					Namespace:         "default",
					WorkloadName:      "api-server",
					WorkloadType:      rightsizing.WorkloadTypeDeployment,
					ContainerName:     "main",
					RecommendationType: rightsizing.RecommendationTypeScaleDown,
					Status:            rightsizing.StatusPending,
					CreatedAt:         now,
					UpdatedAt:         now,
					AnalysisPeriodStart: now.Add(-24 * time.Hour),
					AnalysisPeriodEnd:   now,
				},
			}, int64(1), nil)

		resp, err := uc.ListRecommendations(context.Background(), &rightsizing.ListRecommendationsRequest{})
		assert.NoError(t, err)
		assert.Len(t, resp.Recommendations, 1)
		assert.Equal(t, int64(1), resp.Total)
		assert.Equal(t, "rec-1", resp.Recommendations[0].ID)
	})

	t.Run("returns empty list", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]rightsizing.RightsizingRecommendation{}, int64(0), nil)

		resp, err := uc.ListRecommendations(context.Background(), &rightsizing.ListRecommendationsRequest{})
		assert.NoError(t, err)
		assert.Empty(t, resp.Recommendations)
		assert.Equal(t, int64(0), resp.Total)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return(nil, int64(0), errors.New("db error"))

		resp, err := uc.ListRecommendations(context.Background(), &rightsizing.ListRecommendationsRequest{})
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "db error")
	})
}

func TestGetRecommendation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRightsizingRepository(ctrl)
	uc := New(Dependencies{
		RightsizingRepo: mockRepo,
	})

	t.Run("returns recommendation by id", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-1").
			Return(&rightsizing.RightsizingRecommendation{
				ID:                "rec-1",
				Namespace:         "default",
				WorkloadName:      "api-server",
				WorkloadType:      rightsizing.WorkloadTypeDeployment,
				ContainerName:     "main",
				RecommendationType: rightsizing.RecommendationTypeScaleDown,
				Status:            rightsizing.StatusPending,
				CreatedAt:         now,
				UpdatedAt:         now,
				AnalysisPeriodStart: now.Add(-24 * time.Hour),
				AnalysisPeriodEnd:   now,
			}, nil)

		resp, err := uc.GetRecommendation(context.Background(), "rec-1")
		assert.NoError(t, err)
		assert.Equal(t, "rec-1", resp.ID)
		assert.Equal(t, "default", resp.Namespace)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, errors.New("not found"))

		resp, err := uc.GetRecommendation(context.Background(), "nonexistent")
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "not found")
	})
}

func TestApplyRecommendation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRightsizingRepository(ctrl)
	mockProvisioner := mocks.NewMockProvisionerRepository(ctrl)
	uc := New(Dependencies{
		RightsizingRepo:  mockRepo,
		ProvisionerRepo:  mockProvisioner,
	})

	now := time.Now()
	pendingRec := &rightsizing.RightsizingRecommendation{
		ID:                   "rec-1",
		Namespace:            "default",
		WorkloadName:         "api-server",
		WorkloadType:         rightsizing.WorkloadTypeDeployment,
		ContainerName:        "main",
		CurrentCPURequest:    "100m",
		CurrentCPULimit:      "200m",
		CurrentMemoryRequest: "128Mi",
		CurrentMemoryLimit:   "256Mi",
		RecommendedCPURequest: "50m",
		RecommendedCPULimit:   "100m",
		RecommendedMemoryRequest: "64Mi",
		RecommendedMemoryLimit: "128Mi",
		Status:               rightsizing.StatusPending,
		CreatedAt:            now,
		UpdatedAt:            now,
		AnalysisPeriodStart:  now.Add(-24 * time.Hour),
		AnalysisPeriodEnd:    now,
	}

	t.Run("applies deployment recommendation", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-1").
			Return(pendingRec, nil)
		mockProvisioner.EXPECT().
			UpdateDeploymentResources(gomock.Any(), "default", "api-server", "main", "50m", "100m", "64Mi", "128Mi").
			Return(nil)
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		err := uc.ApplyRecommendation(context.Background(), "rec-1", "user-1")
		assert.NoError(t, err)
	})

	t.Run("applies statefulset recommendation", func(t *testing.T) {
		ssRec := &rightsizing.RightsizingRecommendation{
			ID:                   "rec-2",
			Namespace:            "default",
			WorkloadName:         "database",
			WorkloadType:         rightsizing.WorkloadTypeStatefulSet,
			ContainerName:        "mysql",
			CurrentCPURequest:    "500m",
			CurrentMemoryRequest: "512Mi",
			RecommendedCPURequest: "250m",
			RecommendedMemoryRequest: "256Mi",
			Status:               rightsizing.StatusPending,
			CreatedAt:            now,
			UpdatedAt:            now,
			AnalysisPeriodStart:  now.Add(-24 * time.Hour),
			AnalysisPeriodEnd:    now,
		}

		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-2").
			Return(ssRec, nil)
		mockProvisioner.EXPECT().
			UpdateStatefulSetResources(gomock.Any(), "default", "database", "mysql", "250m", "", "256Mi", "").
			Return(nil)
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		err := uc.ApplyRecommendation(context.Background(), "rec-2", "user-1")
		assert.NoError(t, err)
	})

	t.Run("rejects non-pending status", func(t *testing.T) {
		appliedRec := &rightsizing.RightsizingRecommendation{
			ID:     "rec-3",
			Status: rightsizing.StatusApplied,
		}
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-3").
			Return(appliedRec, nil)

		err := uc.ApplyRecommendation(context.Background(), "rec-3", "user-1")
		assert.ErrorContains(t, err, "not in pending status")
	})

	t.Run("handles kubernetes update failure", func(t *testing.T) {
		failRec := &rightsizing.RightsizingRecommendation{
			ID:                   "rec-fail",
			Namespace:            "default",
			WorkloadName:         "api-server",
			WorkloadType:         rightsizing.WorkloadTypeDeployment,
			ContainerName:        "main",
			CurrentCPURequest:    "100m",
			CurrentCPULimit:      "200m",
			CurrentMemoryRequest: "128Mi",
			CurrentMemoryLimit:   "256Mi",
			RecommendedCPURequest: "50m",
			RecommendedCPULimit:   "100m",
			RecommendedMemoryRequest: "64Mi",
			RecommendedMemoryLimit: "128Mi",
			Status:               rightsizing.StatusPending,
			CreatedAt:            now,
			UpdatedAt:            now,
			AnalysisPeriodStart:  now.Add(-24 * time.Hour),
			AnalysisPeriodEnd:    now,
		}
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-fail").
			Return(failRec, nil)
		mockProvisioner.EXPECT().
			UpdateDeploymentResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("kubernetes error"))
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		err := uc.ApplyRecommendation(context.Background(), "rec-fail", "user-1")
		assert.ErrorContains(t, err, "failed to apply recommendation")
	})

	t.Run("returns error for unsupported workload type", func(t *testing.T) {
		invalidRec := &rightsizing.RightsizingRecommendation{
			ID:           "rec-4",
			WorkloadType: "DaemonSet",
			Status:       rightsizing.StatusPending,
		}
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-4").
			Return(invalidRec, nil)

		err := uc.ApplyRecommendation(context.Background(), "rec-4", "user-1")
		assert.ErrorContains(t, err, "unsupported workload type")
	})
}

func TestRollbackRecommendation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRightsizingRepository(ctrl)
	mockProvisioner := mocks.NewMockProvisionerRepository(ctrl)
	uc := New(Dependencies{
		RightsizingRepo:  mockRepo,
		ProvisionerRepo:  mockProvisioner,
	})

	now := time.Now()
	appliedRec := &rightsizing.RightsizingRecommendation{
		ID:                   "rec-1",
		Namespace:            "default",
		WorkloadName:         "api-server",
		WorkloadType:         rightsizing.WorkloadTypeDeployment,
		ContainerName:        "main",
		CurrentCPURequest:    "50m",
		CurrentMemoryRequest: "64Mi",
		PreviousState:        `{"cpu_request":"100m","cpu_limit":"200m","memory_request":"128Mi","memory_limit":"256Mi"}`,
		Status:               rightsizing.StatusApplied,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	t.Run("rolls back to previous resources", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-1").
			Return(appliedRec, nil)
		mockProvisioner.EXPECT().
			UpdateDeploymentResources(gomock.Any(), "default", "api-server", "main", "100m", "200m", "128Mi", "256Mi").
			Return(nil)
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		err := uc.RollbackRecommendation(context.Background(), "rec-1", "user-1")
		assert.NoError(t, err)
	})

	t.Run("rejects non-applied status", func(t *testing.T) {
		pendingRec := &rightsizing.RightsizingRecommendation{
			ID:     "rec-2",
			Status: rightsizing.StatusPending,
		}
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-2").
			Return(pendingRec, nil)

		err := uc.RollbackRecommendation(context.Background(), "rec-2", "user-1")
		assert.ErrorContains(t, err, "not in applied status")
	})

	t.Run("returns error when no previous state", func(t *testing.T) {
		noPrevRec := &rightsizing.RightsizingRecommendation{
			ID:           "rec-3",
			Status:       rightsizing.StatusApplied,
			PreviousState: "",
		}
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-3").
			Return(noPrevRec, nil)

		err := uc.RollbackRecommendation(context.Background(), "rec-3", "user-1")
		assert.ErrorContains(t, err, "no previous state stored")
	})

	t.Run("handles kubernetes update failure", func(t *testing.T) {
		failRec := &rightsizing.RightsizingRecommendation{
			ID:                   "rec-fail-rollback",
			Namespace:            "default",
			WorkloadName:         "api-server",
			WorkloadType:         rightsizing.WorkloadTypeDeployment,
			ContainerName:        "main",
			CurrentCPURequest:    "50m",
			CurrentMemoryRequest: "64Mi",
			PreviousState:        `{"cpu_request":"100m","cpu_limit":"200m","memory_request":"128Mi","memory_limit":"256Mi"}`,
			Status:               rightsizing.StatusApplied,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-fail-rollback").
			Return(failRec, nil)
		mockProvisioner.EXPECT().
			UpdateDeploymentResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("kubernetes error"))

		err := uc.RollbackRecommendation(context.Background(), "rec-fail-rollback", "user-1")
		assert.ErrorContains(t, err, "failed to rollback")
	})
}

func TestDismissRecommendation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRightsizingRepository(ctrl)
	uc := New(Dependencies{
		RightsizingRepo: mockRepo,
	})

	t.Run("dismisses pending recommendation", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-1").
			Return(&rightsizing.RightsizingRecommendation{
				ID:     "rec-1",
				Status: rightsizing.StatusPending,
			}, nil)
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		err := uc.DismissRecommendation(context.Background(), "rec-1", "not needed")
		assert.NoError(t, err)
	})

	t.Run("rejects non-pending status", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "rec-2").
			Return(&rightsizing.RightsizingRecommendation{
				ID:     "rec-2",
				Status: rightsizing.StatusApplied,
			}, nil)

		err := uc.DismissRecommendation(context.Background(), "rec-2", "reason")
		assert.ErrorContains(t, err, "not in pending status")
	})
}

func TestGenerateRecommendations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRightsizingRepository(ctrl)
	mockProvisioner := mocks.NewMockProvisionerRepository(ctrl)
	mockMonitoring := mocks.NewMockMonitoringRepository(ctrl)
	uc := New(Dependencies{
		RightsizingRepo:  mockRepo,
		ProvisionerRepo:  mockProvisioner,
		MonitoringRepo:   mockMonitoring,
	})

	t.Run("generates recommendations for underutilized workloads", func(t *testing.T) {
		// Create a deployment with resources
		deploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "api-server",
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "main",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("1000m"),
										corev1.ResourceMemory: resource.MustParse("512Mi"),
									},
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("2000m"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
		}

		mockProvisioner.EXPECT().
			GetWorkloads("").
			Return([]*appsv1.Deployment{deploy}, nil)

		// No existing pending recommendation
		mockRepo.EXPECT().
			ExistsPendingForContainer(gomock.Any(), "default", "api-server", rightsizing.WorkloadTypeDeployment, "main").
			Return(false, nil)

		// CPU usage metrics (low utilization)
		mockMonitoring.EXPECT().
			Query(gomock.Any(), gomock.Any()).
			Return([]prometheusPkg.QueryResult{{Value: 0.1}}, nil).
			Times(4) // avg cpu, max cpu, avg mem, max mem

		// Create recommendation
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		err := uc.GenerateRecommendations(context.Background())
		assert.NoError(t, err)
	})

	t.Run("skips when pending recommendation exists", func(t *testing.T) {
		deploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "api-server",
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "main"},
						},
					},
				},
			},
		}

		mockProvisioner.EXPECT().
			GetWorkloads("").
			Return([]*appsv1.Deployment{deploy}, nil)

		mockRepo.EXPECT().
			ExistsPendingForContainer(gomock.Any(), "default", "api-server", rightsizing.WorkloadTypeDeployment, "main").
			Return(true, nil)

		err := uc.GenerateRecommendations(context.Background())
		assert.NoError(t, err)
	})

	t.Run("handles no workloads", func(t *testing.T) {
		mockProvisioner.EXPECT().
			GetWorkloads("").
			Return([]*appsv1.Deployment{}, nil)

		err := uc.GenerateRecommendations(context.Background())
		assert.NoError(t, err)
	})

	t.Run("continues on get workloads error", func(t *testing.T) {
		mockProvisioner.EXPECT().
			GetWorkloads("").
			Return(nil, errors.New("kubernetes error"))

		err := uc.GenerateRecommendations(context.Background())
		assert.ErrorContains(t, err, "failed to get workloads")
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("formatCPU", func(t *testing.T) {
		assert.Equal(t, "500m", formatCPU(0.5))
		assert.Equal(t, "1.50", formatCPU(1.5))
		assert.Equal(t, "", formatCPU(0))
	})

	t.Run("formatMemory", func(t *testing.T) {
		assert.Equal(t, "512Mi", formatMemory(512*1024*1024))
		assert.Equal(t, "1.00Gi", formatMemory(1024*1024*1024))
		assert.Equal(t, "", formatMemory(0))
	})

	t.Run("parseCPU", func(t *testing.T) {
		assert.Equal(t, 0.5, parseCPU("500m"))
		assert.Equal(t, 1.5, parseCPU("1.5"))
		assert.Equal(t, 0.0, parseCPU(""))
	})

	t.Run("parseMemory", func(t *testing.T) {
		assert.Equal(t, float64(512*1024*1024), parseMemory("512Mi"))
		assert.Equal(t, float64(1024*1024*1024), parseMemory("1Gi"))
		assert.Equal(t, 0.0, parseMemory(""))
	})
}

func TestCalculateConfidence(t *testing.T) {
	u := &usecase{}

	t.Run("high confidence with complete data", func(t *testing.T) {
		score := u.calculateConfidence(0.5, 0.8, 512*1024*1024, 768*1024*1024)
		assert.Equal(t, 100.0, score)
	})

	t.Run("reduces confidence for missing CPU avg", func(t *testing.T) {
		score := u.calculateConfidence(0, 0.8, 512*1024*1024, 768*1024*1024)
		assert.Equal(t, 75.0, score)
	})

	t.Run("reduces confidence for missing data", func(t *testing.T) {
		score := u.calculateConfidence(0, 0, 0, 0)
		assert.Equal(t, 20.0, score)
	})

	t.Run("reduces confidence for high variance", func(t *testing.T) {
		score := u.calculateConfidence(0.1, 0.5, 100*1024*1024, 500*1024*1024)
		assert.Equal(t, 80.0, score) // 100 - 10 - 10 = 80
	})
}
