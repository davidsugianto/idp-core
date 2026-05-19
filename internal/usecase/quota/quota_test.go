package quota

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/model/resourcequota"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateQuota(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	mockProvisioner := mocks.NewMockProvisionerRepository(ctrl)
	uc := New(Dependencies{
		QuotaRepo:       mockRepo,
		ProvisionerRepo: mockProvisioner,
	})

	t.Run("creates quota with all fields", func(t *testing.T) {
		mockRepo.EXPECT().
			ExistsForNamespace(gomock.Any(), "default").
			Return(false, nil)
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)
		mockProvisioner.EXPECT().
			GetPods("default").
			Return([]*corev1.Pod{}, nil)
		mockRepo.EXPECT().
			UpdateUsage(gomock.Any(), "default", gomock.Any()).
			Return(nil)

		podLimit := 10
		resp, err := uc.CreateQuota(context.Background(), &resourcequota.CreateResourceQuotaRequest{
			Namespace:        "default",
			TeamID:           "team-1",
			CPURequestLimit:  "4",
			MemoryRequestLimit: "8Gi",
			PodCountLimit:    &podLimit,
			Enforce:          true,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "default", resp.Namespace)
		assert.Equal(t, "team-1", resp.TeamID)
		assert.Equal(t, "4", resp.CPURequestLimit)
		assert.Equal(t, resourcequota.StatusActive, resp.Status)
	})

	t.Run("rejects duplicate quota for namespace", func(t *testing.T) {
		mockRepo.EXPECT().
			ExistsForNamespace(gomock.Any(), "default").
			Return(true, nil)

		resp, err := uc.CreateQuota(context.Background(), &resourcequota.CreateResourceQuotaRequest{
			Namespace: "default",
			TeamID:    "team-1",
		})
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "quota already exists")
	})

	t.Run("propagates exists check error", func(t *testing.T) {
		mockRepo.EXPECT().
			ExistsForNamespace(gomock.Any(), "default").
			Return(false, errors.New("db error"))

		resp, err := uc.CreateQuota(context.Background(), &resourcequota.CreateResourceQuotaRequest{
			Namespace: "default",
			TeamID:    "team-1",
		})
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "db error")
	})

	t.Run("propagates create error", func(t *testing.T) {
		mockRepo.EXPECT().
			ExistsForNamespace(gomock.Any(), "default").
			Return(false, nil)
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("db error"))

		resp, err := uc.CreateQuota(context.Background(), &resourcequota.CreateResourceQuotaRequest{
			Namespace: "default",
			TeamID:    "team-1",
		})
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "db error")
	})
}

func TestGetQuota(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	uc := New(Dependencies{QuotaRepo: mockRepo})

	t.Run("returns quota by id", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "quota-1").
			Return(&resourcequota.ResourceQuota{
				ID:        "quota-1",
				Namespace: "default",
				TeamID:    "team-1",
				Status:    resourcequota.StatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil)

		resp, err := uc.GetQuota(context.Background(), "quota-1")
		assert.NoError(t, err)
		assert.Equal(t, "quota-1", resp.ID)
		assert.Equal(t, "default", resp.Namespace)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, errors.New("not found"))

		resp, err := uc.GetQuota(context.Background(), "nonexistent")
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "not found")
	})
}

func TestGetQuotaByNamespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	uc := New(Dependencies{QuotaRepo: mockRepo})

	t.Run("returns quota by namespace", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByNamespace(gomock.Any(), "default").
			Return(&resourcequota.ResourceQuota{
				ID:        "quota-1",
				Namespace: "default",
				Status:    resourcequota.StatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil)

		resp, err := uc.GetQuotaByNamespace(context.Background(), "default")
		assert.NoError(t, err)
		assert.Equal(t, "default", resp.Namespace)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByNamespace(gomock.Any(), "nonexistent").
			Return(nil, errors.New("not found"))

		resp, err := uc.GetQuotaByNamespace(context.Background(), "nonexistent")
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "not found")
	})
}

