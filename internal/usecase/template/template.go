package template

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
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
		if !canTransitionTemplateVersion(version.Status, *req.Status) {
			return nil, ErrInvalidLifecycleTransition
		}
		version.Status = *req.Status
	}

	version.UpdatedAt = time.Now()
	if err := u.templateRepo.UpdateVersion(ctx, version); err != nil {
		return nil, err
	}

	return templateModel.ToTemplateVersionResponse(version), nil
}

func (u *usecase) ReplaceParameters(ctx context.Context, templateID, versionID string, req *templateModel.ReplaceTemplateParametersRequest) ([]templateModel.TemplateParameter, error) {
	if _, err := u.ensureVersionBelongsToTemplate(ctx, templateID, versionID); err != nil {
		return nil, err
	}
	if req == nil {
		return []templateModel.TemplateParameter{}, nil
	}

	parameters := make([]templateModel.TemplateParameter, len(req.Parameters))
	seenNames := make(map[string]struct{}, len(req.Parameters))
	now := time.Now()

	for i, parameter := range req.Parameters {
		name := strings.TrimSpace(parameter.Name)
		if name == "" {
			return nil, ErrTemplateParameterRequired
		}
		if _, exists := seenNames[name]; exists {
			return nil, ErrTemplateParameterDuplicate
		}
		seenNames[name] = struct{}{}

		if parameter.DisplayName == "" {
			parameter.DisplayName = name
		}
		parameter.ID = uuid.New().String()
		parameter.TemplateID = templateID
		parameter.VersionID = versionID
		parameter.Name = name
		parameter.Order = i
		parameter.CreatedAt = now
		parameters[i] = parameter
	}

	if err := u.templateRepo.ReplaceParameters(ctx, templateID, versionID, parameters); err != nil {
		return nil, err
	}

	return parameters, nil
}

func (u *usecase) ReplaceResources(ctx context.Context, templateID, versionID string, req *templateModel.ReplaceTemplateResourcesRequest) ([]templateModel.TemplateResource, error) {
	if _, err := u.ensureVersionBelongsToTemplate(ctx, templateID, versionID); err != nil {
		return nil, err
	}
	if req == nil {
		return []templateModel.TemplateResource{}, nil
	}

	resources := make([]templateModel.TemplateResource, len(req.Resources))
	now := time.Now()

	for i, resource := range req.Resources {
		resourceName := strings.TrimSpace(resource.Name)
		if resourceName == "" {
			return nil, ErrTemplateResourceRequired
		}
		resource.ID = uuid.New().String()
		resource.TemplateID = templateID
		resource.VersionID = versionID
		resource.Name = resourceName
		resource.Order = i
		resource.CreatedAt = now
		resource.UpdatedAt = now
		resources[i] = resource
	}

	if err := u.templateRepo.ReplaceResources(ctx, templateID, versionID, resources); err != nil {
		return nil, err
	}

	return resources, nil
}

func (u *usecase) ValidateVersionInputs(ctx context.Context, templateID, versionID string, req *templateModel.ValidateTemplateVersionRequest) (*templateModel.ValidateTemplateVersionResponse, error) {
	if _, err := u.ensureVersionBelongsToTemplate(ctx, templateID, versionID); err != nil {
		return nil, err
	}

	parameters, err := u.templateRepo.ListParametersByVersion(ctx, versionID)
	if err != nil {
		return nil, err
	}

	inputMap := map[string]any{}
	if req != nil && req.Inputs != nil {
		inputMap = req.Inputs
	}

	errors := make([]string, 0)
	for _, parameter := range parameters {
		value, ok := inputMap[parameter.Name]
		if !ok || value == nil || strings.TrimSpace(strings.TrimSpace(toString(value))) == "" {
			if parameter.Required {
				errors = append(errors, parameter.Name+" is required")
			}
			continue
		}

		if parameter.Validation != "" {
			var constraints map[string]any
			if err := json.Unmarshal([]byte(parameter.Validation), &constraints); err == nil {
				if rawPattern, ok := constraints["pattern"].(string); ok && rawPattern != "" {
					matcher, compileErr := regexp.Compile(rawPattern)
					if compileErr == nil && !matcher.MatchString(toString(value)) {
						errors = append(errors, parameter.Name+" does not match required pattern")
					}
				}
			}
		}
	}

	return &templateModel.ValidateTemplateVersionResponse{
		Valid:  len(errors) == 0,
		Errors: errors,
	}, nil
}

func (u *usecase) ensureVersionBelongsToTemplate(ctx context.Context, templateID, versionID string) (*templateModel.TemplateVersion, error) {
	if _, err := u.templateRepo.GetByID(ctx, templateID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}

	version, err := u.templateRepo.GetVersionByID(ctx, versionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTemplateVersionNotFound
		}
		return nil, err
	}
	if version.TemplateID != templateID {
		return nil, ErrTemplateVersionNotFound
	}

	return version, nil
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(bytes)
	}
}

func canTransitionTemplateVersion(currentStatus, nextStatus string) bool {
	if currentStatus == nextStatus {
		return true
	}

	switch currentStatus {
	case templateModel.StatusDraft:
		return nextStatus == templateModel.StatusStable || nextStatus == templateModel.StatusDeprecated
	case templateModel.StatusStable:
		return nextStatus == templateModel.StatusDeprecated
	case templateModel.StatusDeprecated:
		return false
	default:
		return false
	}
}
