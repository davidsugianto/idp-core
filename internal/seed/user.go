package seed

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	userModel "github.com/davidsugianto/idp-core/internal/model/user"
)

const (
	platformAdminName = "Platform Admin"
	developerName     = "Developer"
)

// SeedPlatformAdmin ensures the platform admin user exists in the local database
// and has the platform_admin role assigned.
// This user must also exist in Keycloak with the platform-admins group membership.
func (s *Seeder) SeedPlatformAdmin(ctx context.Context) error {
	log.Println("Seeding platform admin user...")

	adminUser, err := s.ensureOIDCUser(ctx, platformAdminEmail, platformAdminName)
	if err != nil {
		return err
	}
	userID := adminUser.ID

	role, err := s.roleRepo.GetByName(ctx, "platform_admin")
	if err != nil {
		return err
	}
	if role == nil {
		log.Println("platform_admin role not found, skipping role assignment")
		return nil
	}

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

func (s *Seeder) SeedDeveloperUser(ctx context.Context) error {
	log.Println("Seeding developer user...")

	_, err := s.ensureOIDCUser(ctx, developerEmail, developerName)
	return err
}

func (s *Seeder) ensureOIDCUser(ctx context.Context, email, name string) (*userModel.User, error) {
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		log.Printf("OIDC user already exists (id=%s, email=%s)", existing.ID, email)
		return existing, nil
	}

	u := &userModel.User{
		ID:       uuid.New().String(),
		Email:    email,
		Name:     name,
		Provider: "oidc",
		Status:   "active",
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		log.Printf("Failed to create OIDC user %s: %v", email, err)
		return nil, err
	}

	log.Printf("Created OIDC user (id=%s, email=%s)", u.ID, email)
	return u, nil
}
