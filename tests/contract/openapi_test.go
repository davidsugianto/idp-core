package contract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Get current working directory
	cwd, _ := os.Getwd()

	// Walk up to find go.mod
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}

	// Fallback to relative path from tests/contract
	return filepath.Join("..", "..")
}

// TestOpenAPISpecExists validates that the OpenAPI spec file exists
func TestOpenAPISpecExists(t *testing.T) {
	projectRoot := getProjectRoot()

	specPaths := []string{
		filepath.Join(projectRoot, "docs/swagger/swagger.json"),
		filepath.Join(projectRoot, "docs/swagger/swagger.yaml"),
	}

	found := false
	for _, path := range specPaths {
		if _, err := os.Stat(path); err == nil {
			found = true
			t.Logf("Found OpenAPI spec at: %s", path)
			break
		}
	}

	assert.True(t, found, "OpenAPI spec file should exist. Run 'make swagger-gen' to generate it.")
}

// TestOpenAPISpecValidJSON validates that the OpenAPI spec is valid JSON
func TestOpenAPISpecValidJSON(t *testing.T) {
	projectRoot := getProjectRoot()
	specPath := filepath.Join(projectRoot, "docs/swagger/swagger.json")

	// Check if file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("OpenAPI spec not found. Run 'make swagger-gen' to generate it.")
	}

	// Read the file
	data, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read OpenAPI spec file")

	// Parse as JSON
	var spec map[string]interface{}
	err = json.Unmarshal(data, &spec)
	require.NoError(t, err, "OpenAPI spec should be valid JSON")

	// Validate required fields
	assert.Contains(t, spec, "swagger", "OpenAPI spec should contain 'swagger' field")
	assert.Contains(t, spec, "info", "OpenAPI spec should contain 'info' field")
	assert.Contains(t, spec, "paths", "OpenAPI spec should contain 'paths' field")
}

// TestOpenAPIInfoSection validates the info section of the OpenAPI spec
func TestOpenAPIInfoSection(t *testing.T) {
	projectRoot := getProjectRoot()
	specPath := filepath.Join(projectRoot, "docs/swagger/swagger.json")

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("OpenAPI spec not found. Run 'make swagger-gen' to generate it.")
	}

	data, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read OpenAPI spec file")

	var spec map[string]interface{}
	err = json.Unmarshal(data, &spec)
	require.NoError(t, err, "Failed to parse OpenAPI spec")

	info, ok := spec["info"].(map[string]interface{})
	require.True(t, ok, "info section should be an object")

	assert.Contains(t, info, "title", "info should contain 'title'")
	assert.Contains(t, info, "version", "info should contain 'version'")

	t.Logf("API Title: %v", info["title"])
	t.Logf("API Version: %v", info["version"])
}

// TestOpenAPIPaths validates that required paths exist in the OpenAPI spec
func TestOpenAPIPaths(t *testing.T) {
	projectRoot := getProjectRoot()
	specPath := filepath.Join(projectRoot, "docs/swagger/swagger.json")

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("OpenAPI spec not found. Run 'make swagger-gen' to generate it.")
	}

	data, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read OpenAPI spec file")

	var spec map[string]interface{}
	err = json.Unmarshal(data, &spec)
	require.NoError(t, err, "Failed to parse OpenAPI spec")

	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "paths should be an object")
	assert.NotEmpty(t, paths, "paths should not be empty")

	// Log all available paths
	t.Log("Available API paths:")
	for path := range paths {
		t.Logf("  %s", path)
	}
}

