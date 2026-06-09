package template

import (
	"context"

	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
	"gorm.io/gorm"
)

func (r *repository) ListResourcesByVersion(ctx context.Context, versionID string) ([]templateModel.TemplateResource, error) {
	var resources []templateModel.TemplateResource
	if err := r.db.WithContext(ctx).
		Where("version_id = ?", versionID).
		Order("`order` ASC, created_at ASC").
		Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

func (r *repository) ReplaceResources(ctx context.Context, templateID, versionID string, resources []templateModel.TemplateResource) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ? AND version_id = ?", templateID, versionID).Delete(&templateModel.TemplateResource{}).Error; err != nil {
			return err
		}

		if len(resources) == 0 {
			return nil
		}

		return tx.Create(&resources).Error
	})
}
