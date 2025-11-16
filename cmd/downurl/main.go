package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/llvch/downurl/internal/config"
	"github.com/llvch/downurl/internal/downloader"
	"github.com/llvch/downurl/internal/filter"
	"github.com/llvch/downurl/internal/parser"
	"github.com/llvch/downurl/internal/processor"
	"github.com/llvch/downurl/internal/reporter"
	"github.com/llvch/downurl/internal/storage"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
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
		os.Exit(1)
	}

	// Run the application
	if err := run(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(cfg *config.Config) error {
	startTime := time.Now()

	log.Printf("Starting downurl...")
	log.Printf("Configuration:")
	log.Printf("  Input file: %s", cfg.InputFile)
	log.Printf("  Output dir: %s", cfg.OutputDir)
	log.Printf("  Workers: %d", cfg.Workers)
	log.Printf("  Timeout: %v", cfg.Timeout)
	log.Printf("  Retry attempts: %d", cfg.RetryAttempts)

	// Build authentication provider
	authProvider, err := cfg.BuildAuthProvider()
	if err != nil {
		return fmt.Errorf("failed to configure authentication: %w", err)
	}
	if authProvider != nil && authProvider.GetType() != "none" {
		log.Printf("  Authentication: %s", authProvider.GetType())
	}

	// Parse URLs from input file
	log.Printf("\n[1/5] Parsing URLs from file...")
	urls, err := parser.ParseURLsFromFile(cfg.InputFile)
	if err != nil {
		return fmt.Errorf("failed to parse URLs: %w", err)
	}
	log.Printf("Found %d URLs to download", len(urls))

	if len(urls) == 0 {
		return fmt.Errorf("no URLs found in input file")
	}

	// Initialize storage
	log.Printf("\n[2/5] Initializing storage...")
	fileStorage := storage.NewFileStorage(cfg.OutputDir)
	if err := fileStorage.Init(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	log.Printf("Storage initialized at: %s", cfg.OutputDir)

	// Initialize HTTP client with authentication
	httpClient := downloader.NewHTTPClientWithAuth(cfg.Timeout, cfg.RetryAttempts, authProvider)

	// Initialize downloader
	dl := downloader.New(httpClient, fileStorage, cfg.Workers)

	// Setup content filter if any filters are configured
	if cfg.FilterType != "" || cfg.ExcludeType != "" || cfg.FilterExt != "" ||
	   cfg.ExcludeExt != "" || cfg.MinSize > 0 || cfg.MaxSize > 0 || cfg.SkipEmpty {
		filterCfg := filter.FilterConfig{
			FilterType:  cfg.FilterType,
			ExcludeType: cfg.ExcludeType,
			FilterExt:   cfg.FilterExt,
			ExcludeExt:  cfg.ExcludeExt,
			MinSize:     cfg.MinSize,
			MaxSize:     cfg.MaxSize,
			SkipEmpty:   cfg.SkipEmpty,
		}
		contentFilter := filter.NewContentFilter(filterCfg)
		dl.SetFilter(contentFilter)
		log.Printf("  Content filtering: enabled")
	}

	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interruption signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Printf("\n\nReceived interrupt signal, shutting down gracefully...")
		cancel()
	}()

	// Download all files
	log.Printf("\n[3/5] Downloading files with %d workers...", cfg.Workers)
	results := dl.DownloadAll(ctx, urls)

	// Check if context was cancelled
	if ctx.Err() != nil {
		log.Printf("Download process was interrupted")
	}

	// Process downloaded files if any processing is enabled
	var proc *processor.Processor
	if cfg.ScanSecrets || cfg.ScanEndpoints || cfg.JSBeautify {
		log.Printf("\n[4/7] Processing downloaded files...")
		processorCfg := processor.Config{
			ScanSecrets:    cfg.ScanSecrets,
			ScanEndpoints:  cfg.ScanEndpoints,
			JSBeautify:     cfg.JSBeautify,
			SecretsEntropy: cfg.SecretsEntropy,
		}
		proc = processor.NewProcessor(processorCfg)

		// Process each result
		for _, result := range results {
			if err := proc.ProcessResult(result, cfg.OutputDir); err != nil {
				log.Printf("[WARN] Failed to process result for %s: %v", result.URL, err)
			}
		}
		log.Printf("Processing complete")

		// Save secrets if requested
		if cfg.ScanSecrets && cfg.SecretsOutput != "" {
			log.Printf("\n[5/7] Saving secrets...")
			secretsPath := filepath.Join(cfg.OutputDir, cfg.SecretsOutput)
			if err := proc.SaveSecrets(secretsPath); err != nil {
				log.Printf("[WARN] Failed to save secrets: %v", err)
			} else {
				log.Printf("Secrets saved to: %s", secretsPath)
			}
		}

		// Save endpoints if requested
		if cfg.ScanEndpoints && cfg.EndpointsOutput != "" {
			log.Printf("\n[6/7] Saving endpoints...")
			endpointsPath := filepath.Join(cfg.OutputDir, cfg.EndpointsOutput)
			if err := proc.SaveEndpoints(endpointsPath); err != nil {
				log.Printf("[WARN] Failed to save endpoints: %v", err)
			} else {
				log.Printf("Endpoints saved to: %s", endpointsPath)
			}
		}
	}

	// Generate output in requested format
	stepNum := 4
	if proc != nil {
		stepNum = 7
	}
	log.Printf("\n[%d/%d] Generating report...", stepNum, stepNum)

	// Generate output report
	var reportPath string
	if proc != nil && (cfg.OutputFormat == "json" || cfg.OutputFormat == "csv" || cfg.OutputFormat == "markdown") {
		// Use processor reporter for advanced formats
		outputPath := cfg.OutputFile
		if outputPath == "" {
			ext := ".txt"
			switch cfg.OutputFormat {
			case "json":
				ext = ".json"
			case "csv":
				ext = ".csv"
			case "markdown":
				ext = ".md"
			}
			outputPath = filepath.Join(cfg.OutputDir, "report"+ext)
		}

		// Generate based on format
		var err error
		switch cfg.OutputFormat {
		case "json":
			err = proc.GetReporter().GenerateJSON(outputPath, cfg.PrettyJSON)
		case "csv":
			err = proc.GetReporter().GenerateCSV(outputPath)
		case "markdown":
			err = proc.GetReporter().GenerateMarkdown(outputPath)
		}

		if err != nil {
			return fmt.Errorf("failed to generate %s output: %w", cfg.OutputFormat, err)
		}
		log.Printf("Report saved to: %s", outputPath)
		reportPath = outputPath
	} else {
		// Basic text report
		rep := reporter.New()
		rep.AddBatch(results)

		reportPath = filepath.Join(cfg.OutputDir, "report.txt")
		if err := rep.Generate(reportPath); err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}
		log.Printf("Report saved to: %s", reportPath)
	}

	// Create tar.gz archive
	finalStep := stepNum + 1
	log.Printf("\n[%d/%d] Creating tar.gz archive...", finalStep, finalStep)
	archiver := storage.NewArchiver()
	tarPath := filepath.Join(cfg.OutputDir, "output.tar.gz")
	if err := archiver.CreateTarGz(cfg.OutputDir, tarPath); err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	log.Printf("Archive created: %s", tarPath)

	// Print summary
	elapsed := time.Since(startTime)
	log.Printf("")
	log.Println(separator(60))
	log.Printf("Download Summary:")
	log.Printf("  Total URLs: %d", len(urls))

	successful := 0
	failed := 0
	totalDownloaded := 0
	totalErrors := 0

	for _, r := range results {
		if r.IsSuccess() {
			successful++
		} else {
			failed++
		}
		totalDownloaded += len(r.Downloaded)
		totalErrors += len(r.Errors)
	}

	log.Printf("  Successful: %d", successful)
	log.Printf("  Failed: %d", failed)
	log.Printf("  Total files downloaded: %d", totalDownloaded)
	log.Printf("  Total errors: %d", totalErrors)
	log.Printf("  Time elapsed: %v", elapsed)
	log.Println(separator(60))
	log.Printf("\nAll done! Output saved to: %s/", cfg.OutputDir)
	log.Printf("Report: %s", reportPath)
	log.Printf("Archive: %s", tarPath)

	return nil
}

func separator(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "="
	}
	return result
}
