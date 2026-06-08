package template

import (
	"context"
	"errors"

	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
	templateRepo "github.com/davidsugianto/idp-core/internal/repository/template"
)

var (
	ErrTemplateNotFound        = errors.New("template not found")
	ErrTemplateNameRequired    = errors.New("template name is required")
	ErrTemplateExists          = errors.New("template already exists")
	ErrTemplateVersionExists   = errors.New("template version already exists")
	ErrInvalidVisibility       = errors.New("invalid visibility")
	ErrInvalidStatus           = errors.New("invalid status")
	ErrTemplateVersionNotFound = errors.New("template version not found")
	ErrTemplateVersionRequired = errors.New("template version is required")
)

type Usecase interface {
	Create(ctx context.Context, req *templateModel.CreateTemplateRequest) (*templateModel.TemplateResponse, error)
	Get(ctx context.Context, id string) (*templateModel.TemplateResponse, error)
	List(ctx context.Context, req *templateModel.ListTemplatesRequest) (*templateModel.TemplateListResponse, error)
	Update(ctx context.Context, id string, req *templateModel.UpdateTemplateRequest) (*templateModel.TemplateResponse, error)
	Delete(ctx context.Context, id string) error

	CreateVersion(ctx context.Context, templateID string, req *templateModel.CreateTemplateVersionRequest) (*templateModel.TemplateVersionResponse, error)
	GetVersion(ctx context.Context, versionID string) (*templateModel.TemplateVersionResponse, error)
	ListVersions(ctx context.Context, templateID string, req *templateModel.ListTemplateVersionsRequest) (*templateModel.TemplateVersionListResponse, error)
	UpdateVersion(ctx context.Context, versionID string, req *templateModel.UpdateTemplateVersionRequest) (*templateModel.TemplateVersionResponse, error)
}

type usecase struct {
	templateRepo templateRepo.Repository
}

type Dependencies struct {
	TemplateRepo templateRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{templateRepo: deps.TemplateRepo}
}
