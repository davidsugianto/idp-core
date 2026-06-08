package template

import (
	"context"

	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, template *templateModel.Template) error
	GetByID(ctx context.Context, id string) (*templateModel.Template, error)
	List(ctx context.Context, req *templateModel.ListTemplatesRequest) ([]templateModel.Template, int64, error)
	Update(ctx context.Context, template *templateModel.Template) error
	Delete(ctx context.Context, id string) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	CreateVersion(ctx context.Context, version *templateModel.TemplateVersion) error
	GetVersionByID(ctx context.Context, id string) (*templateModel.TemplateVersion, error)
	GetVersionByTemplateAndVersion(ctx context.Context, templateID, version string) (*templateModel.TemplateVersion, error)
	ListVersionsByTemplate(ctx context.Context, templateID string, req *templateModel.ListTemplateVersionsRequest) ([]templateModel.TemplateVersion, int64, error)
	UpdateVersion(ctx context.Context, version *templateModel.TemplateVersion) error
	ClearLatestVersion(ctx context.Context, templateID string) error

	CreateParameter(ctx context.Context, parameter *templateModel.TemplateParameter) error
	CreateResource(ctx context.Context, resource *templateModel.TemplateResource) error
	CreateInstance(ctx context.Context, instance *templateModel.TemplateInstance) error
}

type repository struct {
	db *gorm.DB
}

type Dependencies struct {
	Database *gorm.DB
}

func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}
