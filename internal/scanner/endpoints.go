package scanner

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// EndpointType represents the type of endpoint
type EndpointType string

const (
	EndpointTypeREST      EndpointType = "rest_api"
	EndpointTypeGraphQL   EndpointType = "graphql"
	EndpointTypeWebSocket EndpointType = "websocket"
	EndpointTypeGeneric   EndpointType = "generic"
)

// HTTPMethod represents HTTP methods
type HTTPMethod string

const (
	MethodGET    HTTPMethod = "GET"
	MethodPOST   HTTPMethod = "POST"
	MethodPUT    HTTPMethod = "PUT"
	MethodDELETE HTTPMethod = "DELETE"
	MethodPATCH  HTTPMethod = "PATCH"
	MethodHEAD   HTTPMethod = "HEAD"
	MethodAny    HTTPMethod = ""
)

// EndpointFinding represents a discovered endpoint
type EndpointFinding struct {
	File       string       `json:"file"`
	URL        string       `json:"url"`
	Endpoint   string       `json:"endpoint"`
	Method     HTTPMethod   `json:"method"`
	Type       EndpointType `json:"type"`
	Line       int          `json:"line"`
	Context    string       `json:"context,omitempty"`
	Parameters []string     `json:"parameters,omitempty"`
}

// EndpointPattern defines a pattern for detecting endpoints
type EndpointPattern struct {
	Name   string
	Regex  *regexp.Regexp
	Method HTTPMethod
	Type   EndpointType
}

// EndpointScanner scans files for API endpoints
type EndpointScanner struct {
	patterns       []EndpointPattern
	includeContext bool
}

// NewEndpointScanner creates a new endpoint scanner
func NewEndpointScanner() *EndpointScanner {
	return &EndpointScanner{
		patterns:       buildEndpointPatterns(),
		includeContext: true,
	}
}

