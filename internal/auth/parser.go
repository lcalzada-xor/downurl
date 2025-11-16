package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseHeadersFile parses a headers file and returns a map of headers
// Format: "Header-Name: value"
func ParseHeadersFile(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open headers file: %w", err)
	}
	defer file.Close()

	headers := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse header (format: "Name: value")
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format at line %d: %s (expected 'Name: value')", lineNum, line)
		}

		headerName := strings.TrimSpace(parts[0])
		headerValue := strings.TrimSpace(parts[1])

		if headerName == "" {
			return nil, fmt.Errorf("empty header name at line %d", lineNum)
		}

		headers[headerName] = headerValue
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading headers file: %w", err)
	}

	if len(headers) == 0 {
		return nil, fmt.Errorf("no valid headers found in file")
	}

	return headers, nil
}

// ParseCookiesFile parses a cookies file and returns a map of cookies
// Format: "name=value"
func ParseCookiesFile(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cookies file: %w", err)
	}
	defer file.Close()

	cookies := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse cookie (format: "name=value")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid cookie format at line %d: %s (expected 'name=value')", lineNum, line)
		}

		cookieName := strings.TrimSpace(parts[0])
		cookieValue := strings.TrimSpace(parts[1])

		if cookieName == "" {
			return nil, fmt.Errorf("empty cookie name at line %d", lineNum)
		}

		cookies[cookieName] = cookieValue
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading cookies file: %w", err)
	}

	if len(cookies) == 0 {
		return nil, fmt.Errorf("no valid cookies found in file")
	}

	return cookies, nil
}

// ParseCookieString parses a cookie string (format: "name1=value1; name2=value2")
func ParseCookieString(cookieStr string) map[string]string {
	cookies := make(map[string]string)

	// Split by semicolon
	pairs := strings.Split(cookieStr, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Split by equals sign
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if name != "" {
				cookies[name] = value
			}
		}
	}

	return cookies
}

// ParseBasicAuth parses a basic auth string (format: "username:password")
func ParseBasicAuth(authStr string) (username, password string, err error) {
	parts := strings.SplitN(authStr, ":", 2)
	if len(parts) < 1 || parts[0] == "" {
		return "", "", fmt.Errorf("invalid basic auth format (expected 'username:password')")
	}

	username = parts[0]
	if len(parts) == 2 {
		password = parts[1]
	}

	return username, password, nil
}
