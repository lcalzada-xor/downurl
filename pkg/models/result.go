package models

import "time"

// DownloadResult represents the result of downloading a file from a URL
type DownloadResult struct {
	URL        string        // Original URL
	Host       string        // Hostname extracted from URL
	Downloaded []string      // List of successfully downloaded file paths
	Errors     []string      // List of error messages
	Duration   time.Duration // Time taken to download
}

// Summary returns a summary of the download result
func (r *DownloadResult) Summary() (downloaded, errors int) {
	return len(r.Downloaded), len(r.Errors)
}

// IsSuccess returns true if the download was successful
func (r *DownloadResult) IsSuccess() bool {
	return len(r.Downloaded) > 0 && len(r.Errors) == 0
}
