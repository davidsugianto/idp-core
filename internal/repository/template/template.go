package template

import (
	"context"

	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
)

func (r *repository) Create(ctx context.Context, template *templateModel.Template) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*templateModel.Template, error) {
	var template templateModel.Template
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *repository) List(ctx context.Context, req *templateModel.ListTemplatesRequest) ([]templateModel.Template, int64, error) {
	query := r.db.WithContext(ctx).Model(&templateModel.Template{})

	if req.TeamID != "" {
		query = query.Where("team_id = ?", req.TeamID)
	}
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Visibility != "" {
		query = query.Where("visibility = ?", req.Visibility)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ? OR slug ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var templates []templateModel.Template
	query = query.Order("created_at DESC")
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	if err := query.Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

func (r *repository) Update(ctx context.Context, template *templateModel.Template) error {
	return r.db.WithContext(ctx).Save(template).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&templateModel.Template{}).Error
}

func (r *repository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&templateModel.Template{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
