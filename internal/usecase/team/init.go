package team

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/davidsugianto/idp-core/internal/model/team"
	teamRepo "github.com/davidsugianto/idp-core/internal/repository/team"
	userRepo "github.com/davidsugianto/idp-core/internal/repository/user"
)

var (
	ErrTeamNotFound       = errors.New("team not found")
	ErrTeamAlreadyExists  = errors.New("team already exists")
	ErrInvalidTeamID      = errors.New("invalid team id")
	ErrInvalidSlug        = errors.New("invalid slug format")
	ErrMemberNotFound     = errors.New("team member not found")
	ErrMemberAlreadyExists = errors.New("user is already a team member")
	ErrCannotRemoveOwner  = errors.New("cannot remove the last owner")
)

// Usecase defines the interface for team business logic
type Usecase interface {
	// Team CRUD
	Create(ctx context.Context, req team.CreateTeamRequest) (*team.Team, error)
	Get(ctx context.Context, id string) (*team.Team, error)
	GetBySlug(ctx context.Context, slug string) (*team.Team, error)
	List(ctx context.Context, limit, offset int) (*team.TeamListResponse, error)
	Update(ctx context.Context, id string, req team.UpdateTeamRequest) (*team.Team, error)
	Delete(ctx context.Context, id string) error

	// Team Members
	AddMember(ctx context.Context, teamID string, req team.AddTeamMemberRequest) (*team.TeamMember, error)
	UpdateMember(ctx context.Context, teamID, userID string, req team.UpdateTeamMemberRequest) error
	RemoveMember(ctx context.Context, teamID, userID string) error
	ListMembers(ctx context.Context, teamID string) (*team.TeamWithMembersResponse, error)
	ListUserTeams(ctx context.Context, userID string) ([]team.Team, error)
}

type usecase struct {
	teamRepo teamRepo.Repository
	userRepo userRepo.Repository
}

// Dependencies holds usecase dependencies
type Dependencies struct {
	TeamRepo teamRepo.Repository
	UserRepo userRepo.Repository
}

// New creates a new team usecase
func New(deps Dependencies) Usecase {
	return &usecase{
		teamRepo: deps.TeamRepo,
		userRepo: deps.UserRepo,
	}
}

// slugRegex validates slug format (lowercase alphanumeric and hyphens)
var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// isValidSlug checks if a slug is valid
func isValidSlug(slug string) bool {
	if len(slug) < 2 || len(slug) > 63 {
		return false
	}
	return slugRegex.MatchString(slug)
}

// Create creates a new team
func (u *usecase) Create(ctx context.Context, req team.CreateTeamRequest) (*team.Team, error) {
	// Validate slug
	slug := strings.ToLower(req.Slug)
	if !isValidSlug(slug) {
		return nil, ErrInvalidSlug
	}

	// Check if team already exists
	existing, err := u.teamRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrTeamAlreadyExists
	}

	// Create new team
	newTeam := &team.Team{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
	}

	if err := u.teamRepo.Create(ctx, newTeam); err != nil {
		return nil, err
	}

	return newTeam, nil
}

// Get retrieves a team by ID
func (u *usecase) Get(ctx context.Context, id string) (*team.Team, error) {
	if id == "" {
		return nil, ErrInvalidTeamID
	}

	t, err := u.teamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTeamNotFound
	}

	return t, nil
}

// GetBySlug retrieves a team by slug
func (u *usecase) GetBySlug(ctx context.Context, slug string) (*team.Team, error) {
	t, err := u.teamRepo.GetBySlug(ctx, strings.ToLower(slug))
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTeamNotFound
	}
	return t, nil
}

// List retrieves a paginated list of teams
func (u *usecase) List(ctx context.Context, limit, offset int) (*team.TeamListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	teams, total, err := u.teamRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return team.ToTeamListResponse(teams, total), nil
}

// Update updates a team
func (u *usecase) Update(ctx context.Context, id string, req team.UpdateTeamRequest) (*team.Team, error) {
	t, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		t.Name = req.Name
	}
	if req.Description != "" {
		t.Description = req.Description
	}
	if req.Status != "" {
		t.Status = req.Status
	}

	if err := u.teamRepo.Update(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

// Delete soft deletes a team
func (u *usecase) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidTeamID
	}

	t, err := u.teamRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if t == nil {
		return ErrTeamNotFound
	}

	return u.teamRepo.SoftDelete(ctx, id)
}

// AddMember adds a member to a team
func (u *usecase) AddMember(ctx context.Context, teamID string, req team.AddTeamMemberRequest) (*team.TeamMember, error) {
	// Verify team exists
	_, err := u.Get(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Verify user exists
	usr, err := u.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if usr == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if already a member
	existing, err := u.teamRepo.GetMember(ctx, teamID, req.UserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrMemberAlreadyExists
	}

	// Add member
	member := &team.TeamMember{
		TeamID: teamID,
		UserID: req.UserID,
		Role:   req.Role,
	}

	if err := u.teamRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// UpdateMember updates a team member's role
func (u *usecase) UpdateMember(ctx context.Context, teamID, userID string, req team.UpdateTeamMemberRequest) error {
	// Verify team exists
	_, err := u.Get(ctx, teamID)
	if err != nil {
		return err
	}

	// Check if member exists
	member, err := u.teamRepo.GetMember(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrMemberNotFound
	}

	return u.teamRepo.UpdateMemberRole(ctx, teamID, userID, req.Role)
}

// RemoveMember removes a member from a team
func (u *usecase) RemoveMember(ctx context.Context, teamID, userID string) error {
	// Verify team exists
	_, err := u.Get(ctx, teamID)
	if err != nil {
		return err
	}

	// Check if member exists
	member, err := u.teamRepo.GetMember(ctx, teamID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrMemberNotFound
	}

	// Prevent removing the last owner
	if member.Role == team.RoleOwner {
		members, err := u.teamRepo.ListMembers(ctx, teamID)
		if err != nil {
			return err
		}
		ownerCount := 0
		for _, m := range members {
			if m.Role == team.RoleOwner {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return ErrCannotRemoveOwner
		}
	}

	return u.teamRepo.RemoveMember(ctx, teamID, userID)
}

// ListMembers retrieves all members of a team
func (u *usecase) ListMembers(ctx context.Context, teamID string) (*team.TeamWithMembersResponse, error) {
	// Get team
	t, err := u.Get(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Get members
	members, err := u.teamRepo.ListMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Build response with user info
	memberResponses := make([]team.TeamMemberResponse, len(members))
	for i, m := range members {
		usr, _ := u.userRepo.GetByID(ctx, m.UserID)
		memberResponses[i] = team.TeamMemberResponse{
			ID:        m.ID,
			Role:      m.Role,
			CreatedAt: m.CreatedAt,
		}
		if usr != nil {
			memberResponses[i].User = &team.UserBasicInfo{
				ID:    usr.ID,
				Email: usr.Email,
				Name:  usr.Name,
			}
		}
	}

	return &team.TeamWithMembersResponse{
		TeamResponse: *team.ToTeamResponse(t),
		Members:      memberResponses,
	}, nil
}

// ListUserTeams retrieves all teams a user belongs to
func (u *usecase) ListUserTeams(ctx context.Context, userID string) ([]team.Team, error) {
	memberships, err := u.teamRepo.ListTeamsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	teams := make([]team.Team, len(memberships))
	for i, m := range memberships {
		t, err := u.teamRepo.GetByID(ctx, m.TeamID)
		if err != nil {
			return nil, err
		}
		if t != nil {
			teams[i] = *t
		}
	}

	return teams, nil
}
