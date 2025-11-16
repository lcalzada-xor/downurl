package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the downloader
type Config struct {
	InputFile     string        // Path to file containing URLs
	OutputDir     string        // Directory to save downloaded files
	Workers       int           // Number of concurrent workers
	Timeout       time.Duration // HTTP request timeout
	RetryAttempts int           // Number of retry attempts per download

	// Authentication options
	AuthBearer    string // Bearer token for authentication
	AuthBasic     string // Basic auth in format "username:password"
	AuthHeader    string // Custom Authorization header value
	HeadersFile   string // Path to file containing custom headers
	CookiesFile   string // Path to file containing cookies
	CookieString  string // Cookie string in format "name1=value1; name2=value2"
	UserAgent     string // Custom User-Agent header

	// Scanner options
	ScanSecrets     bool    // Enable secret scanning
	ScanEndpoints   bool    // Enable endpoint discovery
	SecretsEntropy  float64 // Minimum entropy for secret detection
	SecretsOutput   string  // Output file for secrets
	EndpointsOutput string  // Output file for endpoints

	// Filter options
	FilterType   string // Filter by content type (comma-separated)
	ExcludeType  string // Exclude content types (comma-separated)
	FilterExt    string // Filter by extension (comma-separated)
	ExcludeExt   string // Exclude extensions (comma-separated)
	MinSize      int64  // Minimum file size in bytes
	MaxSize      int64  // Maximum file size in bytes (0 = use default)
	SkipEmpty    bool   // Skip empty files

	// JS Analysis options
	JSBeautify       bool   // Beautify minified JavaScript
	ExtractStrings   bool   // Extract strings from JS files
	StringsMinLength int    // Minimum string length
	StringsPattern   string // Pattern to match in strings

	// Output options
	OutputFormat string // Output format: text, json, csv, markdown
	OutputFile   string // Output file path (for JSON/CSV/Markdown)
	PrettyJSON   bool   // Pretty print JSON
}

