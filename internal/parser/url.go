package parser

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"unicode"
)

// ParseURLsFromFile reads URLs from a file and returns them as a slice
func ParseURLsFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Validate URL
		parsedURL, err := url.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("invalid URL at line %d: %s", lineNum, line)
		}

		// Validate URL scheme (only http and https allowed)
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			return nil, fmt.Errorf("invalid URL scheme at line %d: %s (only http/https allowed)", lineNum, parsedURL.Scheme)
		}

		// Validate hostname exists
		if parsedURL.Host == "" {
			return nil, fmt.Errorf("invalid URL (missing host) at line %d: %s", lineNum, line)
		}

		urls = append(urls, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return urls, nil
}

// FilenameFromURL generates a safe filename from a URL
func FilenameFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// Fallback to hash if URL is invalid
		return hashFilename(rawURL, "")
	}

	// Extract filename from path
	name := path.Base(parsed.Path)
	name = strings.TrimSuffix(name, "/")

	// If no valid name or no extension, generate hash-based name
	if name == "" || name == "." || name == "/" || !strings.Contains(name, ".") {
		ext := detectExtension(rawURL)
		return hashFilename(rawURL, ext)
	}

	// Sanitize filename
	return sanitizeFilename(name)
}

// HostnameFromURL extracts the hostname from a URL
func HostnameFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "unknown"
	}
	if parsed.Host == "" {
		return "unknown"
	}
	return parsed.Host
}

// PathFromURL extracts the path component from a URL
func PathFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Path
}

// sanitizeFilename replaces unsafe characters with underscores
func sanitizeFilename(name string) string {
	var result strings.Builder
	result.Grow(len(name))

	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}

	return result.String()
}

// hashFilename generates a filename based on URL hash
func hashFilename(rawURL, ext string) string {
	hash := sha1.Sum([]byte(rawURL))
	hashStr := fmt.Sprintf("%x", hash)[:10]

	if ext == "" {
		ext = detectExtension(rawURL)
	}

	return hashStr + ext
}

// detectExtension tries to detect file extension from URL
func detectExtension(rawURL string) string {
	if strings.Contains(rawURL, ".js") || strings.HasSuffix(rawURL, ".js") || strings.HasSuffix(rawURL, ".mjs") {
		return ".js"
	}
	if strings.Contains(rawURL, ".css") || strings.HasSuffix(rawURL, ".css") {
		return ".css"
	}
	if strings.Contains(rawURL, ".json") || strings.HasSuffix(rawURL, ".json") {
		return ".json"
	}
	return ""
}
