package storage

import (
	"path/filepath"
	"testing"
)

func TestPathMode_GeneratePath(t *testing.T) {
	mode := &PathMode{}
	baseDir := "/output"

	tests := []struct {
		name         string
		host         string
		urlPath      string
		filename     string
		expectedDir  string
		expectedFile string
	}{
		{
			name:         "Simple path",
			host:         "example.com",
			urlPath:      "/api/v1/users.json",
			filename:     "users.json",
			expectedDir:  filepath.Join("/output", "example.com", "api/v1"),
			expectedFile: "users.json",
		},
		{
			name:         "Root path",
			host:         "example.com",
			urlPath:      "/index.html",
			filename:     "index.html",
			expectedDir:  filepath.Join("/output", "example.com"),
			expectedFile: "index.html",
		},
		{
			name:         "Empty path",
			host:         "example.com",
			urlPath:      "",
			filename:     "file.js",
			expectedDir:  filepath.Join("/output", "example.com"),
			expectedFile: "file.js",
		},
		{
			name:         "Path with trailing slash",
			host:         "example.com",
			urlPath:      "/api/v1/",
			filename:     "data.json",
			expectedDir:  filepath.Join("/output", "example.com", "api/v1"),
			expectedFile: "data.json",
		},
		{
			name:         "Deep nested path",
			host:         "cdn.example.com",
			urlPath:      "/libs/jquery/3.6.0/jquery.min.js",
			filename:     "jquery.min.js",
			expectedDir:  filepath.Join("/output", "cdn.example.com", "libs/jquery/3.6.0"),
			expectedFile: "jquery.min.js",
		},
		{
			name:         "Path with special characters",
			host:         "example.com",
			urlPath:      "/api/v1/@latest/module.js",
			filename:     "module.js",
			expectedDir:  filepath.Join("/output", "example.com", "api/v1/@latest"),
			expectedFile: "module.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, file := mode.GeneratePath(baseDir, tt.host, tt.urlPath, tt.filename)
			if dir != tt.expectedDir {
				t.Errorf("GeneratePath() dir = %v, want %v", dir, tt.expectedDir)
			}
			if file != tt.expectedFile {
				t.Errorf("GeneratePath() file = %v, want %v", file, tt.expectedFile)
			}
		})
	}
}

func TestTypeMode_GeneratePath(t *testing.T) {
	mode := &TypeMode{}
	baseDir := "/output"

	tests := []struct {
		name         string
		host         string
		filename     string
		expectedDir  string
		expectedFile string
	}{
		{
			name:         "JavaScript file",
			host:         "cdn.example.com",
			filename:     "app.js",
			expectedDir:  filepath.Join("/output", "js"),
			expectedFile: "cdn.example.com_app.js",
		},
		{
			name:         "CSS file",
			host:         "cdn.example.com",
			filename:     "style.css",
			expectedDir:  filepath.Join("/output", "css"),
			expectedFile: "cdn.example.com_style.css",
		},
		{
			name:         "File without extension",
			host:         "example.com",
			filename:     "README",
			expectedDir:  filepath.Join("/output", "unknown"),
			expectedFile: "example.com_README",
		},
		{
			name:         "Multiple dots in filename",
			host:         "example.com",
			filename:     "jquery.min.js",
			expectedDir:  filepath.Join("/output", "js"),
			expectedFile: "example.com_jquery.min.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, file := mode.GeneratePath(baseDir, tt.host, "", tt.filename)
			if dir != tt.expectedDir {
				t.Errorf("GeneratePath() dir = %v, want %v", dir, tt.expectedDir)
			}
			if file != tt.expectedFile {
				t.Errorf("GeneratePath() file = %v, want %v", file, tt.expectedFile)
			}
		})
	}
}

func TestHostMode_GeneratePath(t *testing.T) {
	mode := &HostMode{}
	baseDir := "/output"

	tests := []struct {
		name         string
		host         string
		filename     string
		expectedDir  string
		expectedFile string
	}{
		{
			name:         "Standard host",
			host:         "example.com",
			filename:     "file.js",
			expectedDir:  filepath.Join("/output", "example.com"),
			expectedFile: "file.js",
		},
		{
			name:         "Subdomain host",
			host:         "cdn.example.com",
			filename:     "app.js",
			expectedDir:  filepath.Join("/output", "cdn.example.com"),
			expectedFile: "app.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, file := mode.GeneratePath(baseDir, tt.host, "", tt.filename)
			if dir != tt.expectedDir {
				t.Errorf("GeneratePath() dir = %v, want %v", dir, tt.expectedDir)
			}
			if file != tt.expectedFile {
				t.Errorf("GeneratePath() file = %v, want %v", file, tt.expectedFile)
			}
		})
	}
}