func TestListQuotas(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	uc := New(Dependencies{QuotaRepo: mockRepo})

	t.Run("returns quotas", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]resourcequota.ResourceQuota{
				{ID: "quota-1", Namespace: "ns-1", Status: resourcequota.StatusActive, CreatedAt: now, UpdatedAt: now},
				{ID: "quota-2", Namespace: "ns-2", Status: resourcequota.StatusActive, CreatedAt: now, UpdatedAt: now},
			}, int64(2), nil)

		resp, err := uc.ListQuotas(context.Background(), &resourcequota.ListResourceQuotasRequest{})
		assert.NoError(t, err)
		assert.Len(t, resp.Quotas, 2)
		assert.Equal(t, int64(2), resp.Total)
	})

	t.Run("returns empty list", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]resourcequota.ResourceQuota{}, int64(0), nil)

		resp, err := uc.ListQuotas(context.Background(), &resourcequota.ListResourceQuotasRequest{})
		assert.NoError(t, err)
		assert.Empty(t, resp.Quotas)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return(nil, int64(0), errors.New("db error"))

		resp, err := uc.ListQuotas(context.Background(), &resourcequota.ListResourceQuotasRequest{})
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "db error")
	})
}

func TestUpdateQuota(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	uc := New(Dependencies{QuotaRepo: mockRepo})

	existing := &resourcequota.ResourceQuota{
		ID:             "quota-1",
		Namespace:      "default",
		CPURequestLimit: "4",
		Status:         resourcequota.StatusActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	t.Run("updates cpu limit", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "quota-1").
			Return(existing, nil)
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		newCPU := "8"
		resp, err := uc.UpdateQuota(context.Background(), "quota-1", &resourcequota.UpdateResourceQuotaRequest{
			CPURequestLimit: &newCPU,
		})
		assert.NoError(t, err)
		assert.Equal(t, "8", resp.CPURequestLimit)
	})

	t.Run("updates enforce flag", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "quota-1").
			Return(existing, nil)
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		enforce := false
		resp, err := uc.UpdateQuota(context.Background(), "quota-1", &resourcequota.UpdateResourceQuotaRequest{
			Enforce: &enforce,
		})
		assert.NoError(t, err)
		assert.False(t, resp.Enforce)
	})

	t.Run("returns not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, errors.New("not found"))

		newCPU := "8"
		resp, err := uc.UpdateQuota(context.Background(), "nonexistent", &resourcequota.UpdateResourceQuotaRequest{
			CPURequestLimit: &newCPU,
		})
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "not found")
	})
}

func TestDeleteQuota(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	uc := New(Dependencies{QuotaRepo: mockRepo})

	t.Run("deletes quota", func(t *testing.T) {
		mockRepo.EXPECT().
			Delete(gomock.Any(), "quota-1").
			Return(nil)

		err := uc.DeleteQuota(context.Background(), "quota-1")
		assert.NoError(t, err)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		mockRepo.EXPECT().
			Delete(gomock.Any(), "quota-1").
			Return(errors.New("db error"))

		err := uc.DeleteQuota(context.Background(), "quota-1")
		assert.ErrorContains(t, err, "db error")
	})
}

