package scanner

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
)

// SecretType represents the type of secret found
type SecretType string

const (
	SecretTypeAWSKey       SecretType = "AWS Access Key"
	SecretTypeAWSSecret    SecretType = "AWS Secret Key"
	SecretTypeGitHubToken  SecretType = "GitHub Token"
	SecretTypeSlackToken   SecretType = "Slack Token"
	SecretTypeGoogleAPIKey SecretType = "Google API Key"
	SecretTypeJWT          SecretType = "JWT Token"
	SecretTypePrivateKey   SecretType = "Private Key"
	SecretTypeGenericAPI   SecretType = "Generic API Key"
	SecretTypePassword     SecretType = "Password in Code"
	SecretTypeDatabaseURL  SecretType = "Database URL"
	SecretTypeGenericHigh  SecretType = "High Entropy String"
)

// Confidence level for secret detection
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

// SecretPattern defines a pattern for detecting secrets
type SecretPattern struct {
	Name       SecretType
	Regex      *regexp.Regexp
	Confidence Confidence
	MinLength  int
	MaxLength  int
}

// SecretFinding represents a found secret
type SecretFinding struct {
	File       string     `json:"file"`
	URL        string     `json:"url"`
	Line       int        `json:"line"`
	SecretType SecretType `json:"secret_type"`
	Match      string     `json:"match"`
	Context    string     `json:"context"`
	Confidence Confidence `json:"confidence"`
}

// SecretScanner scans files for secrets
type SecretScanner struct {
	patterns        []SecretPattern
	minEntropy      float64
	entropyMinLen   int
	includeContext  bool
	contextLines    int
}

// NewSecretScanner creates a new secret scanner
func NewSecretScanner(minEntropy float64) *SecretScanner {
	return &SecretScanner{
		patterns:       buildPatterns(),
		minEntropy:     minEntropy,
		entropyMinLen:  20,
		includeContext: true,
		contextLines:   2,
	}
}

// buildPatterns creates the list of secret patterns
func buildPatterns() []SecretPattern {
	return []SecretPattern{
		{
			Name:       SecretTypeAWSKey,
			Regex:      regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
			Confidence: ConfidenceHigh,
		},
		{
			Name:       SecretTypeAWSSecret,
			Regex:      regexp.MustCompile(`(?i)aws[_-]?secret[_-]?access[_-]?key['"\s]*[:=]\s*['"]([a-zA-Z0-9/+=]{40})['"]`),
			Confidence: ConfidenceHigh,
		},
		{
			Name:       SecretTypeGitHubToken,
			Regex:      regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
			Confidence: ConfidenceHigh,
		},
		{
			Name:       SecretTypeGitHubToken,
			Regex:      regexp.MustCompile(`gho_[a-zA-Z0-9]{36}`),
			Confidence: ConfidenceHigh,
		},
		{
			Name:       SecretTypeSlackToken,
			Regex:      regexp.MustCompile(`xox[baprs]-[0-9a-zA-Z-]{10,48}`),
			Confidence: ConfidenceHigh,
		},
		{
			Name:       SecretTypeGoogleAPIKey,
			Regex:      regexp.MustCompile(`AIza[0-9A-Za-z_\-]{35}`),
			Confidence: ConfidenceHigh,
		},
		{
			Name:       SecretTypeJWT,
			Regex:      regexp.MustCompile(`eyJ[a-zA-Z0-9_\-]*\.eyJ[a-zA-Z0-9_\-]*\.[a-zA-Z0-9_\-]*`),
			Confidence: ConfidenceMedium,
		},
		{
			Name:       SecretTypePrivateKey,
			Regex:      regexp.MustCompile(`-----BEGIN (RSA|DSA|EC|OPENSSH|PGP) PRIVATE KEY-----`),
			Confidence: ConfidenceHigh,
		},
		{
			Name:       SecretTypeDatabaseURL,
			Regex:      regexp.MustCompile(`(?i)(mongodb|postgres|mysql|redis)://[^\s'"]+`),
			Confidence: ConfidenceMedium,
		},
		{
			Name:       SecretTypePassword,
			Regex:      regexp.MustCompile(`(?i)password\s*[:=]\s*['"]([^'"]{8,})['"]`),
			Confidence: ConfidenceLow,
		},
		{
			Name:       SecretTypeGenericAPI,
			Regex:      regexp.MustCompile(`(?i)api[_-]?key\s*[:=]\s*['"]([a-zA-Z0-9_\-]{16,})['"]`),
			Confidence: ConfidenceMedium,
		},
	}
}