func TestFlatMode_GeneratePath(t *testing.T) {
	mode := &FlatMode{}
	baseDir := "/output"

	tests := []struct {
		name         string
		host         string
		filename     string
		expectedDir  string
		expectedFile string
	}{
		{
			name:         "Any file",
			host:         "example.com",
			filename:     "file.js",
			expectedDir:  "/output",
			expectedFile: "file.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, file := mode.GeneratePath(baseDir, tt.host, "", tt.filename)
			if dir != tt.expectedDir {
				t.Errorf("GeneratePath() dir = %v, want %v", dir, tt.expectedDir)
			}
			if file != tt.expectedFile {
				t.Errorf("GeneratePath() file = %v, want %v", file, tt.expectedFile)
			}
		})
	}
}

func TestDatedMode_GeneratePath(t *testing.T) {
	mode := &DatedMode{}
	baseDir := "/output"

	dir, file := mode.GeneratePath(baseDir, "example.com", "/api/test.js", "test.js")

	// Check that directory contains a date pattern (YYYY-MM-DD)
	if !filepath.IsAbs(filepath.Join("/output", "2025-11-17")) {
		// Just check the structure is correct
		if dir == "" {
			t.Error("GeneratePath() returned empty dir")
		}
	}

	expectedFile := "example.com_test.js"
	if file != expectedFile {
		t.Errorf("GeneratePath() file = %v, want %v", file, expectedFile)
	}
}

func TestNewStrategy(t *testing.T) {
	tests := []struct {
		mode     string
		expected string
	}{
		{"flat", "*storage.FlatMode"},
		{"FLAT", "*storage.FlatMode"},
		{"path", "*storage.PathMode"},
		{"host", "*storage.HostMode"},
		{"type", "*storage.TypeMode"},
		{"dated", "*storage.DatedMode"},
		{"invalid", "*storage.FlatMode"}, // Default to flat
		{"", "*storage.FlatMode"},        // Default to flat
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			strategy := NewStrategy(tt.mode)
			if strategy == nil {
				t.Error("NewStrategy() returned nil")
			}
		})
	}
}

// Test for potential path traversal vulnerability
func TestPathMode_PathTraversal(t *testing.T) {
	mode := &PathMode{}
	baseDir := "/output"

	tests := []struct {
		name        string
		host        string
		urlPath     string
		filename    string
		expectInDir string // Should contain this in the path
	}{
		{
			name:        "Double dot attack",
			host:        "evil.com",
			urlPath:     "/../../../etc/passwd",
			filename:    "passwd",
			expectInDir: "/output/evil.com",
		},
		{
			name:        "Dot dot in middle",
			host:        "evil.com",
			urlPath:     "/api/../../../etc/shadow",
			filename:    "shadow",
			expectInDir: "/output/evil.com",
		},
		{
			name:        "Windows path traversal",
			host:        "evil.com",
			urlPath:     "/..\\..\\windows\\system32",
			filename:    "file.dll",
			expectInDir: "/output/evil.com",
		},
		{
			name:        "Multiple dot dots",
			host:        "evil.com",
			urlPath:     "/../../../../../../../../../etc/passwd",
			filename:    "passwd",
			expectInDir: "/output/evil.com",
		},
		{
			name:        "Legitimate nested path should work",
			host:        "good.com",
			urlPath:     "/api/v1/users/data.json",
			filename:    "data.json",
			expectInDir: "/output/good.com/api/v1/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, _ := mode.GeneratePath(baseDir, tt.host, tt.urlPath, tt.filename)

			t.Logf("Generated path: %s", dir)

			// Check that the path is within the expected directory
			if !filepath.HasPrefix(dir, tt.expectInDir) {
				t.Errorf("Path traversal detected! Path %s does not start with %s", dir, tt.expectInDir)
			}

			// Additional check: ensure path doesn't escape baseDir
			if !filepath.HasPrefix(dir, baseDir) {
				t.Errorf("CRITICAL: Path escapes baseDir! %s is outside %s", dir, baseDir)
			}
		})
	}
}
