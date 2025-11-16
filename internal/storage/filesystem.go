package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// FileStorage handles file system operations
type FileStorage struct {
	baseDir   string
	fileLocks map[string]*sync.Mutex
	mu        sync.Mutex
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(baseDir string) *FileStorage {
	return &FileStorage{
		baseDir:   baseDir,
		fileLocks: make(map[string]*sync.Mutex),
	}
}

// SaveFile saves data to a file in the host's js directory
func (fs *FileStorage) SaveFile(host, filename string, data []byte) (string, error) {
	// Create directory structure: baseDir/host/js/
	dir := filepath.Join(fs.baseDir, host, "js")
	if err := fs.ensureDir(dir); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Full file path
	fullPath := filepath.Join(dir, filename)

	// Get or create lock for this file path
	fs.mu.Lock()
	lock, exists := fs.fileLocks[fullPath]
	if !exists {
		lock = &sync.Mutex{}
		fs.fileLocks[fullPath] = lock
	}
	fs.mu.Unlock()

	// Lock this specific file to prevent race conditions
	lock.Lock()
	defer lock.Unlock()

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		// File exists, create unique name with counter
		return fs.saveFileWithUniqueName(dir, filename, fullPath, data)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, nil
}

// SaveFileFromReader saves data from an io.Reader to a file in the host's js directory
func (fs *FileStorage) SaveFileFromReader(host, filename string, reader io.Reader) (string, int64, error) {
	// Create directory structure: baseDir/host/js/
	dir := filepath.Join(fs.baseDir, host, "js")
	if err := fs.ensureDir(dir); err != nil {
		return "", 0, fmt.Errorf("failed to create directory: %w", err)
	}

	// Full file path
	fullPath := filepath.Join(dir, filename)

	// Get or create lock for this file path
	fs.mu.Lock()
	lock, exists := fs.fileLocks[fullPath]
	if !exists {
		lock = &sync.Mutex{}
		fs.fileLocks[fullPath] = lock
	}
	fs.mu.Unlock()

	// Lock this specific file to prevent race conditions
	lock.Lock()
	defer lock.Unlock()

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		// File exists, create unique name with counter
		return fs.saveFileFromReaderWithUniqueName(dir, filename, fullPath, reader)
	}

	// Create file
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy from reader to file
	bytesWritten, err := io.Copy(file, reader)
	if err != nil {
		return "", bytesWritten, fmt.Errorf("failed to write file: %w", err)
	}

	return fullPath, bytesWritten, nil
}

// saveFileFromReaderWithUniqueName creates a unique filename if collision occurs
func (fs *FileStorage) saveFileFromReaderWithUniqueName(dir, originalName, existingPath string, reader io.Reader) (string, int64, error) {
	// Extract extension
	ext := filepath.Ext(originalName)
	nameWithoutExt := originalName[:len(originalName)-len(ext)]

	// Try up to 1000 variations
	for i := 1; i <= 1000; i++ {
		newName := fmt.Sprintf("%s_%d%s", nameWithoutExt, i, ext)
		newPath := filepath.Join(dir, newName)

		// Check if this variation exists
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			// Create file with new name
			file, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return "", 0, fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()

			// Copy from reader to file
			bytesWritten, err := io.Copy(file, reader)
			if err != nil {
				return "", bytesWritten, fmt.Errorf("failed to write file: %w", err)
			}
			return newPath, bytesWritten, nil
		}
	}

	return "", 0, fmt.Errorf("failed to create unique filename after 1000 attempts")
}

// saveFileWithUniqueName creates a unique filename if collision occurs
func (fs *FileStorage) saveFileWithUniqueName(dir, originalName, existingPath string, data []byte) (string, error) {
	// Extract extension
	ext := filepath.Ext(originalName)
	nameWithoutExt := originalName[:len(originalName)-len(ext)]

	// Try up to 1000 variations
	for i := 1; i <= 1000; i++ {
		newName := fmt.Sprintf("%s_%d%s", nameWithoutExt, i, ext)
		newPath := filepath.Join(dir, newName)

		// Check if this variation exists
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			// Write file with new name
			if err := os.WriteFile(newPath, data, 0644); err != nil {
				return "", fmt.Errorf("failed to write file: %w", err)
			}
			return newPath, nil
		}
	}

	return "", fmt.Errorf("failed to create unique filename after 1000 attempts")
}

// ensureDir creates a directory if it doesn't exist
func (fs *FileStorage) ensureDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}

// GetBaseDir returns the base directory
func (fs *FileStorage) GetBaseDir() string {
	return fs.baseDir
}

// Init ensures the base directory exists
func (fs *FileStorage) Init() error {
	return fs.ensureDir(fs.baseDir)
}
