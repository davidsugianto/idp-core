package team

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/team"
	"gorm.io/gorm"
)

// Repository defines the interface for team persistence operations
type Repository interface {
	// Team CRUD
	Create(ctx context.Context, t *team.Team) error
	GetByID(ctx context.Context, id string) (*team.Team, error)
	GetBySlug(ctx context.Context, slug string) (*team.Team, error)
	List(ctx context.Context, limit, offset int) ([]team.Team, int64, error)
	ListByStatus(ctx context.Context, status string) ([]team.Team, error)
	Update(ctx context.Context, t *team.Team) error
	SoftDelete(ctx context.Context, id string) error

	// Team Member operations
	AddMember(ctx context.Context, member *team.TeamMember) error
	GetMember(ctx context.Context, teamID, userID string) (*team.TeamMember, error)
	ListMembers(ctx context.Context, teamID string) ([]team.TeamMember, error)
	ListTeamsByUser(ctx context.Context, userID string) ([]team.TeamMember, error)
	UpdateMemberRole(ctx context.Context, teamID, userID, role string) error
	RemoveMember(ctx context.Context, teamID, userID string) error
	IsTeamMember(ctx context.Context, teamID, userID string) (bool, string, error)
}

type repository struct {
	db *gorm.DB
}

// Dependencies holds repository dependencies
type Dependencies struct {
	Database *gorm.DB
}

// New creates a new team repository
func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}
