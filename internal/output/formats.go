package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lcalzada-xor/downurl/internal/scanner"
)

// Format represents output format type
type Format string

const (
	FormatText     Format = "text"
	FormatJSON     Format = "json"
	FormatCSV      Format = "csv"
	FormatMarkdown Format = "markdown"
)

// ScanReport represents a complete scan report
type ScanReport struct {
	Metadata  Metadata                  `json:"metadata"`
	Downloads []DownloadInfo            `json:"downloads"`
	Findings  Findings                  `json:"findings"`
	Statistics Statistics               `json:"statistics"`
}

// Metadata contains scan metadata
type Metadata struct {
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	DurationSeconds float64   `json:"duration_seconds"`
	TotalURLs      int       `json:"total_urls"`
	Successful     int       `json:"successful"`
	Failed         int       `json:"failed"`
}

// DownloadInfo contains download information
type DownloadInfo struct {
	URL          string    `json:"url"`
	Path         string    `json:"path"`
	SizeBytes    int64     `json:"size_bytes"`
	ContentType  string    `json:"content_type"`
	SHA256       string    `json:"sha256,omitempty"`
	DownloadedAt time.Time `json:"downloaded_at"`
	Status       string    `json:"status"`
	Error        string    `json:"error,omitempty"`
}

// Findings contains all findings
type Findings struct {
	Secrets   []scanner.SecretFinding   `json:"secrets,omitempty"`
	Endpoints []scanner.EndpointFinding `json:"endpoints,omitempty"`
}

// Statistics contains download statistics
type Statistics struct {
	TotalFiles         int            `json:"total_files"`
	TotalSizeBytes     int64          `json:"total_size_bytes"`
	ByContentType      map[string]int `json:"by_content_type"`
	SecretsCount       int            `json:"secrets_count"`
	EndpointsCount     int            `json:"endpoints_count"`
	HighConfidenceSecrets int         `json:"high_confidence_secrets"`
}

// Reporter generates output in different formats
type Reporter struct {
	report ScanReport
}

// NewReporter creates a new reporter
func NewReporter() *Reporter {
	return &Reporter{
		report: ScanReport{
			Downloads: []DownloadInfo{},
			Findings: Findings{
				Secrets:   []scanner.SecretFinding{},
				Endpoints: []scanner.EndpointFinding{},
			},
			Statistics: Statistics{
				ByContentType: make(map[string]int),
			},
		},
	}
}

// SetMetadata sets scan metadata
func (r *Reporter) SetMetadata(meta Metadata) {
	r.report.Metadata = meta
}

// AddDownload adds a download to the report
func (r *Reporter) AddDownload(info DownloadInfo) {
	r.report.Downloads = append(r.report.Downloads, info)

	// Update statistics
	if info.Status == "success" {
		r.report.Statistics.TotalFiles++
		r.report.Statistics.TotalSizeBytes += info.SizeBytes

		if info.ContentType != "" {
			r.report.Statistics.ByContentType[info.ContentType]++
		}
	}
}

// AddSecrets adds secret findings
func (r *Reporter) AddSecrets(secrets []scanner.SecretFinding) {
	r.report.Findings.Secrets = append(r.report.Findings.Secrets, secrets...)
	r.report.Statistics.SecretsCount = len(r.report.Findings.Secrets)

	// Count high confidence secrets
	for _, secret := range secrets {
		if secret.Confidence == scanner.ConfidenceHigh {
			r.report.Statistics.HighConfidenceSecrets++
		}
	}
}

// AddEndpoints adds endpoint findings
func (r *Reporter) AddEndpoints(endpoints []scanner.EndpointFinding) {
	r.report.Findings.Endpoints = append(r.report.Findings.Endpoints, endpoints...)
	r.report.Statistics.EndpointsCount = len(r.report.Findings.Endpoints)
}

