package template

import (
	"context"
	"time"

	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (u *usecase) Create(ctx context.Context, req *templateModel.CreateTemplateRequest) (*templateModel.TemplateResponse, error) {
	if req.Name == "" {
		return nil, ErrTemplateNameRequired
	}

	slug := templateModel.GenerateSlug(req.Name)
	exists, err := u.templateRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTemplateExists
	}

	visibility := req.Visibility
	if visibility == "" {
		visibility = templateModel.VisibilityTeam
	}
	if !templateModel.ValidVisibility(visibility) {
		return nil, ErrInvalidVisibility
	}

	now := time.Now()
	template := &templateModel.Template{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Category:    req.Category,
		Author:      req.Author,
		AuthorEmail: req.AuthorEmail,
		Visibility:  visibility,
		TeamID:      req.TeamID,
		Status:      templateModel.StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	return templateModel.ToTemplateResponse(template), nil
}

func (u *usecase) Get(ctx context.Context, id string) (*templateModel.TemplateResponse, error) {
	template, err := u.templateRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}

	return templateModel.ToTemplateResponse(template), nil
}

func (u *usecase) List(ctx context.Context, req *templateModel.ListTemplatesRequest) (*templateModel.TemplateListResponse, error) {
	templates, total, err := u.templateRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	return templateModel.ToTemplateListResponse(templates, total), nil
}

func (u *usecase) Update(ctx context.Context, id string, req *templateModel.UpdateTemplateRequest) (*templateModel.TemplateResponse, error) {
	template, err := u.templateRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		template.Name = *req.Name
		template.Slug = templateModel.GenerateSlug(*req.Name)
	}
	if req.Description != nil {
		template.Description = *req.Description
	}
	if req.Category != nil {
		template.Category = *req.Category
	}
	if req.Author != nil {
		template.Author = *req.Author
	}
	if req.AuthorEmail != nil {
		template.AuthorEmail = *req.AuthorEmail
	}
	if req.TeamID != nil {
		template.TeamID = *req.TeamID
	}
	if req.Visibility != nil {
		if !templateModel.ValidVisibility(*req.Visibility) {
			return nil, ErrInvalidVisibility
		}
		template.Visibility = *req.Visibility
	}
	if req.Status != nil {
		if !templateModel.ValidStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		template.Status = *req.Status
	}

	template.UpdatedAt = time.Now()
	if err := u.templateRepo.Update(ctx, template); err != nil {
		return nil, err
	}

	return templateModel.ToTemplateResponse(template), nil
}

func (u *usecase) Delete(ctx context.Context, id string) error {
	_, err := u.templateRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTemplateNotFound
		}
		return err
	}

	return u.templateRepo.Delete(ctx, id)
}

func (u *usecase) CreateVersion(ctx context.Context, templateID string, req *templateModel.CreateTemplateVersionRequest) (*templateModel.TemplateVersionResponse, error) {
	if req.Version == "" {
		return nil, ErrTemplateVersionRequired
	}

	_, err := u.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}

	existing, err := u.templateRepo.GetVersionByTemplateAndVersion(ctx, templateID, req.Version)
	if err == nil && existing != nil {
		return nil, ErrTemplateVersionExists
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = templateModel.StatusDraft
	}
	if !templateModel.ValidStatus(status) {
		return nil, ErrInvalidStatus
	}

	isLatest := true
	if req.IsLatest != nil {
		isLatest = *req.IsLatest
	}
	if isLatest {
		if err := u.templateRepo.ClearLatestVersion(ctx, templateID); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	version := &templateModel.TemplateVersion{
		ID:          uuid.New().String(),
		TemplateID:  templateID,
		Version:     req.Version,
		Description: req.Description,
		Changelog:   req.Changelog,
		IsLatest:    isLatest,
		IsStable:    req.IsStable,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.templateRepo.CreateVersion(ctx, version); err != nil {
		return nil, err
	}

	return templateModel.ToTemplateVersionResponse(version), nil
}

func (u *usecase) GetVersion(ctx context.Context, versionID string) (*templateModel.TemplateVersionResponse, error) {
	version, err := u.templateRepo.GetVersionByID(ctx, versionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTemplateVersionNotFound
		}
		return nil, err
	}

	return templateModel.ToTemplateVersionResponse(version), nil
}

func (u *usecase) ListVersions(ctx context.Context, templateID string, req *templateModel.ListTemplateVersionsRequest) (*templateModel.TemplateVersionListResponse, error) {
	versions, total, err := u.templateRepo.ListVersionsByTemplate(ctx, templateID, req)
	if err != nil {
		return nil, err
	}

	return templateModel.ToTemplateVersionListResponse(versions, total), nil
}

func (u *usecase) UpdateVersion(ctx context.Context, versionID string, req *templateModel.UpdateTemplateVersionRequest) (*templateModel.TemplateVersionResponse, error) {
	version, err := u.templateRepo.GetVersionByID(ctx, versionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTemplateVersionNotFound
		}
		return nil, err
	}

	if req.Description != nil {
		version.Description = *req.Description
	}
	if req.Changelog != nil {
		version.Changelog = *req.Changelog
	}
	if req.IsLatest != nil {
		if *req.IsLatest {
			if err := u.templateRepo.ClearLatestVersion(ctx, version.TemplateID); err != nil {
				return nil, err
			}
		}
		version.IsLatest = *req.IsLatest
	}
	if req.IsStable != nil {
		version.IsStable = *req.IsStable
	}
	if req.Status != nil {
		if !templateModel.ValidStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		version.Status = *req.Status
	}

	version.UpdatedAt = time.Now()
	if err := u.templateRepo.UpdateVersion(ctx, version); err != nil {
		return nil, err
	}

	return templateModel.ToTemplateVersionResponse(version), nil
}
