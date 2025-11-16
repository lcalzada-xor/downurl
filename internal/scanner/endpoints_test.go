package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndpointScanner_FetchAPI(t *testing.T) {
	scanner := NewEndpointScanner()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")

	content := `
fetch('/api/users');
fetch('/api/v1/products/123');
axios.get('/api/posts');
axios.post('/api/comments');
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	findings, err := scanner.ScanFile(testFile, "https://example.com/app.js")
	if err != nil {
		t.Fatalf("ScanFile() error = %v", err)
	}

	if len(findings) < 4 {
		t.Errorf("Expected at least 4 endpoints, got %d", len(findings))
	}

	// Check for specific endpoints
	found := make(map[string]bool)
	for _, finding := range findings {
		found[finding.Endpoint] = true
	}

	expectedEndpoints := []string{
		"/api/users",
		"/api/v1/products/123",
		"/api/posts",
		"/api/comments",
	}

	for _, expected := range expectedEndpoints {
		if !found[expected] {
			t.Errorf("Expected to find endpoint %s", expected)
		}
	}
}

func TestEndpointScanner_Methods(t *testing.T) {
	scanner := NewEndpointScanner()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")

	content := `
axios.get('/api/users');
axios.post('/api/users');
axios.put('/api/users/1');
axios.delete('/api/users/1');
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	findings, err := scanner.ScanFile(testFile, "https://example.com/app.js")
	if err != nil {
		t.Fatalf("ScanFile() error = %v", err)
	}

	// Check methods
	methods := make(map[HTTPMethod]bool)
	for _, finding := range findings {
		methods[finding.Method] = true
	}

	expectedMethods := []HTTPMethod{MethodGET, MethodPOST, MethodPUT, MethodDELETE}
	for _, method := range expectedMethods {
		if !methods[method] {
			t.Errorf("Expected to find method %s", method)
		}
	}
}

func TestEndpointScanner_Parameters(t *testing.T) {
	scanner := NewEndpointScanner()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")

	content := `
fetch('/api/users/{id}');
fetch('/api/posts/:postId/comments/:commentId');
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	findings, err := scanner.ScanFile(testFile, "https://example.com/app.js")
	if err != nil {
		t.Fatalf("ScanFile() error = %v", err)
	}

	for _, finding := range findings {
		if strings.Contains(finding.Endpoint, "{id}") {
			if len(finding.Parameters) == 0 || finding.Parameters[0] != "id" {
				t.Errorf("Expected parameter 'id', got %v", finding.Parameters)
			}
		}
		if strings.Contains(finding.Endpoint, ":postId") {
			found := false
			for _, param := range finding.Parameters {
				if param == "postId" {
					found = true
				}
			}
			if !found {
				t.Errorf("Expected parameter 'postId', got %v", finding.Parameters)
			}
		}
	}
}

func TestFormatBurpSuite(t *testing.T) {
	findings := []EndpointFinding{
		{Endpoint: "/api/users", Method: MethodGET},
		{Endpoint: "/api/users", Method: MethodPOST},
		{Endpoint: "/api/products", Method: MethodGET},
	}

	output := FormatBurpSuite(findings, "https://example.com")

	expected := []string{
		"GET https://example.com/api/users",
		"POST https://example.com/api/users",
		"GET https://example.com/api/products",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain '%s'", exp)
		}
	}
}

func TestFormatNuclei(t *testing.T) {
	findings := []EndpointFinding{
		{Endpoint: "/api/users", Method: MethodGET},
		{Endpoint: "/api/products", Method: MethodGET},
	}

	output := FormatNuclei(findings)

	if !strings.Contains(output, "id: discovered-endpoints") {
		t.Error("Expected Nuclei template to contain id")
	}

	if !strings.Contains(output, "{{BaseURL}}/api/users") {
		t.Error("Expected Nuclei template to contain /api/users endpoint")
	}

	if !strings.Contains(output, "{{BaseURL}}/api/products") {
		t.Error("Expected Nuclei template to contain /api/products endpoint")
	}
}
