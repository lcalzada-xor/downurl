package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// AuthType represents the type of authentication
type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeBearer AuthType = "bearer"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeCustom AuthType = "custom"
)

// Provider handles authentication for HTTP requests
type Provider struct {
	authType AuthType
	token    string
	username string
	password string
	headers  map[string]string
	cookies  map[string]string
}

// Config represents authentication configuration
type Config struct {
	Type     AuthType
	Token    string            // For Bearer token
	Username string            // For Basic auth
	Password string            // For Basic auth
	Headers  map[string]string // Custom headers
	Cookies  map[string]string // Custom cookies
}

// NewProvider creates a new authentication provider
func NewProvider(cfg Config) (*Provider, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return &Provider{
		authType: cfg.Type,
		token:    cfg.Token,
		username: cfg.Username,
		password: cfg.Password,
		headers:  cfg.Headers,
		cookies:  cfg.Cookies,
	}, nil
}

// ApplyAuth applies authentication to an HTTP request
func (p *Provider) ApplyAuth(req *http.Request) error {
	if p == nil {
		return nil
	}

	// Apply authentication based on type
	switch p.authType {
	case AuthTypeBearer:
		if err := p.applyBearer(req); err != nil {
			return err
		}
	case AuthTypeBasic:
		if err := p.applyBasic(req); err != nil {
			return err
		}
	}

	// Apply custom headers
	if err := p.applyHeaders(req); err != nil {
		return err
	}

	// Apply cookies
	if err := p.applyCookies(req); err != nil {
		return err
	}

	return nil
}

// applyBearer applies Bearer token authentication
func (p *Provider) applyBearer(req *http.Request) error {
	if p.token == "" {
		return fmt.Errorf("bearer token is empty")
	}

	// Support both formats: with and without "Bearer " prefix
	token := p.token
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = "Bearer " + token
	}

	req.Header.Set("Authorization", token)
	return nil
}

// applyBasic applies Basic authentication
func (p *Provider) applyBasic(req *http.Request) error {
	if p.username == "" {
		return fmt.Errorf("username is required for basic auth")
	}

	// Password can be empty (some APIs allow this)
	auth := p.username + ":" + p.password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encoded)
	return nil
}

// applyHeaders applies custom headers
func (p *Provider) applyHeaders(req *http.Request) error {
	for key, value := range p.headers {
		// Special handling for Authorization header
		if strings.ToLower(key) == "authorization" {
			// Only set if not already set by auth type
			if req.Header.Get("Authorization") == "" {
				req.Header.Set(key, value)
			}
		} else {
			req.Header.Set(key, value)
		}
	}
	return nil
}

// applyCookies applies cookies to the request
func (p *Provider) applyCookies(req *http.Request) error {
	for name, value := range p.cookies {
		cookie := &http.Cookie{
			Name:  name,
			Value: value,
		}
		req.AddCookie(cookie)
	}
	return nil
}

// validateConfig validates the authentication configuration
func validateConfig(cfg Config) error {
	switch cfg.Type {
	case AuthTypeNone:
		return nil
	case AuthTypeBearer:
		if cfg.Token == "" {
			return fmt.Errorf("token is required for bearer authentication")
		}
	case AuthTypeBasic:
		if cfg.Username == "" {
			return fmt.Errorf("username is required for basic authentication")
		}
	case AuthTypeCustom:
		if len(cfg.Headers) == 0 && len(cfg.Cookies) == 0 {
			return fmt.Errorf("headers or cookies required for custom authentication")
		}
	default:
		return fmt.Errorf("unsupported authentication type: %s", cfg.Type)
	}
	return nil
}

// GetType returns the authentication type
func (p *Provider) GetType() AuthType {
	if p == nil {
		return AuthTypeNone
	}
	return p.authType
}
