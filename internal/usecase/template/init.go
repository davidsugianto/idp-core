package template

import (
	"context"
	"errors"
	"regexp"

	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
	templateRepo "github.com/davidsugianto/idp-core/internal/repository/template"
)

var (
	ErrTemplateNotFound           = errors.New("template not found")
	ErrTemplateNameRequired       = errors.New("template name is required")
	ErrTemplateExists             = errors.New("template already exists")
	ErrTemplateVersionExists      = errors.New("template version already exists")
	ErrInvalidVisibility          = errors.New("invalid visibility")
	ErrInvalidStatus              = errors.New("invalid status")
	ErrTemplateVersionNotFound    = errors.New("template version not found")
	ErrTemplateVersionRequired    = errors.New("template version is required")
	ErrInvalidLifecycleTransition = errors.New("invalid lifecycle transition")
	ErrTemplateParameterRequired  = errors.New("template parameter name is required")
	ErrTemplateParameterDuplicate = errors.New("template parameter name already exists")
	ErrTemplateResourceRequired   = errors.New("template resource name is required")
	ErrTemplateInputInvalid       = errors.New("template inputs are invalid")
)

var templateParameterNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

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
	ReplaceParameters(ctx context.Context, templateID, versionID string, req *templateModel.ReplaceTemplateParametersRequest) ([]templateModel.TemplateParameter, error)
	ReplaceResources(ctx context.Context, templateID, versionID string, req *templateModel.ReplaceTemplateResourcesRequest) ([]templateModel.TemplateResource, error)
	ValidateVersionInputs(ctx context.Context, templateID, versionID string, req *templateModel.ValidateTemplateVersionRequest) (*templateModel.ValidateTemplateVersionResponse, error)
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
