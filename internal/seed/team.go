package seed

import (
	"context"
	"log"
	"time"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	teamModel "github.com/davidsugianto/idp-core/internal/model/team"
	"github.com/google/uuid"
)

const (
	defaultTeamName = "Engineering"
	defaultTeamSlug = "engineering"

	platformAdminEmail = "admin@example.com"
	developerEmail     = "developer@example.com"
)

var defaultTeamMembers = []struct {
	email    string
	roleName string
	teamRole string
}{
	{email: platformAdminEmail, roleName: "team_admin", teamRole: teamModel.RoleOwner},
	{email: developerEmail, roleName: "team_developer", teamRole: teamModel.RoleMember},
}

func (s *Seeder) SeedDefaultTeam(ctx context.Context) error {
	if s.teamRepo == nil {
		return nil
	}

	log.Println("Seeding default team...")

	teamRecord, err := s.teamRepo.GetBySlug(ctx, defaultTeamSlug)
	if err != nil {
		return err
	}
	if teamRecord == nil {
		teamRecord = &teamModel.Team{
			Name:        defaultTeamName,
			Slug:        defaultTeamSlug,
			Description: "Default team for local development",
		}
		if err := s.teamRepo.Create(ctx, teamRecord); err != nil {
			return err
		}
		log.Printf("Created default team %s (id=%s)", defaultTeamSlug, teamRecord.ID)
	} else {
		log.Printf("Default team %s already exists (id=%s)", defaultTeamSlug, teamRecord.ID)
	}

	for _, member := range defaultTeamMembers {
		if err := s.seedDefaultTeamMember(ctx, teamRecord.ID, member.email, member.roleName, member.teamRole); err != nil {
			return err
		}
	}

	return nil
}

func (s *Seeder) seedDefaultTeamMember(ctx context.Context, teamID, email, roleName, teamRole string) error {
	userRecord, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if userRecord == nil {
		log.Printf("Skipping default team membership for %s because user does not exist", email)
		return nil
	}

	existingMember, err := s.teamRepo.GetMember(ctx, teamID, userRecord.ID)
	if err != nil {
		return err
	}
	if existingMember == nil {
		if err := s.teamRepo.AddMember(ctx, &teamModel.TeamMember{
			ID:        uuid.New().String(),
			TeamID:    teamID,
			UserID:    userRecord.ID,
			Role:      teamRole,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}); err != nil {
			return err
		}
		log.Printf("Added %s to default team as %s", email, teamRole)
	} else {
		log.Printf("User %s is already a member of default team", email)
	}

	roleRecord, err := s.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return err
	}
	if roleRecord == nil {
		log.Printf("Skipping team role assignment for %s because role %s does not exist", email, roleName)
		return nil
	}

	userRoles, err := s.roleRepo.GetUserRolesByTeam(ctx, userRecord.ID, teamID)
	if err != nil {
		return err
	}
	for _, userRole := range userRoles {
		if userRole.RoleID == roleRecord.ID && userRole.TeamID == teamID {
			log.Printf("Role %s already assigned to %s for default team", roleName, email)
			return nil
		}
	}

	if err := s.roleRepo.AssignRole(ctx, &roleModel.UserRole{
		ID:        uuid.New().String(),
		UserID:    userRecord.ID,
		RoleID:    roleRecord.ID,
		TeamID:    teamID,
		GrantedAt: time.Now(),
	}); err != nil {
		return err
	}

	log.Printf("Assigned role %s to %s for default team", roleName, email)
	return nil
}