// Load parses command line flags and environment variables to create a Config
func Load() *Config {
	cfg := &Config{}

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: downurl --input <urls.txt> [options]\n")
		fmt.Fprintf(os.Stderr, "\nBasic Options:\n")
		fmt.Fprintf(os.Stderr, "  --input, -i string      Input file containing URLs (required)\n")
		fmt.Fprintf(os.Stderr, "  --output, -o string     Output directory (default: output)\n")
		fmt.Fprintf(os.Stderr, "  --workers, -w int       Number of concurrent workers (default: 10)\n")
		fmt.Fprintf(os.Stderr, "  --timeout, -t duration  HTTP request timeout (default: 15s)\n")
		fmt.Fprintf(os.Stderr, "  --retry, -r int         Number of retry attempts (default: 3)\n")
		fmt.Fprintf(os.Stderr, "\nAuthentication Options:\n")
		fmt.Fprintf(os.Stderr, "  --auth-bearer, -b string    Bearer token authentication\n")
		fmt.Fprintf(os.Stderr, "  --auth-basic, -B string     Basic auth (format: username:password)\n")
		fmt.Fprintf(os.Stderr, "  --auth-header, -H string    Custom Authorization header value\n")
		fmt.Fprintf(os.Stderr, "  --headers-file, -h string   File with custom headers (format: 'Name: value')\n")
		fmt.Fprintf(os.Stderr, "  --cookies-file, -C string   File with cookies (format: 'name=value')\n")
		fmt.Fprintf(os.Stderr, "  --cookie, -c string         Cookie string (format: 'name1=value1; name2=value2')\n")
		fmt.Fprintf(os.Stderr, "  --user-agent, -u string     Custom User-Agent header\n")
		fmt.Fprintf(os.Stderr, "\nScanner Options:\n")
		fmt.Fprintf(os.Stderr, "  --scan-secrets, -s          Enable secret scanning\n")
		fmt.Fprintf(os.Stderr, "  --scan-endpoints, -e        Enable endpoint discovery\n")
		fmt.Fprintf(os.Stderr, "  --secrets-entropy, -E float Minimum entropy for secret detection (default: 4.5)\n")
		fmt.Fprintf(os.Stderr, "  --secrets-output, -S string Output file for secrets (JSON)\n")
		fmt.Fprintf(os.Stderr, "  --endpoints-output, -O string Output file for endpoints (JSON)\n")
		fmt.Fprintf(os.Stderr, "\nFilter Options:\n")
		fmt.Fprintf(os.Stderr, "  --filter-type, -T string    Filter by content type (comma-separated)\n")
		fmt.Fprintf(os.Stderr, "  --exclude-type, -X string   Exclude content types (comma-separated)\n")
		fmt.Fprintf(os.Stderr, "  --filter-ext, -F string     Filter by extension (comma-separated)\n")
		fmt.Fprintf(os.Stderr, "  --exclude-ext, -x string    Exclude extensions (comma-separated)\n")
		fmt.Fprintf(os.Stderr, "  --min-size, -m int          Minimum file size in bytes\n")
		fmt.Fprintf(os.Stderr, "  --max-size, -M int          Maximum file size in bytes (0 = default 100MB)\n")
		fmt.Fprintf(os.Stderr, "  --skip-empty, -k            Skip empty files\n")
		fmt.Fprintf(os.Stderr, "\nJS Analysis Options:\n")
		fmt.Fprintf(os.Stderr, "  --js-beautify, -j           Beautify minified JavaScript\n")
		fmt.Fprintf(os.Stderr, "  --extract-strings, -a       Extract strings from JS files\n")
		fmt.Fprintf(os.Stderr, "  --strings-min-length, -l int Minimum string length (default: 10)\n")
		fmt.Fprintf(os.Stderr, "  --strings-pattern, -p string Pattern to match in strings (regex)\n")
		fmt.Fprintf(os.Stderr, "\nOutput Options:\n")
		fmt.Fprintf(os.Stderr, "  --output-format, -f string  Output format: text, json, csv, markdown (default: text)\n")
		fmt.Fprintf(os.Stderr, "  --output-file, -P string    Output file path (for JSON/CSV/Markdown)\n")
		fmt.Fprintf(os.Stderr, "  --pretty-json, -J           Pretty print JSON output (default: true)\n")
	}

	// Define flags with long and short versions
	// Basic flags
	flag.StringVar(&cfg.InputFile, "i", "", "Input file containing URLs (required) [shorthand]")
	flag.StringVar(&cfg.InputFile, "input", "", "Input file containing URLs (required)")
	flag.StringVar(&cfg.OutputDir, "o", getEnvOrDefault("OUTPUT_DIR", "output"), "Output directory [shorthand]")
	flag.StringVar(&cfg.OutputDir, "output", getEnvOrDefault("OUTPUT_DIR", "output"), "Output directory")
	flag.IntVar(&cfg.Workers, "w", getEnvIntOrDefault("WORKERS", 10), "Number of concurrent workers [shorthand]")
	flag.IntVar(&cfg.Workers, "workers", getEnvIntOrDefault("WORKERS", 10), "Number of concurrent workers")
	flag.DurationVar(&cfg.Timeout, "t", getEnvDurationOrDefault("TIMEOUT", 15*time.Second), "HTTP request timeout [shorthand]")
	flag.DurationVar(&cfg.Timeout, "timeout", getEnvDurationOrDefault("TIMEOUT", 15*time.Second), "HTTP request timeout")
	flag.IntVar(&cfg.RetryAttempts, "r", getEnvIntOrDefault("RETRY_ATTEMPTS", 3), "Number of retry attempts [shorthand]")
	flag.IntVar(&cfg.RetryAttempts, "retry", getEnvIntOrDefault("RETRY_ATTEMPTS", 3), "Number of retry attempts")

	// Authentication flags
	flag.StringVar(&cfg.AuthBearer, "b", getEnvOrDefault("AUTH_BEARER", ""), "Bearer token for authentication [shorthand]")
	flag.StringVar(&cfg.AuthBearer, "auth-bearer", getEnvOrDefault("AUTH_BEARER", ""), "Bearer token for authentication")
	flag.StringVar(&cfg.AuthBasic, "B", getEnvOrDefault("AUTH_BASIC", ""), "Basic auth (format: username:password) [shorthand]")
	flag.StringVar(&cfg.AuthBasic, "auth-basic", getEnvOrDefault("AUTH_BASIC", ""), "Basic auth (format: username:password)")
	flag.StringVar(&cfg.AuthHeader, "H", getEnvOrDefault("AUTH_HEADER", ""), "Custom Authorization header value [shorthand]")
	flag.StringVar(&cfg.AuthHeader, "auth-header", getEnvOrDefault("AUTH_HEADER", ""), "Custom Authorization header value")
	flag.StringVar(&cfg.HeadersFile, "h", "", "Path to file with custom headers (format: 'Name: value') [shorthand]")
	flag.StringVar(&cfg.HeadersFile, "headers-file", "", "Path to file with custom headers (format: 'Name: value')")
	flag.StringVar(&cfg.CookiesFile, "C", "", "Path to file with cookies (format: 'name=value') [shorthand]")
	flag.StringVar(&cfg.CookiesFile, "cookies-file", "", "Path to file with cookies (format: 'name=value')")
	flag.StringVar(&cfg.CookieString, "c", getEnvOrDefault("COOKIE", ""), "Cookie string (format: 'name1=value1; name2=value2') [shorthand]")
	flag.StringVar(&cfg.CookieString, "cookie", getEnvOrDefault("COOKIE", ""), "Cookie string (format: 'name1=value1; name2=value2')")
	flag.StringVar(&cfg.UserAgent, "u", getEnvOrDefault("USER_AGENT", ""), "Custom User-Agent header [shorthand]")
	flag.StringVar(&cfg.UserAgent, "user-agent", getEnvOrDefault("USER_AGENT", ""), "Custom User-Agent header")

	// Scanner flags
	flag.BoolVar(&cfg.ScanSecrets, "s", false, "Enable secret scanning [shorthand]")
	flag.BoolVar(&cfg.ScanSecrets, "scan-secrets", false, "Enable secret scanning")
	flag.BoolVar(&cfg.ScanEndpoints, "e", false, "Enable endpoint discovery [shorthand]")
	flag.BoolVar(&cfg.ScanEndpoints, "scan-endpoints", false, "Enable endpoint discovery")
	flag.Float64Var(&cfg.SecretsEntropy, "E", 4.5, "Minimum entropy for secret detection [shorthand]")
	flag.Float64Var(&cfg.SecretsEntropy, "secrets-entropy", 4.5, "Minimum entropy for secret detection")
	flag.StringVar(&cfg.SecretsOutput, "S", "", "Output file for secrets (JSON) [shorthand]")
	flag.StringVar(&cfg.SecretsOutput, "secrets-output", "", "Output file for secrets (JSON)")
	flag.StringVar(&cfg.EndpointsOutput, "O", "", "Output file for endpoints (JSON) [shorthand]")
	flag.StringVar(&cfg.EndpointsOutput, "endpoints-output", "", "Output file for endpoints (JSON)")

	// Filter flags
	flag.StringVar(&cfg.FilterType, "T", "", "Filter by content type (comma-separated) [shorthand]")
	flag.StringVar(&cfg.FilterType, "filter-type", "", "Filter by content type (comma-separated)")
	flag.StringVar(&cfg.ExcludeType, "X", "", "Exclude content types (comma-separated) [shorthand]")
	flag.StringVar(&cfg.ExcludeType, "exclude-type", "", "Exclude content types (comma-separated)")
	flag.StringVar(&cfg.FilterExt, "F", "", "Filter by extension (comma-separated) [shorthand]")
	flag.StringVar(&cfg.FilterExt, "filter-ext", "", "Filter by extension (comma-separated)")
	flag.StringVar(&cfg.ExcludeExt, "x", "", "Exclude extensions (comma-separated) [shorthand]")
	flag.StringVar(&cfg.ExcludeExt, "exclude-ext", "", "Exclude extensions (comma-separated)")
	flag.Int64Var(&cfg.MinSize, "m", 0, "Minimum file size in bytes [shorthand]")
	flag.Int64Var(&cfg.MinSize, "min-size", 0, "Minimum file size in bytes")
	flag.Int64Var(&cfg.MaxSize, "M", 0, "Maximum file size in bytes (0 = default 100MB) [shorthand]")
	flag.Int64Var(&cfg.MaxSize, "max-size", 0, "Maximum file size in bytes (0 = default 100MB)")
	flag.BoolVar(&cfg.SkipEmpty, "k", false, "Skip empty files [shorthand]")
	flag.BoolVar(&cfg.SkipEmpty, "skip-empty", false, "Skip empty files")

	// JS Analysis flags
	flag.BoolVar(&cfg.JSBeautify, "j", false, "Beautify minified JavaScript [shorthand]")
	flag.BoolVar(&cfg.JSBeautify, "js-beautify", false, "Beautify minified JavaScript")
	flag.BoolVar(&cfg.ExtractStrings, "a", false, "Extract strings from JS files [shorthand]")
	flag.BoolVar(&cfg.ExtractStrings, "extract-strings", false, "Extract strings from JS files")
	flag.IntVar(&cfg.StringsMinLength, "l", 10, "Minimum string length [shorthand]")
	flag.IntVar(&cfg.StringsMinLength, "strings-min-length", 10, "Minimum string length")
	flag.StringVar(&cfg.StringsPattern, "p", "", "Pattern to match in strings (regex) [shorthand]")
	flag.StringVar(&cfg.StringsPattern, "strings-pattern", "", "Pattern to match in strings (regex)")

	// Output flags
	flag.StringVar(&cfg.OutputFormat, "f", "text", "Output format: text, json, csv, markdown [shorthand]")
	flag.StringVar(&cfg.OutputFormat, "output-format", "text", "Output format: text, json, csv, markdown")
	flag.StringVar(&cfg.OutputFile, "P", "", "Output file path (for JSON/CSV/Markdown) [shorthand]")
	flag.StringVar(&cfg.OutputFile, "output-file", "", "Output file path (for JSON/CSV/Markdown)")
	flag.BoolVar(&cfg.PrettyJSON, "J", true, "Pretty print JSON output [shorthand]")
	flag.BoolVar(&cfg.PrettyJSON, "pretty-json", true, "Pretty print JSON output")

	flag.Parse()

	// Validate required fields
	if cfg.InputFile == "" && flag.NArg() > 0 {
		cfg.InputFile = flag.Arg(0)
	}

	return cfg
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.InputFile == "" {
		return ErrMissingInputFile
	}
	if c.Workers < 1 {
		c.Workers = 1
	}
	if c.Timeout < time.Second {
		c.Timeout = time.Second
	}
	return nil
}

// Helper functions to get environment variables with defaults
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