// GenerateJSON generates JSON output
func (r *Reporter) GenerateJSON(filepath string, pretty bool) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if pretty {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(r.report); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// GenerateCSV generates CSV output
func (r *Reporter) GenerateCSV(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"URL", "Path", "Size", "ContentType", "SHA256", "Status", "Error"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, download := range r.report.Downloads {
		row := []string{
			download.URL,
			download.Path,
			fmt.Sprintf("%d", download.SizeBytes),
			download.ContentType,
			download.SHA256,
			download.Status,
			download.Error,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// GenerateMarkdown generates Markdown output
func (r *Reporter) GenerateMarkdown(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	var md strings.Builder

	// Title
	md.WriteString("# Download Scan Report\n\n")

	// Metadata
	md.WriteString("## Scan Information\n\n")
	md.WriteString(fmt.Sprintf("- **Start Time**: %s\n", r.report.Metadata.StartTime.Format(time.RFC3339)))
	md.WriteString(fmt.Sprintf("- **End Time**: %s\n", r.report.Metadata.EndTime.Format(time.RFC3339)))
	md.WriteString(fmt.Sprintf("- **Duration**: %.2f seconds\n", r.report.Metadata.DurationSeconds))
	md.WriteString(fmt.Sprintf("- **Total URLs**: %d\n", r.report.Metadata.TotalURLs))
	md.WriteString(fmt.Sprintf("- **Successful**: %d\n", r.report.Metadata.Successful))
	md.WriteString(fmt.Sprintf("- **Failed**: %d\n\n", r.report.Metadata.Failed))

	// Statistics
	md.WriteString("## Statistics\n\n")
	md.WriteString(fmt.Sprintf("- **Total Files**: %d\n", r.report.Statistics.TotalFiles))
	md.WriteString(fmt.Sprintf("- **Total Size**: %s\n", formatBytes(r.report.Statistics.TotalSizeBytes)))
	md.WriteString(fmt.Sprintf("- **Secrets Found**: %d (High Confidence: %d)\n",
		r.report.Statistics.SecretsCount, r.report.Statistics.HighConfidenceSecrets))
	md.WriteString(fmt.Sprintf("- **Endpoints Found**: %d\n\n", r.report.Statistics.EndpointsCount))

	// Content Types
	if len(r.report.Statistics.ByContentType) > 0 {
		md.WriteString("### Files by Content Type\n\n")
		for contentType, count := range r.report.Statistics.ByContentType {
			md.WriteString(fmt.Sprintf("- %s: %d files\n", contentType, count))
		}
		md.WriteString("\n")
	}

	// Secrets
	if len(r.report.Findings.Secrets) > 0 {
		md.WriteString("## 🔐 Secrets Found\n\n")

		// Group by confidence
		highConfidence := []scanner.SecretFinding{}
		mediumConfidence := []scanner.SecretFinding{}
		lowConfidence := []scanner.SecretFinding{}

		for _, secret := range r.report.Findings.Secrets {
			switch secret.Confidence {
			case scanner.ConfidenceHigh:
				highConfidence = append(highConfidence, secret)
			case scanner.ConfidenceMedium:
				mediumConfidence = append(mediumConfidence, secret)
			case scanner.ConfidenceLow:
				lowConfidence = append(lowConfidence, secret)
			}
		}

		if len(highConfidence) > 0 {
			md.WriteString("### ⚠️ High Confidence\n\n")
			for _, secret := range highConfidence {
				md.WriteString(fmt.Sprintf("- **%s**\n", secret.SecretType))
				md.WriteString(fmt.Sprintf("  - File: `%s:%d`\n", secret.File, secret.Line))
				md.WriteString(fmt.Sprintf("  - Match: `%s`\n", secret.Match))
				md.WriteString("\n")
			}
		}

		if len(mediumConfidence) > 0 {
			md.WriteString("### ⚡ Medium Confidence\n\n")
			for _, secret := range mediumConfidence {
				md.WriteString(fmt.Sprintf("- **%s**: `%s` in `%s:%d`\n",
					secret.SecretType, secret.Match, secret.File, secret.Line))
			}
			md.WriteString("\n")
		}

		if len(lowConfidence) > 0 {
			md.WriteString(fmt.Sprintf("### ℹ️ Low Confidence (%d findings)\n\n", len(lowConfidence)))
		}
	}

	// Endpoints
	if len(r.report.Findings.Endpoints) > 0 {
		md.WriteString("## 🌐 Endpoints Discovered\n\n")

		// Group by type
		byType := make(map[scanner.EndpointType][]scanner.EndpointFinding)
		for _, endpoint := range r.report.Findings.Endpoints {
			byType[endpoint.Type] = append(byType[endpoint.Type], endpoint)
		}

		for endpointType, endpoints := range byType {
			md.WriteString(fmt.Sprintf("### %s (%d)\n\n", endpointType, len(endpoints)))
			for _, endpoint := range endpoints {
				method := string(endpoint.Method)
				if method == "" {
					method = "GET"
				}
				md.WriteString(fmt.Sprintf("- `%s %s`\n", method, endpoint.Endpoint))
			}
			md.WriteString("\n")
		}
	}

	// Write to file
	if _, err := file.WriteString(md.String()); err != nil {
		return fmt.Errorf("failed to write markdown: %w", err)
	}

	return nil
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetReport returns the current report
func (r *Reporter) GetReport() ScanReport {
	return r.report
}
