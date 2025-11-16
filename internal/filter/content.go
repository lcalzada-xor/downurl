package filter

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// ContentFilter filters downloads based on content type and size
type ContentFilter struct {
	AllowedTypes      []string
	BlockedTypes      []string
	AllowedExtensions []string
	BlockedExtensions []string
	MinSize           int64
	MaxSize           int64
	SkipEmpty         bool
}

// FilterConfig represents filter configuration
type FilterConfig struct {
	FilterType     string // Comma-separated list of allowed types
	ExcludeType    string // Comma-separated list of blocked types
	FilterExt      string // Comma-separated list of allowed extensions
	ExcludeExt     string // Comma-separated list of blocked extensions
	MinSize        int64  // Minimum file size in bytes
	MaxSize        int64  // Maximum file size in bytes
	SkipEmpty      bool   // Skip empty files
}

// NewContentFilter creates a new content filter
func NewContentFilter(cfg FilterConfig) *ContentFilter {
	filter := &ContentFilter{
		MinSize:   cfg.MinSize,
		MaxSize:   cfg.MaxSize,
		SkipEmpty: cfg.SkipEmpty,
	}

	// Parse allowed types
	if cfg.FilterType != "" {
		filter.AllowedTypes = parseList(cfg.FilterType)
	}

	// Parse blocked types
	if cfg.ExcludeType != "" {
		filter.BlockedTypes = parseList(cfg.ExcludeType)
	}

	// Parse allowed extensions
	if cfg.FilterExt != "" {
		filter.AllowedExtensions = parseList(cfg.FilterExt)
		// Ensure extensions start with dot
		for i, ext := range filter.AllowedExtensions {
			if !strings.HasPrefix(ext, ".") {
				filter.AllowedExtensions[i] = "." + ext
			}
		}
	}

	// Parse blocked extensions
	if cfg.ExcludeExt != "" {
		filter.BlockedExtensions = parseList(cfg.ExcludeExt)
		// Ensure extensions start with dot
		for i, ext := range filter.BlockedExtensions {
			if !strings.HasPrefix(ext, ".") {
				filter.BlockedExtensions[i] = "." + ext
			}
		}
	}

	return filter
}

// parseList parses comma-separated list
func parseList(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// ShouldDownload determines if a file should be downloaded based on filters
func (f *ContentFilter) ShouldDownload(url string, contentType string, contentLength int64) (bool, string) {
	// Check content length
	if contentLength >= 0 {
		// Check if empty
		if f.SkipEmpty && contentLength == 0 {
			return false, "file is empty"
		}

		// Check minimum size
		if f.MinSize > 0 && contentLength < f.MinSize {
			return false, fmt.Sprintf("file too small (%d bytes, min: %d)", contentLength, f.MinSize)
		}

		// Check maximum size
		if f.MaxSize > 0 && contentLength > f.MaxSize {
			return false, fmt.Sprintf("file too large (%d bytes, max: %d)", contentLength, f.MaxSize)
		}
	}

	// Extract extension from URL
	ext := filepath.Ext(url)
	if ext != "" {
		// Remove query parameters
		if idx := strings.Index(ext, "?"); idx != -1 {
			ext = ext[:idx]
		}
		ext = strings.ToLower(ext)
	}

	// Check blocked extensions first
	if len(f.BlockedExtensions) > 0 {
		for _, blockedExt := range f.BlockedExtensions {
			if ext == strings.ToLower(blockedExt) {
				return false, fmt.Sprintf("extension blocked: %s", ext)
			}
		}
	}

	// Check allowed extensions
	if len(f.AllowedExtensions) > 0 {
		allowed := false
		for _, allowedExt := range f.AllowedExtensions {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false, fmt.Sprintf("extension not in allowed list: %s", ext)
		}
	}

	// Parse content type
	if contentType != "" {
		contentType = strings.ToLower(contentType)
		// Remove charset and other parameters
		if idx := strings.Index(contentType, ";"); idx != -1 {
			contentType = strings.TrimSpace(contentType[:idx])
		}

		// Check blocked types first
		if len(f.BlockedTypes) > 0 {
			for _, blockedType := range f.BlockedTypes {
				if f.matchContentType(contentType, strings.ToLower(blockedType)) {
					return false, fmt.Sprintf("content-type blocked: %s", contentType)
				}
			}
		}

		// Check allowed types
		if len(f.AllowedTypes) > 0 {
			allowed := false
			for _, allowedType := range f.AllowedTypes {
				if f.matchContentType(contentType, strings.ToLower(allowedType)) {
					allowed = true
					break
				}
			}
			if !allowed {
				return false, fmt.Sprintf("content-type not in allowed list: %s", contentType)
			}
		}
	}

	return true, ""
}

// matchContentType checks if contentType matches pattern (supports wildcards)
func (f *ContentFilter) matchContentType(contentType, pattern string) bool {
	// Exact match
	if contentType == pattern {
		return true
	}

	// Wildcard match (e.g., "image/*")
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(contentType, prefix+"/")
	}

	// Wildcard match (e.g., "*/json")
	if strings.HasPrefix(pattern, "*/") {
		suffix := strings.TrimPrefix(pattern, "*/")
		return strings.HasSuffix(contentType, "/"+suffix)
	}

	return false
}