func TestCheckQuota(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	uc := New(Dependencies{QuotaRepo: mockRepo})

	t.Run("allows when no quota defined", func(t *testing.T) {
		mockRepo.EXPECT().
			GetActiveByNamespace(gomock.Any(), "default").
			Return(nil, errors.New("not found"))

		resp, err := uc.CheckQuota(context.Background(), &resourcequota.QuotaCheckRequest{
			Namespace:  "default",
			CPURequest: "100m",
		})
		assert.NoError(t, err)
		assert.True(t, resp.Allowed)
	})

	t.Run("allows when quota not enforced", func(t *testing.T) {
		mockRepo.EXPECT().
			GetActiveByNamespace(gomock.Any(), "default").
			Return(&resourcequota.ResourceQuota{
				Namespace:       "default",
				CPURequestLimit: "4",
				Enforce:         false,
			}, nil)

		resp, err := uc.CheckQuota(context.Background(), &resourcequota.QuotaCheckRequest{
			Namespace:  "default",
			CPURequest: "100m",
		})
		assert.NoError(t, err)
		assert.True(t, resp.Allowed)
	})

	t.Run("allows when within limit", func(t *testing.T) {
		mockRepo.EXPECT().
			GetActiveByNamespace(gomock.Any(), "default").
			Return(&resourcequota.ResourceQuota{
				Namespace:          "default",
				CPURequestLimit:    "4",
				CurrentCPURequest:  "1",
				Enforce:            true,
			}, nil)

		resp, err := uc.CheckQuota(context.Background(), &resourcequota.QuotaCheckRequest{
			Namespace:  "default",
			CPURequest: "500m",
		})
		assert.NoError(t, err)
		assert.True(t, resp.Allowed)
	})

	t.Run("rejects when exceeds limit", func(t *testing.T) {
		mockRepo.EXPECT().
			GetActiveByNamespace(gomock.Any(), "default").
			Return(&resourcequota.ResourceQuota{
				Namespace:          "default",
				CPURequestLimit:    "2",
				CurrentCPURequest:  "1.5",
				Enforce:            true,
			}, nil)

		resp, err := uc.CheckQuota(context.Background(), &resourcequota.QuotaCheckRequest{
			Namespace:  "default",
			CPURequest: "1",
		})
		assert.NoError(t, err)
		assert.False(t, resp.Allowed)
		assert.Len(t, resp.Reasons, 1)
		assert.Equal(t, "cpu_request", resp.Reasons[0].ResourceType)
	})

	t.Run("checks pod count", func(t *testing.T) {
		podLimit := 10
		currentPods := 9
		mockRepo.EXPECT().
			GetActiveByNamespace(gomock.Any(), "default").
			Return(&resourcequota.ResourceQuota{
				Namespace:       "default",
				PodCountLimit:   &podLimit,
				CurrentPodCount: &currentPods,
				Enforce:         true,
			}, nil)

		resp, err := uc.CheckQuota(context.Background(), &resourcequota.QuotaCheckRequest{
			Namespace: "default",
			PodDelta:  2,
		})
		assert.NoError(t, err)
		assert.False(t, resp.Allowed)
		assert.Len(t, resp.Reasons, 1)
		assert.Equal(t, "pods", resp.Reasons[0].ResourceType)
	})
}

func TestIsQuotaExceeded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	uc := New(Dependencies{QuotaRepo: mockRepo})

	t.Run("returns false when no quota", func(t *testing.T) {
		mockRepo.EXPECT().
			GetActiveByNamespace(gomock.Any(), "default").
			Return(nil, errors.New("not found"))

		exceeded, reasons, err := uc.IsQuotaExceeded(context.Background(), "default")
		assert.NoError(t, err)
		assert.False(t, exceeded)
		assert.Nil(t, reasons)
	})

	t.Run("returns false when within limits", func(t *testing.T) {
		mockRepo.EXPECT().
			GetActiveByNamespace(gomock.Any(), "default").
			Return(&resourcequota.ResourceQuota{
				Namespace:          "default",
				CPURequestLimit:    "4",
				CurrentCPURequest:  "2",
				Enforce:            true,
			}, nil)

		exceeded, reasons, err := uc.IsQuotaExceeded(context.Background(), "default")
		assert.NoError(t, err)
		assert.False(t, exceeded)
		assert.Empty(t, reasons)
	})

	t.Run("returns true when exceeds limit", func(t *testing.T) {
		podLimit := 10
		currentPods := 15
		mockRepo.EXPECT().
			GetActiveByNamespace(gomock.Any(), "default").
			Return(&resourcequota.ResourceQuota{
				Namespace:       "default",
				PodCountLimit:   &podLimit,
				CurrentPodCount: &currentPods,
				Enforce:         true,
			}, nil)

		exceeded, reasons, err := uc.IsQuotaExceeded(context.Background(), "default")
		assert.NoError(t, err)
		assert.True(t, exceeded)
		assert.Len(t, reasons, 1)
	})
}

func TestGetUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	mockProvisioner := mocks.NewMockProvisionerRepository(ctrl)
	uc := New(Dependencies{
		QuotaRepo:       mockRepo,
		ProvisionerRepo: mockProvisioner,
	})

	t.Run("calculates usage from pods", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "test-pod",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "main",
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("256Mi"),
							},
						},
					},
				},
			},
		}

		mockProvisioner.EXPECT().
			GetPods("default").
			Return([]*corev1.Pod{pod}, nil)

		resp, err := uc.GetUsage(context.Background(), "default")
		assert.NoError(t, err)
		assert.Equal(t, "default", resp.Namespace)
		assert.Equal(t, 1, resp.PodCount)
	})

	t.Run("returns empty usage for no pods", func(t *testing.T) {
		mockProvisioner.EXPECT().
			GetPods("empty-ns").
			Return([]*corev1.Pod{}, nil)

		resp, err := uc.GetUsage(context.Background(), "empty-ns")
		assert.NoError(t, err)
		assert.Equal(t, "empty-ns", resp.Namespace)
		assert.Equal(t, 0, resp.PodCount)
	})

	t.Run("propagates provisioner error", func(t *testing.T) {
		mockProvisioner.EXPECT().
			GetPods("default").
			Return(nil, errors.New("kubernetes error"))

		resp, err := uc.GetUsage(context.Background(), "default")
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "kubernetes error")
	})
}

func TestRefreshUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	mockProvisioner := mocks.NewMockProvisionerRepository(ctrl)
	uc := New(Dependencies{
		QuotaRepo:       mockRepo,
		ProvisionerRepo: mockProvisioner,
	})

	t.Run("refreshes and updates usage", func(t *testing.T) {
		mockProvisioner.EXPECT().
			GetPods("default").
			Return([]*corev1.Pod{}, nil)
		mockRepo.EXPECT().
			UpdateUsage(gomock.Any(), "default", gomock.Any()).
			Return(nil)

		err := uc.RefreshUsage(context.Background(), "default")
		assert.NoError(t, err)
	})
}

func TestRefreshAllUsage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockQuotaRepository(ctrl)
	mockProvisioner := mocks.NewMockProvisionerRepository(ctrl)
	uc := New(Dependencies{
		QuotaRepo:       mockRepo,
		ProvisionerRepo: mockProvisioner,
	})

	t.Run("refreshes all quotas", func(t *testing.T) {
		quota1 := resourcequota.ResourceQuota{Namespace: "ns-1"}
		quota2 := resourcequota.ResourceQuota{Namespace: "ns-2"}

		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]resourcequota.ResourceQuota{quota1, quota2}, int64(2), nil)

		mockProvisioner.EXPECT().
			GetPods("ns-1").
			Return([]*corev1.Pod{}, nil)
		mockRepo.EXPECT().
			UpdateUsage(gomock.Any(), "ns-1", gomock.Any()).
			Return(nil)

		mockProvisioner.EXPECT().
			GetPods("ns-2").
			Return([]*corev1.Pod{}, nil)
		mockRepo.EXPECT().
			UpdateUsage(gomock.Any(), "ns-2", gomock.Any()).
			Return(nil)

		err := uc.RefreshAllUsage(context.Background())
		assert.NoError(t, err)
	})

	t.Run("continues on list error", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return(nil, int64(0), errors.New("db error"))

		err := uc.RefreshAllUsage(context.Background())
		assert.ErrorContains(t, err, "db error")
	})
}
