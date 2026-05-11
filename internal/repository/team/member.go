package team

import (
	"context"
	"errors"

	"github.com/davidsugianto/idp-core/internal/model/team"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AddMember adds a member to a team
func (r *repository) AddMember(ctx context.Context, member *team.TeamMember) error {
	if member.ID == "" {
		member.ID = uuid.New().String()
	}
	if member.Role == "" {
		member.Role = team.RoleMember
	}
	return r.db.WithContext(ctx).Create(member).Error
}

// GetMember retrieves a specific team member
func (r *repository) GetMember(ctx context.Context, teamID, userID string) (*team.TeamMember, error) {
	var member team.TeamMember
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

// ListMembers retrieves all members of a team
func (r *repository) ListMembers(ctx context.Context, teamID string) ([]team.TeamMember, error) {
	var members []team.TeamMember
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

// ListTeamsByUser retrieves all teams a user belongs to
func (r *repository) ListTeamsByUser(ctx context.Context, userID string) ([]team.TeamMember, error) {
	var members []team.TeamMember
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&members).Error
	return members, err
}

// UpdateMemberRole updates a team member's role
func (r *repository) UpdateMemberRole(ctx context.Context, teamID, userID, role string) error {
	return r.db.WithContext(ctx).
		Model(&team.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Update("role", role).Error
}

// RemoveMember removes a member from a team
func (r *repository) RemoveMember(ctx context.Context, teamID, userID string) error {
	return r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&team.TeamMember{}).Error
}

// IsTeamMember checks if a user is a member of a team and returns their role
func (r *repository) IsTeamMember(ctx context.Context, teamID, userID string) (bool, string, error) {
	member, err := r.GetMember(ctx, teamID, userID)
	if err != nil {
		return false, "", err
	}
	if member == nil {
		return false, "", nil
	}
	return true, member.Role, nil
}
