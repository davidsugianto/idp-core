package service

import (
	"context"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/mocks"
	depModel "github.com/davidsugianto/idp-core/internal/model/service_dependency"
	svcEnvModel "github.com/davidsugianto/idp-core/internal/model/service_environment"
	serviceModel "github.com/davidsugianto/idp-core/internal/model/service"
	versionModel "github.com/davidsugianto/idp-core/internal/model/service_version"
	environmentModel "github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockServiceRepository(ctrl)
	uc := New(Dependencies{ServiceRepo: mockRepo})

	t.Run("creates service with required fields", func(t *testing.T) {
		mockRepo.EXPECT().
			ExistsByName(gomock.Any(), "api-gateway", "team-1").
			Return(false, nil)
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := uc.Register(context.Background(), &serviceModel.CreateServiceRequest{
			Name:   "api-gateway",
			TeamID: "team-1",
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "api-gateway", resp.Name)
		assert.Equal(t, serviceModel.VisibilityTeam, resp.Visibility)
		assert.Equal(t, serviceModel.StatusActive, resp.Status)
	})

	t.Run("rejects empty name", func(t *testing.T) {
		resp, err := uc.Register(context.Background(), &serviceModel.CreateServiceRequest{
			TeamID: "team-1",
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrServiceNameRequired)
	})

	t.Run("rejects duplicate service", func(t *testing.T) {
		mockRepo.EXPECT().
			ExistsByName(gomock.Any(), "existing-service", "team-1").
			Return(true, nil)

		resp, err := uc.Register(context.Background(), &serviceModel.CreateServiceRequest{
			Name:   "existing-service",
			TeamID: "team-1",
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrServiceAlreadyExists)
	})

	t.Run("rejects invalid visibility", func(t *testing.T) {
		mockRepo.EXPECT().
			ExistsByName(gomock.Any(), "svc-1", "team-1").
			Return(false, nil)

		resp, err := uc.Register(context.Background(), &serviceModel.CreateServiceRequest{
			Name:       "svc-1",
			TeamID:     "team-1",
			Visibility: "invalid",
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidVisibility)
	})
}

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockServiceRepository(ctrl)
	uc := New(Dependencies{ServiceRepo: mockRepo})

	t.Run("returns service by id", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-1").
			Return(&serviceModel.Service{
				ID:        "svc-1",
				Name:      "api-gateway",
				TeamID:    "team-1",
				Status:    serviceModel.StatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil)

		resp, err := uc.Get(context.Background(), "svc-1")
		assert.NoError(t, err)
		assert.Equal(t, "svc-1", resp.ID)
		assert.Equal(t, "api-gateway", resp.Name)
	})

	t.Run("returns not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, gorm.ErrRecordNotFound)

		resp, err := uc.Get(context.Background(), "nonexistent")
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrServiceNotFound)
	})
}

func TestAddDependency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockServiceRepository(ctrl)
	uc := New(Dependencies{ServiceRepo: mockRepo})

	t.Run("creates dependency", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-1").
			Return(&serviceModel.Service{ID: "svc-1", Name: "api-gateway", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-2").
			Return(&serviceModel.Service{ID: "svc-2", Name: "database", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			ExistsDependency(gomock.Any(), "svc-1", "svc-2").
			Return(false, nil)
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-2", gomock.Any()).
			Return([]depModel.ServiceDependency{}, int64(0), nil)
		mockRepo.EXPECT().
			CreateDependency(gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := uc.AddDependency(context.Background(), "svc-1", &depModel.CreateDependencyRequest{
			DependsOnServiceID: "svc-2",
			DependencyType:     depModel.TypeRuntime,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "svc-1", resp.ServiceID)
		assert.Equal(t, "svc-2", resp.DependsOnServiceID)
	})

	t.Run("rejects self-dependency", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-1").
			Return(&serviceModel.Service{ID: "svc-1", Name: "svc", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-1").
			Return(&serviceModel.Service{ID: "svc-1", Name: "svc", CreatedAt: now, UpdatedAt: now}, nil)

		resp, err := uc.AddDependency(context.Background(), "svc-1", &depModel.CreateDependencyRequest{
			DependsOnServiceID: "svc-1",
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrSelfDependency)
	})

	t.Run("rejects duplicate dependency", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-1").
			Return(&serviceModel.Service{ID: "svc-1", Name: "svc", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-2").
			Return(&serviceModel.Service{ID: "svc-2", Name: "db", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			ExistsDependency(gomock.Any(), "svc-1", "svc-2").
			Return(true, nil)

		resp, err := uc.AddDependency(context.Background(), "svc-1", &depModel.CreateDependencyRequest{
			DependsOnServiceID: "svc-2",
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrDependencyAlreadyExists)
	})

	t.Run("detects circular dependency", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-1").
			Return(&serviceModel.Service{ID: "svc-1", Name: "svc", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-2").
			Return(&serviceModel.Service{ID: "svc-2", Name: "db", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			ExistsDependency(gomock.Any(), "svc-1", "svc-2").
			Return(false, nil)
		// svc-2 -> svc-1 exists, so adding svc-1 -> svc-2 creates cycle
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-2", gomock.Any()).
			Return([]depModel.ServiceDependency{
				{ServiceID: "svc-2", DependsOnServiceID: "svc-1"},
			}, int64(1), nil)

		resp, err := uc.AddDependency(context.Background(), "svc-1", &depModel.CreateDependencyRequest{
			DependsOnServiceID: "svc-2",
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrCircularDependency)
	})

	t.Run("rejects invalid dependency type", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-1").
			Return(&serviceModel.Service{ID: "svc-1", Name: "svc", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-2").
			Return(&serviceModel.Service{ID: "svc-2", Name: "db", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			ExistsDependency(gomock.Any(), "svc-1", "svc-2").
			Return(false, nil)
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-2", gomock.Any()).
			Return([]depModel.ServiceDependency{}, int64(0), nil)

		resp, err := uc.AddDependency(context.Background(), "svc-1", &depModel.CreateDependencyRequest{
			DependsOnServiceID: "svc-2",
			DependencyType:     "invalid",
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidDependencyType)
	})
}

func TestGetDependencyGraph(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockServiceRepository(ctrl)
	uc := New(Dependencies{ServiceRepo: mockRepo})

	t.Run("returns dependency graph", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-1").
			Return(&serviceModel.Service{ID: "svc-1", Name: "api-gateway", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-1", gomock.Any()).
			Return([]depModel.ServiceDependency{
				{ServiceID: "svc-1", DependsOnServiceID: "svc-2", DependencyType: depModel.TypeRuntime},
			}, int64(1), nil)
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "svc-2").
			Return(&serviceModel.Service{ID: "svc-2", Name: "database", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			ListDependentsByService(gomock.Any(), "svc-1", gomock.Any()).
			Return([]depModel.ServiceDependency{}, int64(0), nil)
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-2", gomock.Any()).
			Return([]depModel.ServiceDependency{}, int64(0), nil)
		mockRepo.EXPECT().
			ListDependentsByService(gomock.Any(), "svc-2", gomock.Any()).
			Return([]depModel.ServiceDependency{}, int64(0), nil)

		graph, err := uc.GetDependencyGraph(context.Background(), "svc-1")
		assert.NoError(t, err)
		assert.Equal(t, "svc-1", graph.ServiceID)
		assert.Equal(t, "api-gateway", graph.ServiceName)
		assert.Len(t, graph.Nodes, 2)
		assert.Len(t, graph.Edges, 1)
	})

	t.Run("returns not found for nonexistent service", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, gorm.ErrRecordNotFound)

		graph, err := uc.GetDependencyGraph(context.Background(), "nonexistent")
		assert.Nil(t, graph)
		assert.ErrorIs(t, err, ErrServiceNotFound)
	})
}

func TestDeployToEnvironment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvcRepo := mocks.NewMockServiceRepository(ctrl)
	mockEnvRepo := mocks.NewMockEnvironmentRepository(ctrl)
	uc := New(Dependencies{
		ServiceRepo:   mockSvcRepo,
		EnvironmentRepo: mockEnvRepo,
	})

	now := time.Now()

	t.Run("deploys version to environment", func(t *testing.T) {
		mockSvcRepo.EXPECT().
			GetVersionByID(gomock.Any(), "ver-1").
			Return(&versionModel.ServiceVersion{ID: "ver-1", ServiceID: "svc-1", Version: "1.0.0", CreatedAt: now}, nil)
		mockEnvRepo.EXPECT().
			GetByID(gomock.Any(), "env-1").
			Return(&environmentModel.Environment{ID: "env-1", Name: "production", Namespace: "prod", Status: "ready", CreatedAt: now, UpdatedAt: now}, nil)
		mockSvcRepo.EXPECT().
			GetActiveDeployment(gomock.Any(), "ver-1", "env-1").
			Return(nil, gorm.ErrRecordNotFound)
		mockSvcRepo.EXPECT().
			CreateServiceEnvironment(gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := uc.DeployToEnvironment(context.Background(), "ver-1", &svcEnvModel.DeployRequest{
			EnvironmentID: "env-1",
		}, "user-1")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "ver-1", resp.ServiceVersionID)
		assert.Equal(t, "env-1", resp.EnvironmentID)
		assert.Equal(t, "user-1", resp.DeployedBy)
		assert.Equal(t, svcEnvModel.StatusDeployed, resp.Status)
	})

	t.Run("returns error for nonexistent version", func(t *testing.T) {
		mockSvcRepo.EXPECT().
			GetVersionByID(gomock.Any(), "nonexistent").
			Return(nil, gorm.ErrRecordNotFound)

		resp, err := uc.DeployToEnvironment(context.Background(), "nonexistent", &svcEnvModel.DeployRequest{
			EnvironmentID: "env-1",
		}, "user-1")
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrVersionNotFound)
	})
}

func TestCheckCircularDependency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockServiceRepository(ctrl)
	u := &usecase{serviceRepo: mockRepo}

	t.Run("no cycle detected", func(t *testing.T) {
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-2", gomock.Any()).
			Return([]depModel.ServiceDependency{
				{ServiceID: "svc-2", DependsOnServiceID: "svc-3"},
			}, int64(1), nil)
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-3", gomock.Any()).
			Return([]depModel.ServiceDependency{}, int64(0), nil)

		err := u.checkCircularDependency(context.Background(), "svc-1", "svc-2")
		assert.NoError(t, err)
	})

	t.Run("cycle detected", func(t *testing.T) {
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-2", gomock.Any()).
			Return([]depModel.ServiceDependency{
				{ServiceID: "svc-2", DependsOnServiceID: "svc-1"},
			}, int64(1), nil)

		err := u.checkCircularDependency(context.Background(), "svc-1", "svc-2")
		assert.ErrorIs(t, err, ErrCircularDependency)
	})

	t.Run("indirect cycle detected", func(t *testing.T) {
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-2", gomock.Any()).
			Return([]depModel.ServiceDependency{
				{ServiceID: "svc-2", DependsOnServiceID: "svc-3"},
			}, int64(1), nil)
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-3", gomock.Any()).
			Return([]depModel.ServiceDependency{
				{ServiceID: "svc-3", DependsOnServiceID: "svc-1"},
			}, int64(1), nil)

		err := u.checkCircularDependency(context.Background(), "svc-1", "svc-2")
		assert.ErrorIs(t, err, ErrCircularDependency)
	})
}

func TestListDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockServiceRepository(ctrl)
	uc := New(Dependencies{ServiceRepo: mockRepo})

	t.Run("returns dependencies", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			ListDependenciesByService(gomock.Any(), "svc-1", gomock.Any()).
			Return([]depModel.ServiceDependency{
				{ID: "dep-1", ServiceID: "svc-1", DependsOnServiceID: "svc-2", DependencyType: depModel.TypeRuntime, CreatedAt: now, UpdatedAt: now},
			}, int64(1), nil)

		resp, err := uc.ListDependencies(context.Background(), "svc-1", &depModel.ListDependenciesRequest{})
		assert.NoError(t, err)
		assert.Len(t, resp.Dependencies, 1)
		assert.Equal(t, int64(1), resp.Total)
	})
}

func TestListDependents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockServiceRepository(ctrl)
	uc := New(Dependencies{ServiceRepo: mockRepo})

	t.Run("returns dependents", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			ListDependentsByService(gomock.Any(), "svc-1", gomock.Any()).
			Return([]depModel.ServiceDependency{
				{ID: "dep-1", ServiceID: "svc-2", DependsOnServiceID: "svc-1", DependencyType: depModel.TypeRuntime, CreatedAt: now, UpdatedAt: now},
			}, int64(1), nil)

		resp, err := uc.ListDependents(context.Background(), "svc-1", &depModel.ListDependenciesRequest{})
		assert.NoError(t, err)
		assert.Len(t, resp.Dependencies, 1)
		assert.Equal(t, int64(1), resp.Total)
	})
}

func TestRemoveDependency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockServiceRepository(ctrl)
	uc := New(Dependencies{ServiceRepo: mockRepo})

	t.Run("deletes dependency", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetDependencyByID(gomock.Any(), "dep-1").
			Return(&depModel.ServiceDependency{ID: "dep-1", CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			DeleteDependency(gomock.Any(), "dep-1").
			Return(nil)

		err := uc.RemoveDependency(context.Background(), "dep-1")
		assert.NoError(t, err)
	})

	t.Run("returns not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetDependencyByID(gomock.Any(), "nonexistent").
			Return(nil, gorm.ErrRecordNotFound)

		err := uc.RemoveDependency(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, ErrDependencyNotFound)
	})
}

func TestUpdateDependency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockServiceRepository(ctrl)
	uc := New(Dependencies{ServiceRepo: mockRepo})

	t.Run("updates dependency type", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetDependencyByID(gomock.Any(), "dep-1").
			Return(&depModel.ServiceDependency{ID: "dep-1", DependencyType: depModel.TypeRuntime, CreatedAt: now, UpdatedAt: now}, nil)
		mockRepo.EXPECT().
			UpdateDependency(gomock.Any(), gomock.Any()).
			Return(nil)

		newType := depModel.TypeBuild
		resp, err := uc.UpdateDependency(context.Background(), "dep-1", &depModel.UpdateDependencyRequest{
			DependencyType: &newType,
		})
		assert.NoError(t, err)
		assert.Equal(t, depModel.TypeBuild, resp.DependencyType)
	})

	t.Run("rejects invalid dependency type", func(t *testing.T) {
		now := time.Now()
		mockRepo.EXPECT().
			GetDependencyByID(gomock.Any(), "dep-1").
			Return(&depModel.ServiceDependency{ID: "dep-1", CreatedAt: now, UpdatedAt: now}, nil)

		invalidType := "invalid"
		resp, err := uc.UpdateDependency(context.Background(), "dep-1", &depModel.UpdateDependencyRequest{
			DependencyType: &invalidType,
		})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidDependencyType)
	})
}
