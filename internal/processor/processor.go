package processor

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lcalzada-xor/downurl/internal/filter"
	"github.com/lcalzada-xor/downurl/internal/jsanalyzer"
	"github.com/lcalzada-xor/downurl/internal/output"
	"github.com/lcalzada-xor/downurl/internal/scanner"
	"github.com/lcalzada-xor/downurl/pkg/models"
)

// Processor handles post-download processing
type Processor struct {
	scanSecrets     bool
	scanEndpoints   bool
	jsBeautify      bool
	secretScanner   *scanner.SecretScanner
	endpointScanner *scanner.EndpointScanner
	beautifier      *jsanalyzer.Beautifier
	reporter        *output.Reporter
}

// Config represents processor configuration
type Config struct {
	ScanSecrets    bool
	ScanEndpoints  bool
	JSBeautify     bool
	SecretsEntropy float64
}

// NewProcessor creates a new processor
func NewProcessor(cfg Config) *Processor {
	p := &Processor{
		scanSecrets:   cfg.ScanSecrets,
		scanEndpoints: cfg.ScanEndpoints,
		jsBeautify:    cfg.JSBeautify,
		reporter:      output.NewReporter(),
	}

	if cfg.ScanSecrets {
		p.secretScanner = scanner.NewSecretScanner(cfg.SecretsEntropy)
	}

	if cfg.ScanEndpoints {
		p.endpointScanner = scanner.NewEndpointScanner()
	}

	if cfg.JSBeautify {
		p.beautifier = jsanalyzer.NewBeautifier()
	}

	return p
}

// ProcessResult processes a single download result
func (p *Processor) ProcessResult(result models.DownloadResult, outputDir string) error {
	if !result.IsSuccess() {
		return nil
	}

	// Process each downloaded file
	for _, filePath := range result.Downloaded {
		if err := p.processFile(filePath, result.URL, outputDir); err != nil {
			// Log error but continue
			continue
		}
	}

	return nil
}

// processFile processes a single file
func (p *Processor) processFile(filePath, url, outputDir string) error {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Detect content type
	contentType := filter.DetectContentType(data, filePath)

	// Calculate SHA256
	hash := sha256.Sum256(data)
	sha256Hash := fmt.Sprintf("%x", hash)

	// Add to reporter
	downloadInfo := output.DownloadInfo{
		URL:         url,
		Path:        filePath,
		SizeBytes:   int64(len(data)),
		ContentType: contentType,
		SHA256:      sha256Hash,
		Status:      "success",
	}
	p.reporter.AddDownload(downloadInfo)

	// Process based on content type
	isJS := filter.IsJavaScript(contentType) || strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".mjs")

	if isJS {
		// JS-specific processing
		if err := p.processJavaScript(filePath, url, data, outputDir); err != nil {
			// Log error but continue
		}
	}

	// General text file processing
	if filter.IsText(contentType) {
		if p.scanSecrets {
			secrets, err := p.secretScanner.ScanFile(filePath, url)
			if err == nil && len(secrets) > 0 {
				p.reporter.AddSecrets(secrets)
			}
		}

		if p.scanEndpoints {
			endpoints, err := p.endpointScanner.ScanFile(filePath, url)
			if err == nil && len(endpoints) > 0 {
				p.reporter.AddEndpoints(endpoints)
			}
		}
	}

	return nil
}

// processJavaScript processes JavaScript files
func (p *Processor) processJavaScript(filePath, url string, data []byte, outputDir string) error {
	code := string(data)

	// Check if minified
	if p.jsBeautify && jsanalyzer.IsMinified(code) {
		// Beautify
		beautified := p.beautifier.Beautify(code)

		// Save beautified version
		beautifiedPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".beautified.js"
		if err := os.WriteFile(beautifiedPath, []byte(beautified), 0644); err != nil {
			return fmt.Errorf("failed to write beautified file: %w", err)
		}

		// Scan beautified version instead
		if p.scanSecrets {
			secrets, err := p.secretScanner.ScanFile(beautifiedPath, url)
			if err == nil && len(secrets) > 0 {
				p.reporter.AddSecrets(secrets)
			}
		}

		if p.scanEndpoints {
			endpoints, err := p.endpointScanner.ScanFile(beautifiedPath, url)
			if err == nil && len(endpoints) > 0 {
				p.reporter.AddEndpoints(endpoints)
			}
		}
	}

	return nil
}

// GetReporter returns the reporter
func (p *Processor) GetReporter() *output.Reporter {
	return p.reporter
}

// SaveSecrets saves secrets to JSON file
func (p *Processor) SaveSecrets(filepath string) error {
	report := p.reporter.GetReport()
	if len(report.Findings.Secrets) == 0 {
		return nil
	}

	data, err := json.MarshalIndent(report.Findings.Secrets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write secrets file: %w", err)
	}

	return nil
}

// SaveEndpoints saves endpoints to JSON file
func (p *Processor) SaveEndpoints(filepath string) error {
	report := p.reporter.GetReport()
	if len(report.Findings.Endpoints) == 0 {
		return nil
	}

	data, err := json.MarshalIndent(report.Findings.Endpoints, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal endpoints: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write endpoints file: %w", err)
	}

	return nil
}
