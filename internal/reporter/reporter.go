package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/lcalzada-xor/downurl/pkg/models"
)

// Reporter collects and generates reports from download results
type Reporter struct {
	results []models.DownloadResult
	mu      sync.Mutex
}

// New creates a new Reporter instance
func New() *Reporter {
	return &Reporter{
		results: make([]models.DownloadResult, 0),
	}
}

// Add adds a download result to the reporter (thread-safe)
func (r *Reporter) Add(result models.DownloadResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results = append(r.results, result)
}

// AddBatch adds multiple results at once (thread-safe)
func (r *Reporter) AddBatch(results []models.DownloadResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results = append(r.results, results...)
}

// Generate creates a text report file
func (r *Reporter) Generate(outputPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "Download Report\n")
	fmt.Fprintf(file, "Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "Total URLs: %d\n", len(r.results))
	fmt.Fprintf(file, "%s\n\n", separator(60))

	// Calculate statistics
	stats := r.calculateStats()
	fmt.Fprintf(file, "Statistics:\n")
	fmt.Fprintf(file, "  Successful: %d\n", stats.Successful)
	fmt.Fprintf(file, "  Failed: %d\n", stats.Failed)
	fmt.Fprintf(file, "  Total Downloaded: %d files\n", stats.TotalDownloaded)
	fmt.Fprintf(file, "  Total Errors: %d\n", stats.TotalErrors)
	fmt.Fprintf(file, "  Average Duration: %v\n", stats.AvgDuration)
	fmt.Fprintf(file, "%s\n\n", separator(60))

	// Write individual results
	fmt.Fprintf(file, "Detailed Results:\n\n")

	// Sort results by URL for consistent output
	sortedResults := make([]models.DownloadResult, len(r.results))
	copy(sortedResults, r.results)
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].URL < sortedResults[j].URL
	})

	for i, result := range sortedResults {
		fmt.Fprintf(file, "[%d] URL: %s\n", i+1, result.URL)
		fmt.Fprintf(file, "    Host: %s\n", result.Host)
		fmt.Fprintf(file, "    Duration: %v\n", result.Duration)
		fmt.Fprintf(file, "    Downloaded: %d files\n", len(result.Downloaded))

		for _, path := range result.Downloaded {
			fmt.Fprintf(file, "      - %s\n", path)
		}

		fmt.Fprintf(file, "    Errors: %d\n", len(result.Errors))
		for _, errMsg := range result.Errors {
			fmt.Fprintf(file, "      - %s\n", errMsg)
		}

		fmt.Fprintf(file, "\n")
	}

	return nil
}

// Stats holds aggregated statistics
type Stats struct {
	Successful      int
	Failed          int
	TotalDownloaded int
	TotalErrors     int
	AvgDuration     time.Duration
}

// calculateStats computes statistics from results
func (r *Reporter) calculateStats() Stats {
	stats := Stats{}
	var totalDuration time.Duration

	for _, result := range r.results {
		if result.IsSuccess() {
			stats.Successful++
		} else {
			stats.Failed++
		}

		stats.TotalDownloaded += len(result.Downloaded)
		stats.TotalErrors += len(result.Errors)
		totalDuration += result.Duration
	}

	if len(r.results) > 0 {
		stats.AvgDuration = totalDuration / time.Duration(len(r.results))
	}

	return stats
}

// GetResults returns a copy of all results (thread-safe)
func (r *Reporter) GetResults() []models.DownloadResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	results := make([]models.DownloadResult, len(r.results))
	copy(results, r.results)
	return results
}

// separator generates a separator line
func separator(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "="
	}
	return result
}
