package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ConfigFile represents a .downurlrc configuration file
type ConfigFile struct {
	Defaults  map[string]string
	Auth      map[string]map[string]string
	Filters   map[string]string
	RateLimit map[string]string
}

// LoadConfigFile loads configuration from .downurlrc
func LoadConfigFile() (*ConfigFile, error) {
	// Try in order: ./.downurlrc, ~/.downurlrc
	paths := []string{
		".downurlrc",
		filepath.Join(os.Getenv("HOME"), ".downurlrc"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return parseConfigFile(path)
		}
	}

	// No config file found, return empty config
	return &ConfigFile{
		Defaults:  make(map[string]string),
		Auth:      make(map[string]map[string]string),
		Filters:   make(map[string]string),
		RateLimit: make(map[string]string),
	}, nil
}

// parseConfigFile parses a simple INI-style config file
func parseConfigFile(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cf := &ConfigFile{
		Defaults:  make(map[string]string),
		Auth:      make(map[string]map[string]string),
		Filters:   make(map[string]string),
		RateLimit: make(map[string]string),
	}

	lines := strings.Split(string(data), "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		// Key-value pair
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		// Expand environment variables
		if strings.Contains(value, "${") {
			value = os.ExpandEnv(value)
		}

		switch currentSection {
		case "defaults":
			cf.Defaults[key] = value
		case "filters":
			cf.Filters[key] = value
		case "ratelimit":
			cf.RateLimit[key] = value
		default:
			// Auth section (format: [auth.example.com])
			if strings.HasPrefix(currentSection, "auth.") {
				host := strings.TrimPrefix(currentSection, "auth.")
				if cf.Auth[host] == nil {
					cf.Auth[host] = make(map[string]string)
				}
				cf.Auth[host][key] = value
			}
		}
	}

	return cf, nil
}

// ApplyToConfig applies config file settings to Config
func (cf *ConfigFile) ApplyToConfig(c *Config) {
	// Apply defaults if not set via CLI
	if c.OutputDir == "output" && cf.Defaults["output"] != "" {
		c.OutputDir = cf.Defaults["output"]
	}

	if c.StorageMode == "flat" && cf.Defaults["mode"] != "" {
		c.StorageMode = cf.Defaults["mode"]
	}

	if c.Workers == 10 && cf.Defaults["workers"] != "" {
		if workers, err := strconv.Atoi(cf.Defaults["workers"]); err == nil {
			c.Workers = workers
		}
	}

	if c.Timeout == 15*time.Second && cf.Defaults["timeout"] != "" {
		if timeout, err := time.ParseDuration(cf.Defaults["timeout"]); err == nil {
			c.Timeout = timeout
		}
	}

	// Apply filters
	if c.FilterExt == "" && cf.Filters["extensions"] != "" {
		c.FilterExt = cf.Filters["extensions"]
	}

	if c.ExcludeExt == "" && cf.Filters["exclude_extensions"] != "" {
		c.ExcludeExt = cf.Filters["exclude_extensions"]
	}

	if c.MaxSize == 0 && cf.Filters["max_size"] != "" {
		if size, err := parseSize(cf.Filters["max_size"]); err == nil {
			c.MaxSize = size
		}
	}
}

// parseSize parses size strings like "50MB", "1GB"
func parseSize(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))

	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	for suffix, multiplier := range multipliers {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, err
			}
			return int64(num * float64(multiplier)), nil
		}
	}

	return strconv.ParseInt(s, 10, 64)
}

// SaveConfigFile saves current config to .downurlrc
func SaveConfigFile(c *Config, path string) error {
	var sb strings.Builder

	sb.WriteString("# Downurl Configuration File\n")
	sb.WriteString("# Generated on " + time.Now().Format(time.RFC3339) + "\n\n")

	sb.WriteString("[defaults]\n")
	sb.WriteString(fmt.Sprintf("mode = %s\n", c.StorageMode))
	sb.WriteString(fmt.Sprintf("workers = %d\n", c.Workers))
	sb.WriteString(fmt.Sprintf("timeout = %s\n", c.Timeout.String()))
	sb.WriteString(fmt.Sprintf("output = %s\n", c.OutputDir))
	sb.WriteString("\n")

	if c.FilterExt != "" || c.ExcludeExt != "" || c.MaxSize > 0 {
		sb.WriteString("[filters]\n")
		if c.FilterExt != "" {
			sb.WriteString(fmt.Sprintf("extensions = %s\n", c.FilterExt))
		}
		if c.ExcludeExt != "" {
			sb.WriteString(fmt.Sprintf("exclude_extensions = %s\n", c.ExcludeExt))
		}
		if c.MaxSize > 0 {
			sb.WriteString(fmt.Sprintf("max_size = %d\n", c.MaxSize))
		}
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
