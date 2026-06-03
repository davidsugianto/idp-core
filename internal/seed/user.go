package seed

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	userModel "github.com/davidsugianto/idp-core/internal/model/user"
)

// SeedPlatformAdmin ensures the platform admin user exists in the local database
// and has the platform_admin role assigned.
// This user must also exist in Keycloak with the platform-admins group membership.
func (s *Seeder) SeedPlatformAdmin(ctx context.Context) error {
	log.Println("Seeding platform admin user...")

	adminEmail := "admin@example.com"
	adminName := "Platform Admin"

	// Check if user already exists
	existing, err := s.userRepo.GetByEmail(ctx, adminEmail)
	if err != nil {
		return err
	}

	var userID string
	if existing != nil {
		log.Printf("Platform admin user already exists (id=%s), skipping creation", existing.ID)
		userID = existing.ID
	} else {
		u := &userModel.User{
			ID:       uuid.New().String(),
			Email:    adminEmail,
			Name:     adminName,
			Provider: "oidc",
			Status:   "active",
		}
		if err := s.userRepo.Create(ctx, u); err != nil {
			log.Printf("Failed to create platform admin user: %v", err)
			return err
		}
		userID = u.ID
		log.Printf("Created platform admin user (id=%s, email=%s)", userID, adminEmail)
	}

	// Look up the platform_admin role
	role, err := s.roleRepo.GetByName(ctx, "platform_admin")
	if err != nil {
		return err
	}
	if role == nil {
		log.Println("platform_admin role not found, skipping role assignment")
		return nil
	}

	// Check if role is already assigned
	userRoles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return err
	}
	for _, ur := range userRoles {
		if ur.RoleID == role.ID && ur.TeamID == "" {
			log.Printf("platform_admin role already assigned to user (id=%s)", userID)
			return nil
		}
	}

	// Assign platform_admin role
	userRole := &roleModel.UserRole{
		ID:        uuid.New().String(),
		UserID:    userID,
		RoleID:    role.ID,
		GrantedAt: time.Now(),
	}
	if err := s.roleRepo.AssignRole(ctx, userRole); err != nil {
		log.Printf("Failed to assign platform_admin role: %v", err)
		return err
	}

	log.Printf("Assigned platform_admin role to user (id=%s)", userID)
	return nil
}