package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/lcalzada-xor/downurl/pkg/models"
)

// ResultsTable displays download results in a table format
type ResultsTable struct {
	results []models.DownloadResult
}

// NewResultsTable creates a new results table
func NewResultsTable(results []models.DownloadResult) *ResultsTable {
	return &ResultsTable{results: results}
}

// Render renders the results table
func (rt *ResultsTable) Render() string {
	if len(rt.results) == 0 {
		return "No results to display"
	}

	var sb strings.Builder

	// Calculate column widths
	urlWidth := 40
	sizeWidth := 10
	timeWidth := 10
	statusWidth := 8

	// Header
	sb.WriteString("┌" + strings.Repeat("─", urlWidth+2) +
		"┬" + strings.Repeat("─", sizeWidth+2) +
		"┬" + strings.Repeat("─", timeWidth+2) +
		"┬" + strings.Repeat("─", statusWidth+2) + "┐\n")

	sb.WriteString(fmt.Sprintf("│ %-*s │ %-*s │ %-*s │ %-*s │\n",
		urlWidth, "URL",
		sizeWidth, "Size",
		timeWidth, "Time",
		statusWidth, "Status"))

	sb.WriteString("├" + strings.Repeat("─", urlWidth+2) +
		"┼" + strings.Repeat("─", sizeWidth+2) +
		"┼" + strings.Repeat("─", timeWidth+2) +
		"┼" + strings.Repeat("─", statusWidth+2) + "┤\n")

	// Rows (limit to 20 for display)
	displayCount := len(rt.results)
	if displayCount > 20 {
		displayCount = 20
	}

	for i := 0; i < displayCount; i++ {
		result := rt.results[i]

		// Truncate URL if too long
		url := result.URL
		if len(url) > urlWidth {
			url = url[:urlWidth-3] + "..."
		}

		// Calculate total size
		var totalSize int64
		for range result.Downloaded {
			// We don't have size info in result, using placeholder
			totalSize += 0 // TODO: Add size tracking
		}

		size := "-"
		if totalSize > 0 {
			size = formatBytes(totalSize)
		}

		duration := formatDuration(result.Duration)

		status := "✓"
		statusColor := ColorGreen
		if !result.IsSuccess() {
			status = "✗"
			statusColor = ColorRed
		}

		sb.WriteString(fmt.Sprintf("│ %-*s │ %-*s │ %-*s │ %s%-*s%s │\n",
			urlWidth, url,
			sizeWidth, size,
			timeWidth, duration,
			statusColor, statusWidth, status, ColorReset))
	}

	// Footer
	sb.WriteString("└" + strings.Repeat("─", urlWidth+2) +
		"┴" + strings.Repeat("─", sizeWidth+2) +
		"┴" + strings.Repeat("─", timeWidth+2) +
		"┴" + strings.Repeat("─", statusWidth+2) + "┘\n")

	if len(rt.results) > displayCount {
		sb.WriteString(fmt.Sprintf("\n... and %d more results (see full report)\n",
			len(rt.results)-displayCount))
	}

	return sb.String()
}

// RenderSummary renders a detailed summary
func RenderSummary(results []models.DownloadResult, elapsed time.Duration, outputDir string) string {
	var sb strings.Builder

	// Header
	sb.WriteString(strings.Repeat("═", 60) + "\n")
	sb.WriteString(Colorize("📊 Download Summary", ColorCyan) + "\n")
	sb.WriteString(strings.Repeat("═", 60) + "\n\n")

	// Calculate stats
	total := len(results)
	successful := 0
	failed := 0
	var totalBytes int64
	var totalErrors int

	for _, r := range results {
		if r.IsSuccess() {
			successful++
		} else {
			failed++
		}
		totalErrors += len(r.Errors)
	}

	// Duration and success rate
	sb.WriteString(fmt.Sprintf("⏱️  Duration: %s\n", Colorize(formatDuration(elapsed), ColorYellow)))
	successRate := float64(successful) / float64(total) * 100
	sb.WriteString(fmt.Sprintf("✓  Success: %s (%s)\n",
		Colorize(fmt.Sprintf("%d/%d", successful, total), ColorGreen),
		Colorize(fmt.Sprintf("%.1f%%", successRate), ColorGreen)))

	if failed > 0 {
		sb.WriteString(fmt.Sprintf("✗  Failed: %s (%s)\n",
			Colorize(fmt.Sprintf("%d", failed), ColorRed),
			Colorize(fmt.Sprintf("%.1f%%", float64(failed)/float64(total)*100), ColorRed)))
	}

	sb.WriteString("\n")

	// Performance
	sb.WriteString(Colorize("🚀 Performance:", ColorCyan) + "\n")
	avgSpeed := float64(totalBytes) / elapsed.Seconds() / 1024 / 1024
	if totalBytes > 0 {
		sb.WriteString(fmt.Sprintf("   - Average speed: %.2f MB/s\n", avgSpeed))
		sb.WriteString(fmt.Sprintf("   - Total downloaded: %s\n", formatBytes(totalBytes)))
	}
	sb.WriteString(fmt.Sprintf("   - Average time per file: %s\n",
		formatDuration(elapsed/time.Duration(total))))

	sb.WriteString("\n")

	// Storage info
	sb.WriteString(Colorize("💾 Storage:", ColorCyan) + "\n")
	sb.WriteString(fmt.Sprintf("   Location: %s\n", outputDir))

	sb.WriteString("\n")

	// Failures breakdown
	if totalErrors > 0 {
		sb.WriteString(Colorize("⚠️  Failures:", ColorYellow) + "\n")
		errorTypes := make(map[string]int)
		for _, r := range results {
			for _, err := range r.Errors {
				if strings.Contains(err, "timeout") {
					errorTypes["timeout"]++
				} else if strings.Contains(err, "404") || strings.Contains(err, "not found") {
					errorTypes["not found"]++
				} else if strings.Contains(err, "connection refused") {
					errorTypes["connection refused"]++
				} else {
					errorTypes["other"]++
				}
			}
		}

		i := 1
		for errType, count := range errorTypes {
			sb.WriteString(fmt.Sprintf("   %d. %s (%d files)\n", i, errType, count))
			i++
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("═", 60) + "\n")

	return sb.String()
}
