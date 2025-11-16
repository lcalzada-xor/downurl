package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretScanner_AWS_Keys(t *testing.T) {
	scanner := NewSecretScanner(4.5)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")

	content := `
const config = {
	awsKey: 'AKIAIOSFODNN7EXAMPLE',
	awsSecret: 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'
};
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	findings, err := scanner.ScanFile(testFile, "https://example.com/test.js")
	if err != nil {
		t.Fatalf("ScanFile() error = %v", err)
	}

	if len(findings) == 0 {
		t.Error("Expected to find AWS key, got no findings")
	}

	foundAWSKey := false
	for _, finding := range findings {
		if finding.SecretType == SecretTypeAWSKey {
			foundAWSKey = true
			if !strings.Contains(finding.Match, "AKIAIOSFODNN7EXAMPLE") {
				t.Errorf("Expected match to contain AKIAIOSFODNN7EXAMPLE, got %s", finding.Match)
			}
		}
	}

	if !foundAWSKey {
		t.Error("Expected to find AWS Access Key")
	}
}

func TestSecretScanner_JWT(t *testing.T) {
	scanner := NewSecretScanner(4.5)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")

	content := `
const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	findings, err := scanner.ScanFile(testFile, "https://example.com/test.js")
	if err != nil {
		t.Fatalf("ScanFile() error = %v", err)
	}

	foundJWT := false
	for _, finding := range findings {
		if finding.SecretType == SecretTypeJWT {
			foundJWT = true
		}
	}

	if !foundJWT {
		t.Error("Expected to find JWT token")
	}
}

func TestSecretScanner_Entropy(t *testing.T) {
	scanner := NewSecretScanner(4.5)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.js")

	content := `
const apiKey = 'aB3dEf5gH7iJ9kL1mN3oP5qR7sT9uV1wX3yZ5';
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	findings, err := scanner.ScanFile(testFile, "https://example.com/test.js")
	if err != nil {
		t.Fatalf("ScanFile() error = %v", err)
	}

	// Should find high entropy string
	foundHighEntropy := false
	for _, finding := range findings {
		if finding.SecretType == SecretTypeGenericHigh {
			foundHighEntropy = true
		}
	}

	if !foundHighEntropy {
		t.Error("Expected to find high entropy string")
	}
}

func TestSecretScanner_CalculateEntropy(t *testing.T) {
	scanner := NewSecretScanner(4.5)

	tests := []struct {
		name     string
		input    string
		minEntropy float64
	}{
		{
			name:       "low entropy",
			input:      "aaaaaaaaaa",
			minEntropy: 0,
		},
		{
			name:       "high entropy",
			input:      "aB3dEf5gH7iJ9kL1mN",
			minEntropy: 4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entropy := scanner.calculateEntropy(tt.input)
			if entropy < tt.minEntropy {
				t.Errorf("Entropy = %.2f, want >= %.2f", entropy, tt.minEntropy)
			}
		})
	}
}

func TestFilterByConfidence(t *testing.T) {
	findings := []SecretFinding{
		{SecretType: SecretTypeAWSKey, Confidence: ConfidenceHigh},
		{SecretType: SecretTypeJWT, Confidence: ConfidenceMedium},
		{SecretType: SecretTypeGenericHigh, Confidence: ConfidenceLow},
	}

	high := FilterByConfidence(findings, ConfidenceHigh)
	if len(high) != 1 {
		t.Errorf("FilterByConfidence(High) = %d findings, want 1", len(high))
	}

	medium := FilterByConfidence(findings, ConfidenceMedium)
	if len(medium) != 2 {
		t.Errorf("FilterByConfidence(Medium) = %d findings, want 2", len(medium))
	}

	low := FilterByConfidence(findings, ConfidenceLow)
	if len(low) != 3 {
		t.Errorf("FilterByConfidence(Low) = %d findings, want 3", len(low))
	}
}
