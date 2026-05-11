package seed

import (
	"context"
	"log"

	permissionModel "github.com/davidsugianto/idp-core/internal/model/permission"
	permissionRepo "github.com/davidsugianto/idp-core/internal/repository/permission"
	roleRepo "github.com/davidsugianto/idp-core/internal/repository/role"
)

// Seeder handles database seeding
type Seeder struct {
	roleRepo       roleRepo.Repository
	permissionRepo permissionRepo.Repository
}

// NewSeeder creates a new seeder
func NewSeeder(roleRepo roleRepo.Repository, permissionRepo permissionRepo.Repository) *Seeder {
	return &Seeder{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// DefaultPermissions defines all default permissions
var DefaultPermissions = []struct {
	Name        string
	Description string
	Resource    string
	Action      string
}{
	// Environment permissions
	{"environment:create", "Create environments", permissionModel.ResourceEnvironment, permissionModel.ActionCreate},
	{"environment:read", "View environments", permissionModel.ResourceEnvironment, permissionModel.ActionRead},
	{"environment:update", "Update environments", permissionModel.ResourceEnvironment, permissionModel.ActionUpdate},
	{"environment:delete", "Delete environments", permissionModel.ResourceEnvironment, permissionModel.ActionDelete},
	{"environment:manage", "Full environment access", permissionModel.ResourceEnvironment, permissionModel.ActionManage},

	// Team permissions
	{"team:create", "Create teams", permissionModel.ResourceTeam, permissionModel.ActionCreate},
	{"team:read", "View teams", permissionModel.ResourceTeam, permissionModel.ActionRead},
	{"team:update", "Update teams", permissionModel.ResourceTeam, permissionModel.ActionUpdate},
	{"team:delete", "Delete teams", permissionModel.ResourceTeam, permissionModel.ActionDelete},
	{"team:manage", "Full team access", permissionModel.ResourceTeam, permissionModel.ActionManage},

	// User permissions
	{"user:create", "Create users", permissionModel.ResourceUser, permissionModel.ActionCreate},
	{"user:read", "View users", permissionModel.ResourceUser, permissionModel.ActionRead},
	{"user:update", "Update users", permissionModel.ResourceUser, permissionModel.ActionUpdate},
	{"user:delete", "Delete users", permissionModel.ResourceUser, permissionModel.ActionDelete},
	{"user:manage", "Full user access", permissionModel.ResourceUser, permissionModel.ActionManage},

	// Role permissions
	{"role:create", "Create roles", permissionModel.ResourceRole, permissionModel.ActionCreate},
	{"role:read", "View roles", permissionModel.ResourceRole, permissionModel.ActionRead},
	{"role:update", "Update roles", permissionModel.ResourceRole, permissionModel.ActionUpdate},
	{"role:delete", "Delete roles", permissionModel.ResourceRole, permissionModel.ActionDelete},
	{"role:manage", "Full role access", permissionModel.ResourceRole, permissionModel.ActionManage},

	// API Key permissions
	{"api_key:create", "Create API keys", permissionModel.ResourceAPIKey, permissionModel.ActionCreate},
	{"api_key:read", "View API keys", permissionModel.ResourceAPIKey, permissionModel.ActionRead},
	{"api_key:delete", "Delete API keys", permissionModel.ResourceAPIKey, permissionModel.ActionDelete},
	{"api_key:manage", "Full API key access", permissionModel.ResourceAPIKey, permissionModel.ActionManage},

	// Cost permissions
	{"cost:read", "View cost data", permissionModel.ResourceCost, permissionModel.ActionRead},
	{"cost:manage", "Full cost access", permissionModel.ResourceCost, permissionModel.ActionManage},

	// Budget permissions
	{"budget:create", "Create budgets", permissionModel.ResourceBudget, permissionModel.ActionCreate},
	{"budget:read", "View budgets", permissionModel.ResourceBudget, permissionModel.ActionRead},
	{"budget:update", "Update budgets", permissionModel.ResourceBudget, permissionModel.ActionUpdate},
	{"budget:delete", "Delete budgets", permissionModel.ResourceBudget, permissionModel.ActionDelete},
	{"budget:manage", "Full budget access", permissionModel.ResourceBudget, permissionModel.ActionManage},

	// Rightsizing permissions
	{"rightsizing:read", "View rightsizing recommendations", permissionModel.ResourceRightsizing, permissionModel.ActionRead},
	{"rightsizing:update", "Apply rightsizing recommendations", permissionModel.ResourceRightsizing, permissionModel.ActionUpdate},
	{"rightsizing:manage", "Full rightsizing access", permissionModel.ResourceRightsizing, permissionModel.ActionManage},

	// Service permissions
	{"service:create", "Create services", permissionModel.ResourceService, permissionModel.ActionCreate},
	{"service:read", "View services", permissionModel.ResourceService, permissionModel.ActionRead},
	{"service:update", "Update services", permissionModel.ResourceService, permissionModel.ActionUpdate},
	{"service:delete", "Delete services", permissionModel.ResourceService, permissionModel.ActionDelete},
	{"service:manage", "Full service access", permissionModel.ResourceService, permissionModel.ActionManage},

	// Audit log permissions
	{"audit_log:read", "View audit logs", permissionModel.ResourceAuditLog, permissionModel.ActionRead},
	{"audit_log:manage", "Full audit log access", permissionModel.ResourceAuditLog, permissionModel.ActionManage},
}

// SeedPermissions seeds all default permissions
func (s *Seeder) SeedPermissions(ctx context.Context) error {
	log.Println("Seeding permissions...")

	for _, p := range DefaultPermissions {
		existing, err := s.permissionRepo.GetByName(ctx, p.Name)
		if err != nil {
			return err
		}

		if existing != nil {
			continue // Already exists
		}

		permission := &permissionModel.Permission{
			Name:        p.Name,
			Description: p.Description,
			Resource:    p.Resource,
			Action:      p.Action,
		}

		if err := s.permissionRepo.Create(ctx, permission); err != nil {
			log.Printf("Failed to create permission %s: %v", p.Name, err)
			continue
		}

		log.Printf("Created permission: %s", p.Name)
	}

	return nil
}

// GetPermissionIDsByName returns permission IDs for given names
func (s *Seeder) GetPermissionIDsByName(ctx context.Context, names []string) ([]string, error) {
	ids := make([]string, 0, len(names))
	for _, name := range names {
		perm, err := s.permissionRepo.GetByName(ctx, name)
		if err != nil {
			return nil, err
		}
		if perm != nil {
			ids = append(ids, perm.ID)
		}
	}
	return ids, nil
}
