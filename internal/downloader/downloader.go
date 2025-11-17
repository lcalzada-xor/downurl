package downloader

import (
	"context"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/llvch/downurl/internal/filter"
	"github.com/llvch/downurl/internal/parser"
	"github.com/llvch/downurl/internal/ratelimit"
	"github.com/llvch/downurl/internal/storage"
	"github.com/llvch/downurl/pkg/models"
)

// Downloader orchestrates the download process with worker pool
type Downloader struct {
	client       *HTTPClient
	storage      *storage.FileStorage
	workers      int
	filter       *filter.ContentFilter
	skipHeadReq  bool
}

// New creates a new Downloader instance
func New(client *HTTPClient, storage *storage.FileStorage, workers int) *Downloader {
	return &Downloader{
		client:      client,
		storage:     storage,
		workers:     workers,
		skipHeadReq: false,
	}
}

// SetFilter sets the content filter for pre-download filtering
func (d *Downloader) SetFilter(f *filter.ContentFilter) {
	d.filter = f
}

// SetSkipHeadRequest sets whether to skip HEAD requests
func (d *Downloader) SetSkipHeadRequest(skip bool) {
	d.skipHeadReq = skip
}

// Job represents a download job
type Job struct {
	URL   string
	Index int
}

// ProgressCallback is a function that's called when progress is made
type ProgressCallback func(completed, total int)

// Result is an alias for models.DownloadResult for backward compatibility
type Result = models.DownloadResult

// DownloadAll downloads all URLs using a worker pool
func (d *Downloader) DownloadAll(ctx context.Context, urls []string) []*models.DownloadResult {
	return d.DownloadAllWithProgress(ctx, urls, nil)
}

// DownloadAllWithProgress downloads all URLs with progress callback
func (d *Downloader) DownloadAllWithProgress(ctx context.Context, urls []string, callback ProgressCallback) []*models.DownloadResult {
	jobs := make(chan Job, len(urls))
	results := make(chan models.DownloadResult, len(urls))

	var completed int32
	totalJobs := len(urls)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < d.workers; i++ {
		wg.Add(1)
		go d.workerWithCallback(ctx, &wg, jobs, results, &completed, totalJobs, callback)
	}

	// Send jobs to workers
	for i, url := range urls {
		jobs <- Job{URL: url, Index: i}
	}
	close(jobs)

	// Wait for all workers to finish and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	allResults := make([]*models.DownloadResult, 0, len(urls))
	for result := range results {
		res := result
		allResults = append(allResults, &res)
	}

	return allResults
}

// DownloadAllWithRateLimit downloads all URLs with rate limiting
func (d *Downloader) DownloadAllWithRateLimit(ctx context.Context, urls []string, limiter *ratelimit.Limiter, callback ProgressCallback) []*models.DownloadResult {
	jobs := make(chan Job, len(urls))
	results := make(chan models.DownloadResult, len(urls))

	var completed int32
	totalJobs := len(urls)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < d.workers; i++ {
		wg.Add(1)
		go d.workerWithRateLimit(ctx, &wg, jobs, results, limiter, &completed, totalJobs, callback)
	}

	// Send jobs to workers
	for i, url := range urls {
		jobs <- Job{URL: url, Index: i}
	}
	close(jobs)

	// Wait for all workers to finish and close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	allResults := make([]*models.DownloadResult, 0, len(urls))
	for result := range results {
		res := result
		allResults = append(allResults, &res)
	}

	return allResults
}

// worker processes download jobs from the jobs channel
func (d *Downloader) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- models.DownloadResult) {
	defer wg.Done()

	for job := range jobs {
		// Check if context was cancelled before processing
		if ctx.Err() != nil {
			// Create error result for cancelled job
			result := models.DownloadResult{
				URL:        job.URL,
				Host:       parser.HostnameFromURL(job.URL),
				Downloaded: []string{},
				Errors:     []string{"download cancelled by user"},
				Duration:   0,
			}

			// Try to send result, but don't block if context is done
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
			continue
		}

		result := d.processJob(ctx, job)

		// Send result with context awareness
		select {
		case results <- result:
		case <-ctx.Done():
			return
		}
	}
}

// workerWithCallback processes download jobs with progress callback
func (d *Downloader) workerWithCallback(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- models.DownloadResult, completed *int32, total int, callback ProgressCallback) {
	defer wg.Done()

	for job := range jobs {
		// Check if context was cancelled before processing
		if ctx.Err() != nil {
			result := models.DownloadResult{
				URL:        job.URL,
				Host:       parser.HostnameFromURL(job.URL),
				Downloaded: []string{},
				Errors:     []string{"download cancelled by user"},
				Duration:   0,
			}

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}

			// Update progress
			count := atomic.AddInt32(completed, 1)
			if callback != nil {
				callback(int(count), total)
			}
			continue
		}

		result := d.processJob(ctx, job)

		// Send result with context awareness
		select {
		case results <- result:
		case <-ctx.Done():
			return
		}

		// Update progress
		count := atomic.AddInt32(completed, 1)
		if callback != nil {
			callback(int(count), total)
		}
	}
}

