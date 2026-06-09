package template

import (
	"context"

	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
	"gorm.io/gorm"
)

func (r *repository) ListParametersByVersion(ctx context.Context, versionID string) ([]templateModel.TemplateParameter, error) {
	var parameters []templateModel.TemplateParameter
	if err := r.db.WithContext(ctx).
		Where("version_id = ?", versionID).
		Order("`order` ASC, created_at ASC").
		Find(&parameters).Error; err != nil {
		return nil, err
	}
	return parameters, nil
}

func (r *repository) ReplaceParameters(ctx context.Context, templateID, versionID string, parameters []templateModel.TemplateParameter) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ? AND version_id = ?", templateID, versionID).Delete(&templateModel.TemplateParameter{}).Error; err != nil {
			return err
		}

		if len(parameters) == 0 {
			return nil
		}

		return tx.Create(&parameters).Error
	})
}
