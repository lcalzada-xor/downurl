package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStorage_SaveFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	fs := NewFileStorage(tmpDir)

	testData := []byte("test content")
	host := "example.com"
	filename := "test.js"

	path, err := fs.SaveFile(host, filename, testData)
	if err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("SaveFile() did not create file at %s", path)
	}

	// Verify file content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("File content = %s, want %s", content, testData)
	}

	// Verify directory structure
	expectedDir := filepath.Join(tmpDir, host, "js")
	if !dirExists(expectedDir) {
		t.Errorf("Expected directory %s does not exist", expectedDir)
	}
}

func TestFileStorage_Init(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "output")

	fs := NewFileStorage(baseDir)

	if err := fs.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if !dirExists(baseDir) {
		t.Errorf("Init() did not create base directory %s", baseDir)
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