// ShouldProcess determines if downloaded content should be processed
func (f *ContentFilter) ShouldProcess(data []byte, contentType string) (bool, string) {
	// Check if empty
	if f.SkipEmpty && len(data) == 0 {
		return false, "content is empty"
	}

	// Check minimum size
	if f.MinSize > 0 && int64(len(data)) < f.MinSize {
		return false, fmt.Sprintf("content too small (%d bytes, min: %d)", len(data), f.MinSize)
	}

	// Check maximum size
	if f.MaxSize > 0 && int64(len(data)) > f.MaxSize {
		return false, fmt.Sprintf("content too large (%d bytes, max: %d)", len(data), f.MaxSize)
	}

	return true, ""
}

// DetectContentType detects content type from data and extension
func DetectContentType(data []byte, filename string) string {
	// Try HTTP content type detection
	contentType := http.DetectContentType(data)

	// If generic, try extension-based detection
	if contentType == "application/octet-stream" || contentType == "text/plain; charset=utf-8" {
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".js", ".mjs":
			return "text/javascript"
		case ".json":
			return "application/json"
		case ".css":
			return "text/css"
		case ".html", ".htm":
			return "text/html"
		case ".xml":
			return "application/xml"
		case ".yaml", ".yml":
			return "application/yaml"
		case ".txt":
			return "text/plain"
		}
	}

	return contentType
}

// ClassifyContent classifies content type into categories
func ClassifyContent(contentType string) string {
	contentType = strings.ToLower(contentType)

	// Remove parameters
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	switch {
	case strings.HasPrefix(contentType, "text/javascript") ||
	     strings.HasPrefix(contentType, "application/javascript") ||
	     strings.HasPrefix(contentType, "application/x-javascript"):
		return "JavaScript"
	case strings.HasPrefix(contentType, "application/json"):
		return "JSON"
	case strings.HasPrefix(contentType, "text/html"):
		return "HTML"
	case strings.HasPrefix(contentType, "text/css"):
		return "CSS"
	case strings.HasPrefix(contentType, "application/xml") ||
	     strings.HasPrefix(contentType, "text/xml"):
		return "XML"
	case strings.HasPrefix(contentType, "text/plain"):
		return "Text"
	case strings.HasPrefix(contentType, "image/"):
		return "Image"
	case strings.HasPrefix(contentType, "video/"):
		return "Video"
	case strings.HasPrefix(contentType, "audio/"):
		return "Audio"
	case strings.HasPrefix(contentType, "application/pdf"):
		return "PDF"
	case strings.HasPrefix(contentType, "application/zip") ||
	     strings.HasPrefix(contentType, "application/x-gzip"):
		return "Archive"
	default:
		return "Other"
	}
}

// IsText checks if content type is text-based
func IsText(contentType string) bool {
	contentType = strings.ToLower(contentType)
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	textTypes := []string{
		"text/",
		"application/javascript",
		"application/json",
		"application/xml",
		"application/x-javascript",
	}

	for _, prefix := range textTypes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}

	return false
}

// IsJavaScript checks if content is JavaScript
func IsJavaScript(contentType string) bool {
	contentType = strings.ToLower(contentType)
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	jsTypes := []string{
		"text/javascript",
		"application/javascript",
		"application/x-javascript",
	}

	for _, jsType := range jsTypes {
		if contentType == jsType {
			return true
		}
	}

	return false
}
