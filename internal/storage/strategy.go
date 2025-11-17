package storage

import (
	"path/filepath"
	"strings"
	"time"
)

// sanitizePathComponent removes dangerous characters and patterns from a path component
// to prevent directory traversal and other path-based attacks
func sanitizePathComponent(component string) string {
	// Remove null bytes
	component = strings.ReplaceAll(component, "\x00", "")

	// Remove leading/trailing whitespace
	component = strings.TrimSpace(component)

	// Clean the path to remove . and .. elements
	component = filepath.Clean(component)

	// Remove any leading slashes or backslashes
	component = strings.TrimPrefix(component, "/")
	component = strings.TrimPrefix(component, "\\")

	// Remove all .. patterns (even in the middle)
	component = strings.ReplaceAll(component, "..", "")

	// Replace backslashes with forward slashes (Windows compatibility)
	component = strings.ReplaceAll(component, "\\", "/")

	// If the component is now just "." or empty, return "unknown"
	if component == "" || component == "." {
		return "unknown"
	}

	return component
}

// StorageStrategy defines how files should be organized in the filesystem
type StorageStrategy interface {
	// GeneratePath creates the full directory and filename path for a file
	// Parameters:
	//   - baseDir: the root output directory
	//   - host: the hostname from the URL
	//   - urlPath: the path component of the URL (e.g., "/api/v1/users")
	//   - filename: the filename to save
	// Returns: the full directory path where the file should be saved
	GeneratePath(baseDir, host, urlPath, filename string) (dir string, finalFilename string)

	// GetDescription returns a human-readable description of this strategy
	GetDescription() string
}

// NewStrategy creates a storage strategy based on the mode name
func NewStrategy(mode string) StorageStrategy {
	switch strings.ToLower(mode) {
	case "path":
		return &PathMode{}
	case "host":
		return &HostMode{}
	case "type":
		return &TypeMode{}
	case "dated":
		return &DatedMode{}
	case "flat":
		fallthrough
	default:
		return &FlatMode{}
	}
}

// FlatMode stores all files in a single directory
type FlatMode struct{}

func (f *FlatMode) GeneratePath(baseDir, host, urlPath, filename string) (string, string) {
	return baseDir, filename
}

func (f *FlatMode) GetDescription() string {
	return "Flat mode: All files in a single directory"
}

// PathMode replicates the URL path structure
type PathMode struct{}

func (p *PathMode) GeneratePath(baseDir, host, urlPath, filename string) (string, string) {
	// Sanitize host to prevent directory traversal
	host = sanitizePathComponent(host)

	// Clean the URL path
	urlPath = strings.TrimPrefix(urlPath, "/")
	urlPath = strings.TrimSuffix(urlPath, "/")

	// Build path: baseDir/host/path/components/
	if urlPath == "" {
		return filepath.Join(baseDir, host), filename
	}

	// Security: Clean the path to prevent directory traversal
	// filepath.Clean removes .. and . elements
	cleanedPath := filepath.Clean(urlPath)

	// Remove leading slashes and dots to prevent escaping baseDir
	cleanedPath = strings.TrimPrefix(cleanedPath, "/")
	cleanedPath = strings.TrimPrefix(cleanedPath, "../")
	for strings.HasPrefix(cleanedPath, "../") {
		cleanedPath = strings.TrimPrefix(cleanedPath, "../")
	}

	// If path becomes empty or just the filename after cleaning, use host only
	if cleanedPath == "" || cleanedPath == "." || cleanedPath == filename {
		return filepath.Join(baseDir, host), filename
	}

	// Remove the filename from the path if it's at the end
	// This handles cases like "/api/v1/data.json" -> we want "/api/v1"
	if strings.HasSuffix(cleanedPath, "/"+filename) {
		cleanedPath = strings.TrimSuffix(cleanedPath, "/"+filename)
	} else if cleanedPath == filename {
		return filepath.Join(baseDir, host), filename
	}

	// Use the cleaned path as the directory structure
	return filepath.Join(baseDir, host, cleanedPath), filename
}

func (p *PathMode) GetDescription() string {
	return "Path mode: Replicates URL directory structure (host/path/to/file)"
}

// HostMode groups files by hostname only
type HostMode struct{}

func (h *HostMode) GeneratePath(baseDir, host, urlPath, filename string) (string, string) {
	// Sanitize host to prevent directory traversal
	host = sanitizePathComponent(host)
	return filepath.Join(baseDir, host), filename
}

func (h *HostMode) GetDescription() string {
	return "Host mode: Groups files by hostname only"
}

// TypeMode organizes files by their extension/type
type TypeMode struct{}

func (t *TypeMode) GeneratePath(baseDir, host, urlPath, filename string) (string, string) {
	// Sanitize host to prevent directory traversal
	host = sanitizePathComponent(host)

	// Get file extension
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = "unknown"
	} else {
		// Remove the dot from extension
		ext = strings.TrimPrefix(ext, ".")
	}

	// Sanitize extension
	ext = sanitizePathComponent(ext)
	if ext == "unknown" || ext == "" {
		ext = "unknown"
	}

	// Create a filename with host prefix to avoid collisions
	prefixedFilename := host + "_" + filename

	return filepath.Join(baseDir, ext), prefixedFilename
}

func (t *TypeMode) GetDescription() string {
	return "Type mode: Organizes files by extension type"
}

// DatedMode organizes files by download date
type DatedMode struct{}

func (d *DatedMode) GeneratePath(baseDir, host, urlPath, filename string) (string, string) {
	// Sanitize host to prevent directory traversal
	host = sanitizePathComponent(host)

	// Get current date in YYYY-MM-DD format
	dateStr := time.Now().Format("2006-01-02")

	// Create a filename with host prefix to avoid collisions
	prefixedFilename := host + "_" + filename

	return filepath.Join(baseDir, dateStr), prefixedFilename
}

func (d *DatedMode) GetDescription() string {
	return "Dated mode: Organizes files by download date (YYYY-MM-DD)"
}
