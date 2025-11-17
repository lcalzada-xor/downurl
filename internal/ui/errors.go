package ui

import (
	"fmt"
	"strings"
)

// FriendlyError wraps an error with user-friendly messages and suggestions
type FriendlyError struct {
	Title       string
	Description string
	Suggestion  string
	Example     string
	OriginalErr error
}

// Error implements error interface
func (fe *FriendlyError) Error() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(Colorize("❌ "+fe.Title, ColorRed) + "\n\n")

	if fe.Description != "" {
		sb.WriteString(fe.Description + "\n\n")
	}

	if fe.Suggestion != "" {
		sb.WriteString(Colorize("💡 Suggestion:", ColorYellow) + "\n")
		sb.WriteString("   " + fe.Suggestion + "\n\n")
	}

	if fe.Example != "" {
		sb.WriteString(Colorize("📝 Example:", ColorCyan) + "\n")
		sb.WriteString("   " + fe.Example + "\n\n")
	}

	if fe.OriginalErr != nil {
		sb.WriteString(Colorize("Technical details:", ColorWhite) + "\n")
		sb.WriteString("   " + fe.OriginalErr.Error() + "\n")
	}

	return sb.String()
}

// WrapFileNotFound creates a friendly error for file not found
func WrapFileNotFound(filename string, err error) *FriendlyError {
	return &FriendlyError{
		Title:       fmt.Sprintf("File not found: %s", filename),
		Description: fmt.Sprintf("The file '%s' does not exist or cannot be accessed.", filename),
		Suggestion:  "Make sure the file path is correct and the file exists.",
		Example:     fmt.Sprintf("Create the file: echo \"https://example.com/file.js\" > %s", filename),
		OriginalErr: err,
	}
}

// WrapInvalidURL creates a friendly error for invalid URL
func WrapInvalidURL(url string, lineNum int, err error) *FriendlyError {
	desc := fmt.Sprintf("Line %d: \"%s\"\n", lineNum, url)

	// Analyze the error
	reason := "This doesn't look like a valid URL"
	suggestion := "URLs must start with http:// or https://"
	example := "https://example.com/file.js"

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if strings.Contains(url, "://") {
			reason = "Unsupported protocol"
			suggestion = "Only http:// and https:// URLs are supported"
		} else {
			reason = "Missing protocol"
			suggestion = "Did you forget to add https:// at the beginning?"
			example = fmt.Sprintf("https://%s", url)
		}
	} else if !strings.Contains(url, ".") {
		reason = "Invalid hostname"
		suggestion = "The URL doesn't have a valid domain name"
	}

	desc += fmt.Sprintf("   %s\n   %s", strings.Repeat("^", len(url)), reason)

	return &FriendlyError{
		Title:       "Invalid URL",
		Description: desc,
		Suggestion:  suggestion,
		Example:     example,
		OriginalErr: err,
	}
}

// WrapNetworkError creates a friendly error for network issues
func WrapNetworkError(url string, err error) *FriendlyError {
	var suggestion string
	var title string

	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "timeout"):
		title = "Connection timeout"
		suggestion = "The server took too long to respond. Try:\n" +
			"     - Increasing timeout: --timeout 30s\n" +
			"     - Checking your internet connection\n" +
			"     - Verifying the URL is accessible"
	case strings.Contains(errStr, "connection refused"):
		title = "Connection refused"
		suggestion = "The server refused the connection. The server might be down or the URL might be incorrect."
	case strings.Contains(errStr, "no such host"):
		title = "Host not found"
		suggestion = "The hostname doesn't exist or can't be resolved. Check if the URL is correct."
	case strings.Contains(errStr, "TLS"):
		title = "SSL/TLS error"
		suggestion = "There's a problem with the secure connection. The site's certificate might be invalid."
	default:
		title = "Network error"
		suggestion = "A network error occurred. Check your internet connection and the URL."
	}

	return &FriendlyError{
		Title:       title,
		Description: fmt.Sprintf("Failed to download: %s", url),
		Suggestion:  suggestion,
		OriginalErr: err,
	}
}

// WrapPermissionError creates a friendly error for permission issues
func WrapPermissionError(path string, err error) *FriendlyError {
	return &FriendlyError{
		Title:       "Permission denied",
		Description: fmt.Sprintf("Cannot write to: %s", path),
		Suggestion:  "Make sure you have write permissions for this directory, or choose a different output directory with --output",
		Example:     "./downurl -i urls.txt --output ~/downloads",
		OriginalErr: err,
	}
}

// WrapNoURLsError creates a friendly error for empty input
func WrapNoURLsError() *FriendlyError {
	return &FriendlyError{
		Title:       "No URLs found",
		Description: "The input file doesn't contain any valid URLs.",
		Suggestion:  "Make sure your file contains at least one URL per line.",
		Example: "echo \"https://example.com/file.js\" > urls.txt\n" +
			"   echo \"https://cdn.example.com/style.css\" >> urls.txt\n" +
			"   ./downurl -i urls.txt",
	}
}

// PrintUsageHint prints a helpful usage hint
func PrintUsageHint() {
	fmt.Println(Colorize("\n💡 Quick Start:", ColorCyan))
	fmt.Println("   1. Create a file with URLs (one per line)")
	fmt.Println(Colorize("      echo \"https://example.com/file.js\" > urls.txt", ColorWhite))
	fmt.Println()
	fmt.Println("   2. Run downurl")
	fmt.Println(Colorize("      ./downurl -i urls.txt", ColorWhite))
	fmt.Println()
	fmt.Println("   3. Find your files in the 'output' directory")
	fmt.Println()
	fmt.Println(Colorize("   For more help: ./downurl --help", ColorYellow))
	fmt.Println()
}