// ScanFile scans a single file for secrets
func (s *SecretScanner) ScanFile(filepath, url string) ([]SecretFinding, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var findings []SecretFinding
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var lines []string

	// Read all lines for context
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Scan each line
	for i, line := range lines {
		lineNum = i + 1

		// Check pattern-based secrets
		for _, pattern := range s.patterns {
			matches := pattern.Regex.FindAllString(line, -1)
			for _, match := range matches {
				finding := SecretFinding{
					File:       filepath,
					URL:        url,
					Line:       lineNum,
					SecretType: pattern.Name,
					Match:      match,
					Confidence: pattern.Confidence,
				}

				if s.includeContext {
					finding.Context = s.getContext(lines, i, s.contextLines)
				}

				findings = append(findings, finding)
			}
		}

		// Check entropy-based detection
		highEntropyStrings := s.findHighEntropyStrings(line)
		for _, str := range highEntropyStrings {
			finding := SecretFinding{
				File:       filepath,
				URL:        url,
				Line:       lineNum,
				SecretType: SecretTypeGenericHigh,
				Match:      str,
				Confidence: ConfidenceLow,
			}

			if s.includeContext {
				finding.Context = s.getContext(lines, i, s.contextLines)
			}

			findings = append(findings, finding)
		}
	}

	return findings, nil
}

// getContext returns context lines around the match
func (s *SecretScanner) getContext(lines []string, index, contextLines int) string {
	start := index - contextLines
	if start < 0 {
		start = 0
	}

	end := index + contextLines + 1
	if end > len(lines) {
		end = len(lines)
	}

	contextSlice := lines[start:end]
	return strings.Join(contextSlice, "\n")
}

// findHighEntropyStrings finds strings with high Shannon entropy
func (s *SecretScanner) findHighEntropyStrings(line string) []string {
	var highEntropyStrings []string

	// Extract potential string literals
	stringRegex := regexp.MustCompile(`['"]([a-zA-Z0-9+/=_\-]{20,})['"]`)
	matches := stringRegex.FindAllStringSubmatch(line, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		str := match[1]
		if len(str) < s.entropyMinLen {
			continue
		}

		entropy := s.calculateEntropy(str)
		if entropy >= s.minEntropy {
			highEntropyStrings = append(highEntropyStrings, str)
		}
	}

	return highEntropyStrings
}

// calculateEntropy calculates Shannon entropy of a string
func (s *SecretScanner) calculateEntropy(str string) float64 {
	if len(str) == 0 {
		return 0.0
	}

	// Count character frequencies
	freq := make(map[rune]int)
	for _, char := range str {
		freq[char]++
	}

	// Calculate entropy
	var entropy float64
	length := float64(len(str))

	for _, count := range freq {
		probability := float64(count) / length
		if probability > 0 {
			entropy -= probability * math.Log2(probability)
		}
	}

	return entropy
}

// ScanBatch scans multiple files
func (s *SecretScanner) ScanBatch(files map[string]string) ([]SecretFinding, error) {
	var allFindings []SecretFinding

	for filepath, url := range files {
		findings, err := s.ScanFile(filepath, url)
		if err != nil {
			// Log error but continue with other files
			continue
		}
		allFindings = append(allFindings, findings...)
	}

	return allFindings, nil
}

// FilterByConfidence filters findings by confidence level
func FilterByConfidence(findings []SecretFinding, minConfidence Confidence) []SecretFinding {
	confidenceOrder := map[Confidence]int{
		ConfidenceHigh:   3,
		ConfidenceMedium: 2,
		ConfidenceLow:    1,
	}

	minLevel := confidenceOrder[minConfidence]
	var filtered []SecretFinding

	for _, finding := range findings {
		if confidenceOrder[finding.Confidence] >= minLevel {
			filtered = append(filtered, finding)
		}
	}

	return filtered
}
