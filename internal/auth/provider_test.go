package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewProvider_Bearer(t *testing.T) {
	cfg := Config{
		Type:  AuthTypeBearer,
		Token: "test-token-123",
	}

	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	if provider.GetType() != AuthTypeBearer {
		t.Errorf("GetType() = %v, want %v", provider.GetType(), AuthTypeBearer)
	}
}

func TestNewProvider_Basic(t *testing.T) {
	cfg := Config{
		Type:     AuthTypeBasic,
		Username: "testuser",
		Password: "testpass",
	}

	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	if provider.GetType() != AuthTypeBasic {
		t.Errorf("GetType() = %v, want %v", provider.GetType(), AuthTypeBasic)
	}
}

func TestNewProvider_Validation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "Bearer without token",
			cfg: Config{
				Type: AuthTypeBearer,
			},
			wantErr: true,
		},
		{
			name: "Basic without username",
			cfg: Config{
				Type:     AuthTypeBasic,
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "None is valid",
			cfg: Config{
				Type: AuthTypeNone,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProvider(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyAuth_Bearer(t *testing.T) {
	provider, err := NewProvider(Config{
		Type:  AuthTypeBearer,
		Token: "test-token-123",
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	err = provider.ApplyAuth(req)
	if err != nil {
		t.Fatalf("ApplyAuth() error = %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		t.Errorf("Authorization header = %v, should start with 'Bearer '", authHeader)
	}

	if !strings.Contains(authHeader, "test-token-123") {
		t.Errorf("Authorization header = %v, should contain token", authHeader)
	}
}

func TestApplyAuth_BearerWithPrefix(t *testing.T) {
	provider, err := NewProvider(Config{
		Type:  AuthTypeBearer,
		Token: "Bearer test-token-123",
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	err = provider.ApplyAuth(req)
	if err != nil {
		t.Fatalf("ApplyAuth() error = %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	expected := "Bearer test-token-123"
	if authHeader != expected {
		t.Errorf("Authorization header = %v, want %v", authHeader, expected)
	}
}

func TestApplyAuth_Basic(t *testing.T) {
	provider, err := NewProvider(Config{
		Type:     AuthTypeBasic,
		Username: "testuser",
		Password: "testpass",
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	err = provider.ApplyAuth(req)
	if err != nil {
		t.Fatalf("ApplyAuth() error = %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Basic ") {
		t.Errorf("Authorization header = %v, should start with 'Basic '", authHeader)
	}
}

func TestApplyAuth_CustomHeaders(t *testing.T) {
	provider, err := NewProvider(Config{
		Type: AuthTypeCustom,
		Headers: map[string]string{
			"X-API-Key":      "secret123",
			"X-Custom-Header": "value",
		},
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	err = provider.ApplyAuth(req)
	if err != nil {
		t.Fatalf("ApplyAuth() error = %v", err)
	}

	if req.Header.Get("X-API-Key") != "secret123" {
		t.Errorf("X-API-Key header = %v, want secret123", req.Header.Get("X-API-Key"))
	}

	if req.Header.Get("X-Custom-Header") != "value" {
		t.Errorf("X-Custom-Header header = %v, want value", req.Header.Get("X-Custom-Header"))
	}
}

func TestApplyAuth_Cookies(t *testing.T) {
	provider, err := NewProvider(Config{
		Type: AuthTypeCustom,
		Cookies: map[string]string{
			"session": "abc123",
			"token":   "xyz789",
		},
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	err = provider.ApplyAuth(req)
	if err != nil {
		t.Fatalf("ApplyAuth() error = %v", err)
	}

	cookies := req.Cookies()
	if len(cookies) != 2 {
		t.Errorf("Number of cookies = %d, want 2", len(cookies))
	}

	cookieMap := make(map[string]string)
	for _, cookie := range cookies {
		cookieMap[cookie.Name] = cookie.Value
	}

	if cookieMap["session"] != "abc123" {
		t.Errorf("session cookie = %v, want abc123", cookieMap["session"])
	}

	if cookieMap["token"] != "xyz789" {
		t.Errorf("token cookie = %v, want xyz789", cookieMap["token"])
	}
}

func TestApplyAuth_Nil(t *testing.T) {
	var provider *Provider
	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)

	err := provider.ApplyAuth(req)
	if err != nil {
		t.Errorf("ApplyAuth() on nil provider should not error, got %v", err)
	}
}