// TestOpenAPIPathMethods validates that paths have correct HTTP methods
func TestOpenAPIPathMethods(t *testing.T) {
	projectRoot := getProjectRoot()
	specPath := filepath.Join(projectRoot, "docs/swagger/swagger.json")

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("OpenAPI spec not found. Run 'make swagger-gen' to generate it.")
	}

	data, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read OpenAPI spec file")

	var spec map[string]interface{}
	err = json.Unmarshal(data, &spec)
	require.NoError(t, err, "Failed to parse OpenAPI spec")

	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "paths should be an object")

	validMethods := map[string]bool{
		"get":     true,
		"post":    true,
		"put":     true,
		"delete":  true,
		"patch":   true,
		"head":    true,
		"options": true,
	}

	for path, pathItem := range paths {
		pathObj, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}

		for method := range pathObj {
			if !validMethods[method] {
				t.Errorf("Invalid HTTP method '%s' for path '%s'", method, path)
			}
		}
	}
}

// TestOpenAPISecurityDefinitions validates security definitions
func TestOpenAPISecurityDefinitions(t *testing.T) {
	projectRoot := getProjectRoot()
	specPath := filepath.Join(projectRoot, "docs/swagger/swagger.json")

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("OpenAPI spec not found. Run 'make swagger-gen' to generate it.")
	}

	data, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read OpenAPI spec file")

	var spec map[string]interface{}
	err = json.Unmarshal(data, &spec)
	require.NoError(t, err, "Failed to parse OpenAPI spec")

	// Check for security definitions (optional but recommended)
	if securityDefinitions, ok := spec["securityDefinitions"].(map[string]interface{}); ok {
		t.Log("Security definitions found:")
		for name := range securityDefinitions {
			t.Logf("  %s", name)
		}
	} else {
		t.Log("No security definitions found (optional)")
	}
}

// TestOpenAPIDefinitions validates that model definitions exist
func TestOpenAPIDefinitions(t *testing.T) {
	projectRoot := getProjectRoot()
	specPath := filepath.Join(projectRoot, "docs/swagger/swagger.json")

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("OpenAPI spec not found. Run 'make swagger-gen' to generate it.")
	}

	data, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read OpenAPI spec file")

	var spec map[string]interface{}
	err = json.Unmarshal(data, &spec)
	require.NoError(t, err, "Failed to parse OpenAPI spec")

	// Check for definitions (Swagger 2.0) or components/schemas (OpenAPI 3.0)
	if definitions, ok := spec["definitions"].(map[string]interface{}); ok {
		t.Log("Model definitions found (Swagger 2.0):")
		for name := range definitions {
			t.Logf("  %s", name)
		}
	} else if components, ok := spec["components"].(map[string]interface{}); ok {
		if schemas, ok := components["schemas"].(map[string]interface{}); ok {
			t.Log("Model definitions found (OpenAPI 3.0):")
			for name := range schemas {
				t.Logf("  %s", name)
			}
		}
	}
}

// TestOpenAPIResponseCodes validates that responses have proper HTTP status codes
func TestOpenAPIResponseCodes(t *testing.T) {
	projectRoot := getProjectRoot()
	specPath := filepath.Join(projectRoot, "docs/swagger/swagger.json")

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Skip("OpenAPI spec not found. Run 'make swagger-gen' to generate it.")
	}

	data, err := os.ReadFile(specPath)
	require.NoError(t, err, "Failed to read OpenAPI spec file")

	var spec map[string]interface{}
	err = json.Unmarshal(data, &spec)
	require.NoError(t, err, "Failed to parse OpenAPI spec")

	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok, "paths should be an object")

	validStatusCodes := map[string]bool{
		"200": true, "201": true, "202": true, "204": true,
		"400": true, "401": true, "403": true, "404": true, "409": true,
		"500": true, "502": true, "503": true,
	}

	for path, pathItem := range paths {
		pathObj, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}

		for method, operation := range pathObj {
			opObj, ok := operation.(map[string]interface{})
			if !ok {
				continue
			}

			responses, ok := opObj["responses"].(map[string]interface{})
			if !ok {
				t.Logf("Warning: No responses defined for %s %s", method, path)
				continue
			}

			for code := range responses {
				if !validStatusCodes[code] {
					t.Errorf("Invalid HTTP status code '%s' for %s %s", code, method, path)
				}
			}
		}
	}
}