// buildEndpointPatterns creates the list of endpoint patterns
func buildEndpointPatterns() []EndpointPattern {
	return []EndpointPattern{
		// Fetch API
		{
			Name:   "fetch",
			Regex:  regexp.MustCompile(`fetch\s*\(\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodAny,
			Type:   EndpointTypeREST,
		},
		{
			Name:   "fetch with template literal",
			Regex:  regexp.MustCompile(`fetch\s*\(\s*\x60([^\x60]+)\x60`),
			Method: MethodAny,
			Type:   EndpointTypeREST,
		},
		// Axios
		{
			Name:   "axios.get",
			Regex:  regexp.MustCompile(`axios\.get\s*\(\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodGET,
			Type:   EndpointTypeREST,
		},
		{
			Name:   "axios.post",
			Regex:  regexp.MustCompile(`axios\.post\s*\(\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodPOST,
			Type:   EndpointTypeREST,
		},
		{
			Name:   "axios.put",
			Regex:  regexp.MustCompile(`axios\.put\s*\(\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodPUT,
			Type:   EndpointTypeREST,
		},
		{
			Name:   "axios.delete",
			Regex:  regexp.MustCompile(`axios\.delete\s*\(\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodDELETE,
			Type:   EndpointTypeREST,
		},
		{
			Name:   "axios generic",
			Regex:  regexp.MustCompile(`axios\s*\(\s*{[^}]*url\s*:\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodAny,
			Type:   EndpointTypeREST,
		},
		// jQuery AJAX
		{
			Name:   "$.ajax",
			Regex:  regexp.MustCompile(`\$\.ajax\s*\(\s*{[^}]*url\s*:\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodAny,
			Type:   EndpointTypeREST,
		},
		{
			Name:   "$.get",
			Regex:  regexp.MustCompile(`\$\.get\s*\(\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodGET,
			Type:   EndpointTypeREST,
		},
		{
			Name:   "$.post",
			Regex:  regexp.MustCompile(`\$\.post\s*\(\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodPOST,
			Type:   EndpointTypeREST,
		},
		// XMLHttpRequest
		{
			Name:   "xhr.open",
			Regex:  regexp.MustCompile(`\.open\s*\(\s*['"]([A-Z]+)['"]\s*,\s*['"\x60]([^'"\x60]+)['"\x60]`),
			Method: MethodAny,
			Type:   EndpointTypeREST,
		},
		// REST API patterns
		{
			Name:   "api path",
			Regex:  regexp.MustCompile(`['"\x60](/api/[a-zA-Z0-9/_\-{}:]+)['"\x60]`),
			Method: MethodAny,
			Type:   EndpointTypeREST,
		},
		{
			Name:   "v1/v2/v3 api path",
			Regex:  regexp.MustCompile(`['"\x60](/v[0-9]+/[a-zA-Z0-9/_\-{}:]+)['"\x60]`),
			Method: MethodAny,
			Type:   EndpointTypeREST,
		},
		// GraphQL
		{
			Name:   "graphql endpoint",
			Regex:  regexp.MustCompile(`['"\x60]([^'"\x60]*/graphql[^'"\x60]*)['"\x60]`),
			Method: MethodPOST,
			Type:   EndpointTypeGraphQL,
		},
		// WebSocket
		{
			Name:   "websocket",
			Regex:  regexp.MustCompile(`(wss?://[a-zA-Z0-9._\-:/]+)`),
			Method: MethodAny,
			Type:   EndpointTypeWebSocket,
		},
		// Full URLs
		{
			Name:   "https url",
			Regex:  regexp.MustCompile(`['"\x60](https://[a-zA-Z0-9._\-:/]+/[a-zA-Z0-9._\-/{}:]+)['"\x60]`),
			Method: MethodAny,
			Type:   EndpointTypeGeneric,
		},
	}
}

// ScanFile scans a single file for endpoints
func (e *EndpointScanner) ScanFile(filepath, url string) ([]EndpointFinding, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var findings []EndpointFinding
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Track seen endpoints to avoid duplicates
	seen := make(map[string]bool)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check each pattern
		for _, pattern := range e.patterns {
			matches := pattern.Regex.FindAllStringSubmatch(line, -1)

			for _, match := range matches {
				if len(match) < 2 {
					continue
				}

				endpoint := ""
				method := pattern.Method

				// Handle xhr.open special case (has method in capture group)
				if pattern.Name == "xhr.open" && len(match) >= 3 {
					method = HTTPMethod(match[1])
					endpoint = match[2]
				} else {
					endpoint = match[1]
				}

				// Skip if already seen
				key := fmt.Sprintf("%s:%s", method, endpoint)
				if seen[key] {
					continue
				}
				seen[key] = true

				// Extract parameters from endpoint
				params := extractParameters(endpoint)

				finding := EndpointFinding{
					File:       filepath,
					URL:        url,
					Endpoint:   endpoint,
					Method:     method,
					Type:       pattern.Type,
					Line:       lineNum,
					Parameters: params,
				}

				if e.includeContext {
					finding.Context = line
				}

				findings = append(findings, finding)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return findings, nil
}

// extractParameters extracts parameter placeholders from endpoint
func extractParameters(endpoint string) []string {
	var params []string

	// Extract {param} style parameters
	paramRegex := regexp.MustCompile(`\{([^}]+)\}`)
	matches := paramRegex.FindAllStringSubmatch(endpoint, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			params = append(params, match[1])
		}
	}

	// Extract :param style parameters
	colonParamRegex := regexp.MustCompile(`:([a-zA-Z0-9_]+)`)
	matches = colonParamRegex.FindAllStringSubmatch(endpoint, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			params = append(params, match[1])
		}
	}

	return params
}

// ScanBatch scans multiple files
func (e *EndpointScanner) ScanBatch(files map[string]string) ([]EndpointFinding, error) {
	var allFindings []EndpointFinding

	for filepath, url := range files {
		findings, err := e.ScanFile(filepath, url)
		if err != nil {
			// Log error but continue with other files
			continue
		}
		allFindings = append(allFindings, findings...)
	}

	return allFindings, nil
}

// FormatBurpSuite formats endpoints for Burp Suite
func FormatBurpSuite(findings []EndpointFinding, baseURL string) string {
	var lines []string
	seen := make(map[string]bool)

	for _, finding := range findings {
		endpoint := finding.Endpoint

		// Build full URL
		fullURL := endpoint
		if !strings.HasPrefix(endpoint, "http") && !strings.HasPrefix(endpoint, "ws") {
			if strings.HasPrefix(endpoint, "/") {
				fullURL = baseURL + endpoint
			} else {
				fullURL = baseURL + "/" + endpoint
			}
		}

		// Determine method
		method := string(finding.Method)
		if method == "" {
			method = "GET"
		}

		line := fmt.Sprintf("%s %s", method, fullURL)

		// Avoid duplicates
		if !seen[line] {
			seen[line] = true
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// FormatNuclei formats endpoints for Nuclei template
func FormatNuclei(findings []EndpointFinding) string {
	var paths []string
	seen := make(map[string]bool)

	for _, finding := range findings {
		endpoint := finding.Endpoint

		// Only include paths, not full URLs
		if strings.HasPrefix(endpoint, "/") {
			if !seen[endpoint] {
				seen[endpoint] = true
				paths = append(paths, fmt.Sprintf("      - \"{{BaseURL}}%s\"", endpoint))
			}
		}
	}

	if len(paths) == 0 {
		return ""
	}

	template := `id: discovered-endpoints
info:
  name: Discovered Endpoints
  author: downurl
  severity: info

requests:
  - method: GET
    path:
` + strings.Join(paths, "\n")

	return template
}