// workerWithRateLimit processes download jobs with rate limiting
func (d *Downloader) workerWithRateLimit(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- models.DownloadResult, limiter *ratelimit.Limiter, completed *int32, total int, callback ProgressCallback) {
	defer wg.Done()

	for job := range jobs {
		// Check if context was cancelled
		if ctx.Err() != nil {
			result := models.DownloadResult{
				URL:        job.URL,
				Host:       parser.HostnameFromURL(job.URL),
				Downloaded: []string{},
				Errors:     []string{"download cancelled by user"},
				Duration:   0,
			}

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}

			count := atomic.AddInt32(completed, 1)
			if callback != nil {
				callback(int(count), total)
			}
			continue
		}

		// Wait for rate limiter
		if err := limiter.Wait(ctx); err != nil {
			// Rate limiter cancelled by context
			result := models.DownloadResult{
				URL:        job.URL,
				Host:       parser.HostnameFromURL(job.URL),
				Downloaded: []string{},
				Errors:     []string{"rate limiter cancelled"},
				Duration:   0,
			}

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}

			count := atomic.AddInt32(completed, 1)
			if callback != nil {
				callback(int(count), total)
			}
			continue
		}

		result := d.processJob(ctx, job)

		// Send result
		select {
		case results <- result:
		case <-ctx.Done():
			return
		}

		// Update progress
		count := atomic.AddInt32(completed, 1)
		if callback != nil {
			callback(int(count), total)
		}
	}
}

// processJob downloads a single URL and saves it to disk
func (d *Downloader) processJob(ctx context.Context, job Job) models.DownloadResult {
	start := time.Now()
	result := models.DownloadResult{
		URL:        job.URL,
		Host:       parser.HostnameFromURL(job.URL),
		Downloaded: []string{},
		Errors:     []string{},
	}

	// Pre-download filtering with HEAD request (if filter is set and HEAD not skipped)
	if d.filter != nil && !d.skipHeadReq {
		shouldDownload, reason := d.checkShouldDownload(ctx, job.URL)
		if !shouldDownload {
			result.Errors = append(result.Errors, "skipped: "+reason)
			result.Duration = time.Since(start)
			log.Printf("[SKIP] %s: %s", job.URL, reason)
			return result
		}
	}

	// Generate filename
	filename := parser.FilenameFromURL(job.URL)

	// Download and save using streaming (no memory buffering)
	filepath, bytesWritten, err := d.downloadAndSaveStream(ctx, job.URL, result.Host, filename)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		result.Duration = time.Since(start)
		log.Printf("[ERROR] Failed to download %s: %v", job.URL, err)
		return result
	}

	result.Downloaded = append(result.Downloaded, filepath)
	result.Duration = time.Since(start)
	log.Printf("[OK] Downloaded %s -> %s (%d bytes, %v)", job.URL, filepath, bytesWritten, result.Duration)

	return result
}

// checkShouldDownload performs a HEAD request and checks if the file should be downloaded
func (d *Downloader) checkShouldDownload(ctx context.Context, url string) (bool, string) {
	resp, err := d.client.Head(ctx, url)
	if err != nil {
		// If HEAD fails, we still want to try downloading (some servers don't support HEAD)
		log.Printf("[WARN] HEAD request failed for %s: %v, will attempt download", url, err)
		return true, ""
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, "HTTP status: " + resp.Status
	}

	// Get content type and length
	contentType := resp.Header.Get("Content-Type")
	contentLength := resp.ContentLength

	// Apply filter
	return d.filter.ShouldDownload(url, contentType, contentLength)
}

// downloadAndSaveStream downloads a URL and saves it directly to disk using streaming
func (d *Downloader) downloadAndSaveStream(ctx context.Context, url, host, filename string) (string, int64, error) {
	// Create a pipe to connect download and storage
	pr, pw := io.Pipe()

	var downloadErr error
	var bytesDownloaded int64

	// Start downloading in a goroutine
	go func() {
		defer pw.Close()
		bytes, err := d.client.DownloadToWriter(ctx, url, pw)
		bytesDownloaded = bytes
		downloadErr = err
		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	// Extract URL path for storage strategy
	urlPath := parser.PathFromURL(url)

	// Save from the pipe reader
	filepath, bytesWritten, err := d.storage.SaveFileFromReader(host, urlPath, filename, pr)

	// Check if download had an error
	if downloadErr != nil {
		return "", bytesDownloaded, downloadErr
	}

	if err != nil {
		return "", bytesWritten, err
	}

	return filepath, bytesWritten, nil
}
