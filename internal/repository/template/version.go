package template

import (
	"context"

	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
)

func (r *repository) CreateVersion(ctx context.Context, version *templateModel.TemplateVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

func (r *repository) GetVersionByID(ctx context.Context, id string) (*templateModel.TemplateVersion, error) {
	var version templateModel.TemplateVersion
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *repository) GetVersionByTemplateAndVersion(ctx context.Context, templateID, versionValue string) (*templateModel.TemplateVersion, error) {
	var version templateModel.TemplateVersion
	if err := r.db.WithContext(ctx).Where("template_id = ? AND version = ?", templateID, versionValue).First(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *repository) ListVersionsByTemplate(ctx context.Context, templateID string, req *templateModel.ListTemplateVersionsRequest) ([]templateModel.TemplateVersion, int64, error) {
	query := r.db.WithContext(ctx).Model(&templateModel.TemplateVersion{}).Where("template_id = ?", templateID)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var versions []templateModel.TemplateVersion
	query = query.Order("created_at DESC")
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	if err := query.Find(&versions).Error; err != nil {
		return nil, 0, err
	}
	return versions, total, nil
}

func (r *repository) UpdateVersion(ctx context.Context, version *templateModel.TemplateVersion) error {
	return r.db.WithContext(ctx).Save(version).Error
}

func (r *repository) ClearLatestVersion(ctx context.Context, templateID string) error {
	return r.db.WithContext(ctx).Model(&templateModel.TemplateVersion{}).Where("template_id = ?", templateID).Update("is_latest", false).Error
}

func (r *repository) CreateParameter(ctx context.Context, parameter *templateModel.TemplateParameter) error {
	return r.db.WithContext(ctx).Create(parameter).Error
}

func (r *repository) CreateResource(ctx context.Context, resource *templateModel.TemplateResource) error {
	return r.db.WithContext(ctx).Create(resource).Error
}

