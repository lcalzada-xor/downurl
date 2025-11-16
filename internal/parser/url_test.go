package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilenameFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantExt  string // Expected extension
	}{
		{
			name:    "simple js file",
			url:     "https://example.com/script.js",
			wantExt: ".js",
		},
		{
			name:    "js file with query params",
			url:     "https://example.com/script.js?v=123",
			wantExt: "",
		},
		{
			name:    "url without extension",
			url:     "https://example.com/api/data",
			wantExt: "",
		},
		{
			name:    "css file",
			url:     "https://example.com/style.css",
			wantExt: ".css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilenameFromURL(tt.url)
			if got == "" {
				t.Errorf("FilenameFromURL() returned empty string")
			}
			// Check that result contains only safe characters
			for _, r := range got {
				if !isSafeChar(r) {
					t.Errorf("FilenameFromURL() contains unsafe character: %c", r)
				}
			}
		})
	}
}

func TestHostnameFromURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "valid url",
			url:  "https://example.com/path",
			want: "example.com",
		},
		{
			name: "url with port",
			url:  "https://example.com:8080/path",
			want: "example.com:8080",
		},
		{
			name: "invalid url",
			url:  "not a url",
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HostnameFromURL(tt.url)
			if got != tt.want {
				t.Errorf("HostnameFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseURLsFromFile(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "urls.txt")

	content := `https://example.com/file1.js
https://example.com/file2.css

# This is a comment
https://example.com/file3.js
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	urls, err := ParseURLsFromFile(testFile)
	if err != nil {
		t.Fatalf("ParseURLsFromFile() error = %v", err)
	}

	expectedCount := 3
	if len(urls) != expectedCount {
		t.Errorf("ParseURLsFromFile() got %d URLs, want %d", len(urls), expectedCount)
	}

	// Check that comments and empty lines were skipped
	for _, url := range urls {
		if url == "" || url[0] == '#' {
			t.Errorf("ParseURLsFromFile() included invalid URL: %s", url)
		}
	}
}

func TestParseURLsFromFile_NonExistent(t *testing.T) {
	_, err := ParseURLsFromFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ParseURLsFromFile() expected error for non-existent file")
	}
}

func TestParseURLsFromFile_InvalidScheme(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid_scheme.txt")

	content := `https://example.com/valid.js
file:///etc/passwd
ftp://example.com/file.zip
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseURLsFromFile(testFile)
	if err == nil {
		t.Error("ParseURLsFromFile() should reject non-http/https schemes")
	}

	if err != nil && !contains(err.Error(), "invalid URL scheme") {
		t.Errorf("Error should mention invalid scheme, got: %v", err)
	}
}

func TestParseURLsFromFile_MissingHost(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "missing_host.txt")

	content := `https://example.com/valid.js
http://
https:///just/path
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseURLsFromFile(testFile)
	if err == nil {
		t.Error("ParseURLsFromFile() should reject URLs without host")
	}

	if err != nil && !contains(err.Error(), "missing host") {
		t.Errorf("Error should mention missing host, got: %v", err)
	}
}

func TestParseURLsFromFile_ValidURLsOnly(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "valid_urls.txt")

	content := `https://example.com/file1.js
http://example.org/file2.css
https://cdn.example.net/lib.min.js
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	urls, err := ParseURLsFromFile(testFile)
	if err != nil {
		t.Fatalf("ParseURLsFromFile() error = %v, should accept valid URLs", err)
	}

	if len(urls) != 3 {
		t.Errorf("ParseURLsFromFile() got %d URLs, want 3", len(urls))
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		   (s == substr || (len(s) >= len(substr) &&
		   indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Helper function to check if a character is safe for filenames
func isSafeChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '.'
}
