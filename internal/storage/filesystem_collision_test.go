package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestFileStorage_SaveFile_Collision(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	fs := NewFileStorage(tmpDir, "flat")

	testData1 := []byte("test content 1")
	testData2 := []byte("test content 2 - different")
	host := "example.com"
	urlPath := "/api/test.js"
	filename := "test.js"

	// Save first file
	path1, err := fs.SaveFile(host, urlPath, filename, testData1)
	if err != nil {
		t.Fatalf("SaveFile() first call error = %v", err)
	}

	// Save second file with same name (should create unique name)
	path2, err := fs.SaveFile(host, urlPath, filename, testData2)
	if err != nil {
		t.Fatalf("SaveFile() second call error = %v", err)
	}

	// Paths should be different
	if path1 == path2 {
		t.Errorf("SaveFile() did not create unique filename, both paths are %s", path1)
	}

	// Verify both files exist
	if _, err := os.Stat(path1); os.IsNotExist(err) {
		t.Errorf("First file does not exist at %s", path1)
	}

	if _, err := os.Stat(path2); os.IsNotExist(err) {
		t.Errorf("Second file does not exist at %s", path2)
	}

	// Verify content is different
	content1, _ := os.ReadFile(path1)
	content2, _ := os.ReadFile(path2)

	if string(content1) == string(content2) {
		t.Error("File contents should be different")
	}

	// Verify second file has counter suffix (in flat mode, files go to base dir)
	expectedPath2 := filepath.Join(tmpDir, "test_1.js")
	if path2 != expectedPath2 {
		t.Errorf("Second file path = %s, want %s", path2, expectedPath2)
	}
}

func TestFileStorage_SaveFile_ConcurrentWrites(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	fs := NewFileStorage(tmpDir, "flat")

	host := "example.com"
	urlPath := "/concurrent.js"
	filename := "concurrent.js"
	numGoroutines := 10

	var wg sync.WaitGroup
	paths := make([]string, numGoroutines)

	// Write same filename concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			data := []byte("content from goroutine " + string(rune(index)))
			path, err := fs.SaveFile(host, urlPath, filename, data)
			if err != nil {
				t.Errorf("SaveFile() concurrent error = %v", err)
				return
			}
			paths[index] = path
		}(i)
	}

	wg.Wait()

	// Check that all paths are unique
	uniquePaths := make(map[string]bool)
	for _, path := range paths {
		if path == "" {
			t.Error("Got empty path from concurrent write")
			continue
		}
		if uniquePaths[path] {
			t.Errorf("Duplicate path detected: %s", path)
		}
		uniquePaths[path] = true
	}

	// Should have numGoroutines unique files
	if len(uniquePaths) != numGoroutines {
		t.Errorf("Expected %d unique files, got %d", numGoroutines, len(uniquePaths))
	}
}

func TestFileStorage_SaveFile_NoExtension(t *testing.T) {
	tmpDir := t.TempDir()
	fs := NewFileStorage(tmpDir, "flat")

	testData1 := []byte("content 1")
	testData2 := []byte("content 2")
	host := "example.com"
	urlPath := "/noextension"
	filename := "noextension"

	// Save first file
	path1, err := fs.SaveFile(host, urlPath, filename, testData1)
	if err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	// Save second file with same name
	path2, err := fs.SaveFile(host, urlPath, filename, testData2)
	if err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	// Verify unique names were created
	if path1 == path2 {
		t.Error("Files with no extension should still get unique names")
	}

	// Second file should have _1 suffix (in flat mode)
	expectedPath2 := filepath.Join(tmpDir, "noextension_1")
	if path2 != expectedPath2 {
		t.Errorf("Second file path = %s, want %s", path2, expectedPath2)
	}
}
