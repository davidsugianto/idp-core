package template

import (
	"context"

	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
)

func (r *repository) CreateInstance(ctx context.Context, instance *templateModel.TemplateInstance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

func (r *repository) GetInstanceByID(ctx context.Context, id string) (*templateModel.TemplateInstance, error) {
	var instance templateModel.TemplateInstance
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (r *repository) ListInstancesByEnvironment(ctx context.Context, environmentID string) ([]templateModel.TemplateInstance, error) {
	var instances []templateModel.TemplateInstance
	if err := r.db.WithContext(ctx).Where("environment_id = ?", environmentID).Order("created_at DESC").Find(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}
