package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressBar displays download progress
type ProgressBar struct {
	total       int
	current     int
	startTime   time.Time
	totalBytes  int64
	mu          sync.Mutex
	width       int
	showSpeed   bool
	lastUpdate  time.Time
	updateDelay time.Duration
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, showSpeed bool) *ProgressBar {
	return &ProgressBar{
		total:       total,
		current:     0,
		startTime:   time.Now(),
		width:       50,
		showSpeed:   showSpeed,
		lastUpdate:  time.Time{},
		updateDelay: 100 * time.Millisecond,
	}
}

// Increment increases progress by 1
func (pb *ProgressBar) Increment(bytes int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.current++
	pb.totalBytes += bytes
}

// Update sets the current progress value
func (pb *ProgressBar) Update(current int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.current = current
}

// Render returns the progress bar string
func (pb *ProgressBar) Render() string {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Throttle updates
	if time.Since(pb.lastUpdate) < pb.updateDelay && pb.current < pb.total {
		return ""
	}
	pb.lastUpdate = time.Now()

	if pb.total == 0 {
		return ""
	}

	percentage := float64(pb.current) / float64(pb.total) * 100
	filled := int(float64(pb.width) * float64(pb.current) / float64(pb.total))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)

	elapsed := time.Since(pb.startTime)

	// Calculate speed safely (avoid division by zero)
	var speed float64
	if elapsed.Seconds() > 0 {
		speed = float64(pb.totalBytes) / elapsed.Seconds() / 1024 / 1024 // MB/s
	}

	eta := ""
	if pb.current > 0 && pb.current < pb.total {
		remaining := pb.total - pb.current
		avgTime := elapsed / time.Duration(pb.current)
		etaDuration := avgTime * time.Duration(remaining)
		eta = fmt.Sprintf(" | ETA: %s", formatDuration(etaDuration))
	}

	result := fmt.Sprintf("\rProgress: [%s] %.1f%% (%d/%d files)",
		bar, percentage, pb.current, pb.total)

	if pb.showSpeed && pb.totalBytes > 0 && speed > 0 {
		result += fmt.Sprintf(" | %.2f MB/s | Downloaded: %s%s",
			speed, formatBytes(pb.totalBytes), eta)
	}

	return result
}

// Finish completes the progress bar
func (pb *ProgressBar) Finish() {
	pb.mu.Lock()
	pb.current = pb.total
	pb.mu.Unlock()
	fmt.Println(pb.Render())
}

// formatBytes formats bytes to human readable format
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

// formatDuration formats duration to human readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), d.Seconds()-d.Minutes()*60)
	}
	return fmt.Sprintf("%.0fh%.0fm", d.Hours(), d.Minutes()-d.Hours()*60)
}

// Colors for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// Colorize adds color to text if supported
func Colorize(text string, color string) string {
	// TODO: Check if terminal supports colors
	return color + text + ColorReset
}

// Success prints a success message
func Success(msg string) {
	fmt.Printf("✓ %s\n", Colorize(msg, ColorGreen))
}

// Error prints an error message
func Error(msg string) {
	fmt.Printf("✗ %s\n", Colorize(msg, ColorRed))
}

// Warning prints a warning message
func Warning(msg string) {
	fmt.Printf("⚠ %s\n", Colorize(msg, ColorYellow))
}

// Info prints an info message
func Info(msg string) {
	fmt.Printf("ℹ %s\n", Colorize(msg, ColorBlue))
}
