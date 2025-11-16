package config

import (
	"fmt"

	"github.com/llvch/downurl/internal/auth"
)

// BuildAuthProvider creates an auth provider from configuration
func (c *Config) BuildAuthProvider() (*auth.Provider, error) {
	// Determine authentication type
	var authType auth.AuthType
	var authCfg auth.Config

	// Check for conflicting auth methods
	authMethodsCount := 0
	if c.AuthBearer != "" {
		authMethodsCount++
	}
	if c.AuthBasic != "" {
		authMethodsCount++
	}
	if c.AuthHeader != "" {
		authMethodsCount++
	}

	if authMethodsCount > 1 {
		return nil, fmt.Errorf("multiple authentication methods specified (use only one of: -auth-bearer, -auth-basic, -auth-header)")
	}

	// Configure authentication based on flags
	if c.AuthBearer != "" {
		authType = auth.AuthTypeBearer
		authCfg.Type = authType
		authCfg.Token = c.AuthBearer
	} else if c.AuthBasic != "" {
		authType = auth.AuthTypeBasic
		authCfg.Type = authType

		username, password, err := auth.ParseBasicAuth(c.AuthBasic)
		if err != nil {
			return nil, fmt.Errorf("invalid basic auth format: %w", err)
		}
		authCfg.Username = username
		authCfg.Password = password
	} else if c.AuthHeader != "" {
		authType = auth.AuthTypeCustom
		authCfg.Type = authType
		authCfg.Headers = map[string]string{
			"Authorization": c.AuthHeader,
		}
	} else {
		authType = auth.AuthTypeNone
		authCfg.Type = authType
	}

	// Initialize headers map if needed
	if authCfg.Headers == nil {
		authCfg.Headers = make(map[string]string)
	}

	// Load custom headers from file
	if c.HeadersFile != "" {
		headers, err := auth.ParseHeadersFile(c.HeadersFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse headers file: %w", err)
		}
		for k, v := range headers {
			authCfg.Headers[k] = v
		}
	}

	// Add User-Agent if specified
	if c.UserAgent != "" {
		authCfg.Headers["User-Agent"] = c.UserAgent
	}

	// Initialize cookies map if needed
	if authCfg.Cookies == nil {
		authCfg.Cookies = make(map[string]string)
	}

	// Load cookies from file
	if c.CookiesFile != "" {
		cookies, err := auth.ParseCookiesFile(c.CookiesFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse cookies file: %w", err)
		}
		for k, v := range cookies {
			authCfg.Cookies[k] = v
		}
	}

	// Parse cookie string
	if c.CookieString != "" {
		cookies := auth.ParseCookieString(c.CookieString)
		for k, v := range cookies {
			authCfg.Cookies[k] = v
		}
	}

	// If we have headers or cookies but no auth type, use custom
	if authType == auth.AuthTypeNone && (len(authCfg.Headers) > 0 || len(authCfg.Cookies) > 0) {
		authCfg.Type = auth.AuthTypeCustom
	}

	// Create and return provider
	return auth.NewProvider(authCfg)
}
