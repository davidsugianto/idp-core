package seed

import (
	"context"
	"log"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
)

// DefaultRoles defines all default roles
var DefaultRoles = []struct {
	Name            string
	Description     string
	Scope           string
	PermissionNames []string
}{
	{
		Name:        "platform_admin",
		Description: "Platform administrator with full access to all resources",
		Scope:       roleModel.ScopePlatform,
		PermissionNames: []string{
			"environment:manage", "team:manage", "user:manage", "role:manage",
			"api_key:manage", "cost:manage", "budget:manage", "rightsizing:manage",
			"service:manage", "audit_log:manage",
		},
	},
	{
		Name:        "platform_viewer",
		Description: "Read-only access to platform resources",
		Scope:       roleModel.ScopePlatform,
		PermissionNames: []string{
			"environment:read", "team:read", "user:read", "role:read",
			"api_key:read", "cost:read", "budget:read", "rightsizing:read",
			"service:read", "audit_log:read",
		},
	},
	{
		Name:        "team_admin",
		Description: "Team administrator with full access to team resources",
		Scope:       roleModel.ScopeTeam,
		PermissionNames: []string{
			"environment:manage", "team:read", "team:update",
			"user:read", "role:read", "cost:read", "budget:manage",
			"rightsizing:read", "service:manage", "api_key:manage",
		},
	},
	{
		Name:        "team_developer",
		Description: "Developer access to team resources",
		Scope:       roleModel.ScopeTeam,
		PermissionNames: []string{
			"environment:create", "environment:read", "environment:update",
			"team:read", "user:read", "cost:read", "budget:read",
			"rightsizing:read", "service:create", "service:read", "service:update",
			"api_key:create", "api_key:read",
		},
	},
	{
		Name:        "team_viewer",
		Description: "Read-only access to team resources",
		Scope:       roleModel.ScopeTeam,
		PermissionNames: []string{
			"environment:read", "team:read", "user:read", "cost:read",
			"budget:read", "rightsizing:read", "service:read",
		},
	},
}

// SeedRoles seeds all default roles with their permissions
func (s *Seeder) SeedRoles(ctx context.Context) error {
	log.Println("Seeding roles...")

	for _, r := range DefaultRoles {
		existing, err := s.roleRepo.GetByName(ctx, r.Name)
		if err != nil {
			return err
		}

		if existing != nil {
			log.Printf("Role %s already exists, skipping", r.Name)
			continue
		}

		role := &roleModel.Role{
			Name:        r.Name,
			Description: r.Description,
			Scope:       r.Scope,
		}

		if err := s.roleRepo.Create(ctx, role); err != nil {
			log.Printf("Failed to create role %s: %v", r.Name, err)
			continue
		}

		// Get permission IDs
		permIDs, err := s.GetPermissionIDsByName(ctx, r.PermissionNames)
		if err != nil {
			log.Printf("Failed to get permissions for role %s: %v", r.Name, err)
			continue
		}

		// Assign permissions to role
		if len(permIDs) > 0 {
			if err := s.roleRepo.SetPermissions(ctx, role.ID, permIDs); err != nil {
				log.Printf("Failed to set permissions for role %s: %v", r.Name, err)
				continue
			}
		}

		log.Printf("Created role: %s with %d permissions", r.Name, len(permIDs))
	}

	return nil
}

// SeedAll runs all seeders
func (s *Seeder) SeedAll(ctx context.Context) error {
	if err := s.SeedPermissions(ctx); err != nil {
		return err
	}

	if err := s.SeedRoles(ctx); err != nil {
		return err
	}

	log.Println("Seeding completed successfully")
	return nil
}
