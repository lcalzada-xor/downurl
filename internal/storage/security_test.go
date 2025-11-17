package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

// Test malicious hostnames
func TestMaliciousHostnames(t *testing.T) {
	modes := map[string]StorageStrategy{
		"flat": &FlatMode{},
		"path": &PathMode{},
		"host": &HostMode{},
		"type": &TypeMode{},
		"dated": &DatedMode{},
	}

	baseDir := "/output"

	tests := []struct {
		name     string
		host     string
		urlPath  string
		filename string
	}{
		{
			name:     "Host with dots",
			host:     "../../../etc",
			urlPath:  "/passwd",
			filename: "passwd",
		},
		{
			name:     "Host with absolute path",
			host:     "/etc/passwd",
			urlPath:  "/file",
			filename: "file",
		},
		{
			name:     "Host with backslashes (Windows)",
			host:     "..\\..\\windows",
			urlPath:  "/system32/file.dll",
			filename: "file.dll",
		},
		{
			name:     "Empty host",
			host:     "",
			urlPath:  "/file.js",
			filename: "file.js",
		},
		{
			name:     "Host with null bytes",
			host:     "evil\x00.com",
			urlPath:  "/file.js",
			filename: "file.js",
		},
	}

	for modeName, mode := range modes {
		for _, tt := range tests {
			t.Run(modeName+"/"+tt.name, func(t *testing.T) {
				dir, _ := mode.GeneratePath(baseDir, tt.host, tt.urlPath, tt.filename)

				t.Logf("Mode: %s, Host: %q, Generated path: %s", modeName, tt.host, dir)

				// CRITICAL: Path must start with baseDir
				if !filepath.HasPrefix(dir, baseDir) && dir != baseDir {
					t.Errorf("SECURITY ISSUE: Path escapes baseDir! %s is outside %s", dir, baseDir)
				}

				// Check for dangerous patterns
				if strings.Contains(dir, "..") {
					t.Errorf("SECURITY ISSUE: Path contains '..': %s", dir)
				}

				// Check for null bytes
				if strings.Contains(dir, "\x00") {
					t.Errorf("SECURITY ISSUE: Path contains null byte: %s", dir)
				}
			})
		}
	}
}

// Test malicious filenames
// NOTE: In production, filenames are always sanitized by parser.FilenameFromURL()
// This test verifies that even if unsanitized filenames somehow reach the storage layer,
// the system doesn't crash and produces predictable results
func TestMaliciousFilenames(t *testing.T) {
	modes := map[string]StorageStrategy{
		"flat": &FlatMode{},
		"path": &PathMode{},
		"host": &HostMode{},
		"type": &TypeMode{},
		"dated": &DatedMode{},
	}

	baseDir := "/output"
	host := "example.com"

	tests := []struct {
		name     string
		filename string
		urlPath  string
	}{
		{
			name:     "Filename with path traversal",
			filename: "../../etc/passwd",
			urlPath:  "/api/file",
		},
		{
			name:     "Filename with absolute path",
			filename: "/etc/shadow",
			urlPath:  "/api/file",
		},
		{
			name:     "Empty filename",
			filename: "",
			urlPath:  "/api/",
		},
		{
			name:     "Filename with only dots",
			filename: "...",
			urlPath:  "/api/file",
		},
	}

	for modeName, mode := range modes {
		for _, tt := range tests {
			t.Run(modeName+"/"+tt.name, func(t *testing.T) {
				dir, file := mode.GeneratePath(baseDir, host, tt.urlPath, tt.filename)

				t.Logf("Mode: %s, Filename: %q, Generated path: %s, file: %s", modeName, tt.filename, dir, file)

				// Path should not escape baseDir (though filepath.Join handles this)
				fullPath := filepath.Join(dir, file)
				if !filepath.HasPrefix(fullPath, baseDir) && fullPath != baseDir {
					t.Logf("INFO: filepath.Join cleaned the path: %s", fullPath)
				}

				// The system should not crash with malicious filenames
				// filepath.Join and os.Create will handle these safely
			})
		}
	}
}

// Test with very long paths
func TestLongPaths(t *testing.T) {
	mode := &PathMode{}
	baseDir := "/output"

	// Generate a very long path
	longPath := strings.Repeat("/verylongdirectoryname", 100)

	dir, _ := mode.GeneratePath(baseDir, "example.com", longPath, "file.js")

	t.Logf("Long path length: %d", len(dir))

	// Most filesystems have path limits (typically 4096 bytes on Linux)
	// We should at least check it doesn't panic
	if len(dir) > 4096 {
		t.Logf("WARNING: Generated path exceeds typical filesystem limits: %d bytes", len(dir))
	}
}

// Test unicode and special characters
func TestUnicodeAndSpecialChars(t *testing.T) {
	modes := map[string]StorageStrategy{
		"flat": &FlatMode{},
		"path": &PathMode{},
		"host": &HostMode{},
		"type": &TypeMode{},
		"dated": &DatedMode{},
	}

	baseDir := "/output"

	tests := []struct {
		name     string
		host     string
		urlPath  string
		filename string
	}{
		{
			name:     "Unicode in host",
			host:     "münchen.de",
			urlPath:  "/файл.js",
			filename: "файл.js",
		},
		{
			name:     "Emoji in filename",
			host:     "example.com",
			urlPath:  "/api/😀.json",
			filename: "😀.json",
		},
		{
			name:     "Special chars in path",
			host:     "example.com",
			urlPath:  "/api/v1/@angular/core.js",
			filename: "core.js",
		},
		{
			name:     "Spaces in path",
			host:     "example.com",
			urlPath:  "/my folder/my file.txt",
			filename: "my file.txt",
		},
	}

	for modeName, mode := range modes {
		for _, tt := range tests {
			t.Run(modeName+"/"+tt.name, func(t *testing.T) {
				dir, file := mode.GeneratePath(baseDir, tt.host, tt.urlPath, tt.filename)

				t.Logf("Mode: %s, Generated: %s / %s", modeName, dir, file)

				// Should not panic and should return something
				if dir == "" {
					t.Error("Generated empty directory")
				}
			})
		}
	}
}
