package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
	templateUsecase "github.com/davidsugianto/idp-core/internal/usecase/template"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type stubTemplateUsecase struct {
	createVersionResult *templateModel.TemplateVersionResponse
	createVersionErr    error
	updateVersionResult *templateModel.TemplateVersionResponse
	updateVersionErr    error
}

func (s *stubTemplateUsecase) Create(ctx context.Context, req *templateModel.CreateTemplateRequest) (*templateModel.TemplateResponse, error) {
	return nil, nil
}

func (s *stubTemplateUsecase) Get(ctx context.Context, id string) (*templateModel.TemplateResponse, error) {
	return nil, nil
}

func (s *stubTemplateUsecase) List(ctx context.Context, req *templateModel.ListTemplatesRequest) (*templateModel.TemplateListResponse, error) {
	return nil, nil
}

func (s *stubTemplateUsecase) Update(ctx context.Context, id string, req *templateModel.UpdateTemplateRequest) (*templateModel.TemplateResponse, error) {
	return nil, nil
}

func (s *stubTemplateUsecase) Delete(ctx context.Context, id string) error {
	return nil
}

func (s *stubTemplateUsecase) CreateVersion(ctx context.Context, templateID string, req *templateModel.CreateTemplateVersionRequest) (*templateModel.TemplateVersionResponse, error) {
	return s.createVersionResult, s.createVersionErr
}

func (s *stubTemplateUsecase) GetVersion(ctx context.Context, versionID string) (*templateModel.TemplateVersionResponse, error) {
	return nil, nil
}

func (s *stubTemplateUsecase) ListVersions(ctx context.Context, templateID string, req *templateModel.ListTemplateVersionsRequest) (*templateModel.TemplateVersionListResponse, error) {
	return nil, nil
}

func (s *stubTemplateUsecase) UpdateVersion(ctx context.Context, versionID string, req *templateModel.UpdateTemplateVersionRequest) (*templateModel.TemplateVersionResponse, error) {
	return s.updateVersionResult, s.updateVersionErr
}

func (s *stubTemplateUsecase) ReplaceParameters(ctx context.Context, templateID, versionID string, req *templateModel.ReplaceTemplateParametersRequest) ([]templateModel.TemplateParameter, error) {
	return nil, nil
}

func (s *stubTemplateUsecase) ReplaceResources(ctx context.Context, templateID, versionID string, req *templateModel.ReplaceTemplateResourcesRequest) ([]templateModel.TemplateResource, error) {
	return nil, nil
}

func (s *stubTemplateUsecase) ValidateVersionInputs(ctx context.Context, templateID, versionID string, req *templateModel.ValidateTemplateVersionRequest) (*templateModel.ValidateTemplateVersionResponse, error) {
	return &templateModel.ValidateTemplateVersionResponse{Valid: true, Errors: []string{}}, nil
}

func newTemplateTestHandler(templateUC templateUsecase.Usecase) *Handler {
	gin.SetMode(gin.TestMode)

	return New(Dependencies{
		TemplateUseCase:   templateUC,
		AuthConfig:        &config.AuthConfig{JWTSecret: "test-secret"},
		WebhookValidator:  webhook.NewValidator(),
		AllowedOrigins:    []string{"*"},
		OIDCEndSessionURL: "http://localhost/logout",
	})
}

func TestTemplateHandlerExposesPhase3Endpoints(t *testing.T) {
	handlerType := reflect.TypeOf(&Handler{})

	requiredMethods := []string{
		"ReplaceTemplateParameters",
		"ReplaceTemplateResources",
		"ValidateTemplateVersion",
	}

	for _, methodName := range requiredMethods {
		t.Run(methodName, func(t *testing.T) {
			_, ok := handlerType.MethodByName(methodName)
			assert.True(t, ok, "expected Handler to expose %s for Phase 3 template contracts", methodName)
		})
	}
}

func TestCreateTemplateVersionReturnsConflictOnDuplicateVersion(t *testing.T) {
	handler := newTemplateTestHandler(&stubTemplateUsecase{
		createVersionErr: templateUsecase.ErrTemplateVersionExists,
	})

	router := gin.New()
	router.POST("/v1/templates/:id/versions", handler.CreateTemplateVersion)

	body, err := json.Marshal(templateModel.CreateTemplateVersionRequest{
		Version: "v1.0.0",
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/templates/template-1/versions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUpdateTemplateVersionReturnsConflictOnInvalidLifecycleTransition(t *testing.T) {
	handler := newTemplateTestHandler(&stubTemplateUsecase{
		updateVersionErr: templateUsecase.ErrInvalidLifecycleTransition,
	})

	router := gin.New()
	router.PATCH("/v1/templates/:id/versions/:versionId", handler.UpdateTemplateVersion)

	status := templateModel.StatusStable
	body, err := json.Marshal(templateModel.UpdateTemplateVersionRequest{
		Status: &status,
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/v1/templates/template-1/versions/version-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}
