package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseHeadersFile(t *testing.T) {
	tmpDir := t.TempDir()
	headersFile := filepath.Join(tmpDir, "headers.txt")

	content := `Authorization: Bearer token123
X-API-Key: secret456
X-Custom-Header: custom-value

# This is a comment
User-Agent: CustomBot/1.0
`

	if err := os.WriteFile(headersFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	headers, err := ParseHeadersFile(headersFile)
	if err != nil {
		t.Fatalf("ParseHeadersFile() error = %v", err)
	}

	expectedCount := 4
	if len(headers) != expectedCount {
		t.Errorf("ParseHeadersFile() got %d headers, want %d", len(headers), expectedCount)
	}

	if headers["Authorization"] != "Bearer token123" {
		t.Errorf("Authorization header = %v, want 'Bearer token123'", headers["Authorization"])
	}

	if headers["X-API-Key"] != "secret456" {
		t.Errorf("X-API-Key header = %v, want 'secret456'", headers["X-API-Key"])
	}

	if headers["User-Agent"] != "CustomBot/1.0" {
		t.Errorf("User-Agent header = %v, want 'CustomBot/1.0'", headers["User-Agent"])
	}
}

func TestParseHeadersFile_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	headersFile := filepath.Join(tmpDir, "invalid.txt")

	content := `Valid-Header: value
InvalidHeaderWithoutColon
Another-Valid: value2
`

	if err := os.WriteFile(headersFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseHeadersFile(headersFile)
	if err == nil {
		t.Error("ParseHeadersFile() expected error for invalid format")
	}
}

func TestParseHeadersFile_NonExistent(t *testing.T) {
	_, err := ParseHeadersFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ParseHeadersFile() expected error for non-existent file")
	}
}

func TestParseCookiesFile(t *testing.T) {
	tmpDir := t.TempDir()
	cookiesFile := filepath.Join(tmpDir, "cookies.txt")

	content := `session=abc123def456
token=xyz789
user_id=12345

# Comment line
remember_me=true
`

	if err := os.WriteFile(cookiesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cookies, err := ParseCookiesFile(cookiesFile)
	if err != nil {
		t.Fatalf("ParseCookiesFile() error = %v", err)
	}

	expectedCount := 4
	if len(cookies) != expectedCount {
		t.Errorf("ParseCookiesFile() got %d cookies, want %d", len(cookies), expectedCount)
	}

	if cookies["session"] != "abc123def456" {
		t.Errorf("session cookie = %v, want 'abc123def456'", cookies["session"])
	}

	if cookies["token"] != "xyz789" {
		t.Errorf("token cookie = %v, want 'xyz789'", cookies["token"])
	}

	if cookies["remember_me"] != "true" {
		t.Errorf("remember_me cookie = %v, want 'true'", cookies["remember_me"])
	}
}

func TestParseCookiesFile_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	cookiesFile := filepath.Join(tmpDir, "invalid_cookies.txt")

	content := `valid=value
InvalidCookieWithoutEquals
another=valid
`

	if err := os.WriteFile(cookiesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseCookiesFile(cookiesFile)
	if err == nil {
		t.Error("ParseCookiesFile() expected error for invalid format")
	}
}

func TestParseCookieString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "Single cookie",
			input: "session=abc123",
			expected: map[string]string{
				"session": "abc123",
			},
		},
		{
			name:  "Multiple cookies",
			input: "session=abc123; token=xyz789; user=john",
			expected: map[string]string{
				"session": "abc123",
				"token":   "xyz789",
				"user":    "john",
			},
		},
		{
			name:  "Cookies with spaces",
			input: "session = abc123 ; token = xyz789",
			expected: map[string]string{
				"session": "abc123",
				"token":   "xyz789",
			},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCookieString(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("ParseCookieString() got %d cookies, want %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("Cookie %s = %v, want %v", key, result[key], expectedValue)
				}
			}
		})
	}
}

func TestParseBasicAuth(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantUsername string
		wantPassword string
		wantErr      bool
	}{
		{
			name:         "Username and password",
			input:        "user:pass",
			wantUsername: "user",
			wantPassword: "pass",
			wantErr:      false,
		},
		{
			name:         "Username only",
			input:        "user",
			wantUsername: "user",
			wantPassword: "",
			wantErr:      false,
		},
		{
			name:         "Password with colons",
			input:        "user:pass:word:123",
			wantUsername: "user",
			wantPassword: "pass:word:123",
			wantErr:      false,
		},
		{
			name:         "Empty string",
			input:        "",
			wantUsername: "",
			wantPassword: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, password, err := ParseBasicAuth(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBasicAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if username != tt.wantUsername {
				t.Errorf("ParseBasicAuth() username = %v, want %v", username, tt.wantUsername)
			}

			if password != tt.wantPassword {
				t.Errorf("ParseBasicAuth() password = %v, want %v", password, tt.wantPassword)
			}
		})
	}
}
