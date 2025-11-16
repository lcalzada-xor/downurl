package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/llvch/downurl/internal/auth"
)

const (
	// MaxDownloadSize is the maximum size of a single download (100MB)
	MaxDownloadSize = 100 * 1024 * 1024 // 100 MB
)

// HTTPClient wraps http.Client with retry logic and timeout
type HTTPClient struct {
	client        *http.Client
	timeout       time.Duration
	retryAttempts int
	maxSize       int64
	authProvider  *auth.Provider
}

// NewHTTPClient creates a new HTTP client with specified timeout and retry attempts
func NewHTTPClient(timeout time.Duration, retryAttempts int) *HTTPClient {
	return NewHTTPClientWithAuth(timeout, retryAttempts, nil)
}

// NewHTTPClientWithAuth creates a new HTTP client with authentication support
func NewHTTPClientWithAuth(timeout time.Duration, retryAttempts int, authProvider *auth.Provider) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
		timeout:       timeout,
		retryAttempts: retryAttempts,
		maxSize:       MaxDownloadSize,
		authProvider:  authProvider,
	}
}

// Download downloads content from a URL with retry logic (legacy method)
// Deprecated: Use DownloadToWriter for streaming downloads
func (c *HTTPClient) Download(ctx context.Context, url string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		data, err := c.doDownload(ctx, url)
		if err == nil {
			return data, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		if isClientError(err) {
			break
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", c.retryAttempts+1, lastErr)
}

// DownloadToWriter downloads content from a URL and writes it to the provided writer
func (c *HTTPClient) DownloadToWriter(ctx context.Context, url string, writer io.Writer) (int64, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		}

		bytesWritten, err := c.doDownloadStream(ctx, url, writer)
		if err == nil {
			return bytesWritten, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		if isClientError(err) {
			break
		}
	}

	return 0, fmt.Errorf("failed after %d attempts: %w", c.retryAttempts+1, lastErr)
}

// Head performs a HEAD request to get metadata without downloading content
func (c *HTTPClient) Head(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	// Set default user agent
	if c.authProvider == nil || req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "downurl/1.0")
	}

	// Apply authentication if configured
	if c.authProvider != nil {
		if err := c.authProvider.ApplyAuth(req); err != nil {
			return nil, fmt.Errorf("failed to apply authentication: %w", err)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HEAD request failed: %w", err)
	}

	return resp, nil
}

// doDownload performs a single download attempt
func (c *HTTPClient) doDownload(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default user agent if no auth provider or no custom user agent
	if c.authProvider == nil || req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "downurl/1.0")
	}

	// Apply authentication if configured
	if c.authProvider != nil {
		if err := c.authProvider.ApplyAuth(req); err != nil {
			return nil, fmt.Errorf("failed to apply authentication: %w", err)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	// Check content length if provided
	if resp.ContentLength > c.maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d bytes)", resp.ContentLength, c.maxSize)
	}

	// Read response body with size limit
	limitedReader := io.LimitReader(resp.Body, c.maxSize)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check if we hit the limit
	if int64(len(data)) >= c.maxSize {
		return nil, fmt.Errorf("file exceeded maximum size limit of %d bytes", c.maxSize)
	}

	return data, nil
}

// doDownloadStream performs a single download attempt with streaming
func (c *HTTPClient) doDownloadStream(ctx context.Context, url string, writer io.Writer) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default user agent if no auth provider or no custom user agent
	if c.authProvider == nil || req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "downurl/1.0")
	}

	// Apply authentication if configured
	if c.authProvider != nil {
		if err := c.authProvider.ApplyAuth(req); err != nil {
			return 0, fmt.Errorf("failed to apply authentication: %w", err)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	// Check content length if provided
	if resp.ContentLength > 0 && resp.ContentLength > c.maxSize {
		return 0, fmt.Errorf("file too large: %d bytes (max: %d bytes)", resp.ContentLength, c.maxSize)
	}

	// Stream response body to writer with size limit
	limitedReader := io.LimitReader(resp.Body, c.maxSize)
	bytesWritten, err := io.Copy(writer, limitedReader)
	if err != nil {
		return bytesWritten, fmt.Errorf("failed to write response: %w", err)
	}

	// Check if we hit the limit
	if bytesWritten >= c.maxSize {
		return bytesWritten, fmt.Errorf("file exceeded maximum size limit of %d bytes", c.maxSize)
	}

	return bytesWritten, nil
}

// isClientError checks if the error is a 4xx client error
func isClientError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode >= 400 && httpErr.StatusCode < 500
	}
	return false
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Status     string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Status)
}
