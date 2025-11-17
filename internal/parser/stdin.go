package parser

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

// ParseURLsFromStdin reads URLs from stdin
func ParseURLsFromStdin() ([]string, error) {
	return parseURLsFromReader(os.Stdin, "stdin")
}

// ParseURLsFromReader reads URLs from any reader
func parseURLsFromReader(reader io.Reader, source string) ([]string, error) {
	var urls []string
	scanner := bufio.NewScanner(reader)
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
		return nil, fmt.Errorf("error reading from %s: %w", source, err)
	}

	return urls, nil
}

// IsStdinAvailable checks if there's data available on stdin
func IsStdinAvailable() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	// Check if stdin is a pipe or file (not a terminal)
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// ParseSingleURL parses and validates a single URL
func ParseSingleURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %s", rawURL)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("invalid URL scheme: %s (only http/https allowed)", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return "", fmt.Errorf("invalid URL (missing host): %s", rawURL)
	}

	return rawURL, nil
}
