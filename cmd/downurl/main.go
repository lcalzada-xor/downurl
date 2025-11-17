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

	"github.com/lcalzada-xor/downurl/internal/config"
	"github.com/lcalzada-xor/downurl/internal/downloader"
	"github.com/lcalzada-xor/downurl/internal/filter"
	"github.com/lcalzada-xor/downurl/internal/parser"
	"github.com/lcalzada-xor/downurl/internal/processor"
	"github.com/lcalzada-xor/downurl/internal/ratelimit"
	"github.com/lcalzada-xor/downurl/internal/reporter"
	"github.com/lcalzada-xor/downurl/internal/storage"
	"github.com/lcalzada-xor/downurl/internal/ui"
	"github.com/lcalzada-xor/downurl/internal/watcher"
	"github.com/lcalzada-xor/downurl/pkg/models"
)

func main() {
	// Load configuration from flags
	cfg := config.Load()

	// Load config file and apply to config
	configFile, _ := config.LoadConfigFile()
	if configFile != nil {
		configFile.ApplyToConfig(cfg)
	}

	// Save config if requested
	if cfg.SaveConfig != "" {
		if err := config.SaveConfigFile(cfg, cfg.SaveConfig); err != nil {
			fmt.Fprintln(os.Stderr, ui.WrapPermissionError(cfg.SaveConfig, err))
			os.Exit(1)
		}
		ui.Success(fmt.Sprintf("Configuration saved to %s", cfg.SaveConfig))
		return
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		if err == config.ErrMissingInputFile {
			// Special handling for missing input file
			if cfg.SingleURL == "" && !parser.IsStdinAvailable() {
				fmt.Fprintln(os.Stderr, ui.WrapNoURLsError())
				ui.PrintUsageHint()
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Usage: downurl --input <urls.txt> [options]\n")
			fmt.Fprintf(os.Stderr, "\nFor help: downurl --help\n")
			os.Exit(1)
		}
	}

	// Run the application
	if err := run(cfg); err != nil {
		// Print friendly error
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	return runDownload(cfg, context.Background())
}

func runDownload(cfg *config.Config, parentCtx context.Context) error {
	startTime := time.Now()

	if !cfg.Quiet {
		ui.Info("Starting downurl...")
	}

	// Parse URLs based on input mode
	var urls []string
	var err error

	if cfg.SingleURL != "" {
		// Single URL mode
		if !cfg.Quiet {
			log.Printf("[1/5] Processing single URL...")
		}
		validURL, err := parser.ParseSingleURL(cfg.SingleURL)
		if err != nil {
			return ui.WrapInvalidURL(cfg.SingleURL, 1, err)
		}
		urls = []string{validURL}
	} else if cfg.InputFile == "" && parser.IsStdinAvailable() {
		// Stdin mode
		if !cfg.Quiet {
			log.Printf("[1/5] Reading URLs from stdin...")
		}
		urls, err = parser.ParseURLsFromStdin()
		if err != nil {
			return fmt.Errorf("failed to parse URLs from stdin: %w", err)
		}
	} else {
		// File mode
		if !cfg.Quiet {
			log.Printf("[1/5] Parsing URLs from file: %s", cfg.InputFile)
		}
		urls, err = parser.ParseURLsFromFile(cfg.InputFile)
		if err != nil {
			if os.IsNotExist(err) {
				return ui.WrapFileNotFound(cfg.InputFile, err)
			}
			return fmt.Errorf("failed to parse URLs: %w", err)
		}
	}

	// Validate we have URLs
	if len(urls) == 0 {
		return ui.WrapNoURLsError()
	}

	if !cfg.Quiet {
		ui.Success(fmt.Sprintf("Found %d URLs to download", len(urls)))
	}

	// Configuration summary
	if !cfg.Quiet {
		log.Printf("\nConfiguration:")
		log.Printf("  Output dir: %s", cfg.OutputDir)
		log.Printf("  Workers: %d", cfg.Workers)
		log.Printf("  Timeout: %v", cfg.Timeout)
		log.Printf("  Retry attempts: %d", cfg.RetryAttempts)
	}

	// Build authentication provider
	authProvider, err := cfg.BuildAuthProvider()
	if err != nil {
		return fmt.Errorf("failed to configure authentication: %w", err)
	}
	if !cfg.Quiet && authProvider != nil && authProvider.GetType() != "none" {
		log.Printf("  Authentication: %s", authProvider.GetType())
	}

	// Initialize storage
	if !cfg.Quiet {
		log.Printf("\n[2/5] Initializing storage...")
	}
	fileStorage := storage.NewFileStorage(cfg.OutputDir, cfg.StorageMode)
	if err := fileStorage.Init(); err != nil {
		return ui.WrapPermissionError(cfg.OutputDir, err)
	}
	if !cfg.Quiet {
		ui.Success(fmt.Sprintf("Storage initialized at: %s", cfg.OutputDir))
		log.Printf("  Storage mode: %s", cfg.StorageMode)
	}

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
		if !cfg.Quiet {
			log.Printf("  Content filtering: enabled")
		}
	}

	// Setup rate limiter if configured
	var limiter *ratelimit.Limiter
	if cfg.RateLimit != "" {
		limiter, err = ratelimit.ParseRateLimit(cfg.RateLimit)
		if err != nil {
			return fmt.Errorf("invalid rate limit: %w", err)
		}
		if !cfg.Quiet {
			log.Printf("  Rate limiting: %s", cfg.RateLimit)
		}
	}

	// Setup context with cancellation for graceful shutdown
	// Use parentCtx if provided (for watch/schedule), otherwise create new
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// Handle interruption signals only if this is the top-level call
	if parentCtx == context.Background() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			if !cfg.Quiet {
				log.Printf("\n\nReceived interrupt signal, shutting down gracefully...")
			}
			cancel()
		}()
	}

	// Download all files
	if !cfg.Quiet {
		log.Printf("\n[3/5] Downloading files with %d workers...", cfg.Workers)
	}

	// Create progress bar if not disabled
	var pb *ui.ProgressBar
	if !cfg.Quiet && !cfg.NoProgress {
		pb = ui.NewProgressBar(len(urls), true)
		fmt.Print(pb.Render())
	}

	// Download with rate limiting if configured
	var results []*downloader.Result
	if limiter != nil {
		// Download with rate limiting
		results = dl.DownloadAllWithRateLimit(ctx, urls, limiter, func(completed, total int) {
			if pb != nil {
				pb.Update(completed)
				fmt.Print(pb.Render())
			}
		})
	} else {
		// Standard download with progress callback
		results = dl.DownloadAllWithProgress(ctx, urls, func(completed, total int) {
			if pb != nil {
				pb.Update(completed)
				fmt.Print(pb.Render())
			}
		})
	}

	// Finish progress bar
	if pb != nil {
		pb.Finish()
	}

	// Check if context was cancelled
	if ctx.Err() != nil && !cfg.Quiet {
		ui.Warning("Download process was interrupted")
	}

	// Process downloaded files if any processing is enabled
	var proc *processor.Processor
	if cfg.ScanSecrets || cfg.ScanEndpoints || cfg.JSBeautify {
		if !cfg.Quiet {
			log.Printf("\n[4/7] Processing downloaded files...")
		}
		processorCfg := processor.Config{
			ScanSecrets:    cfg.ScanSecrets,
			ScanEndpoints:  cfg.ScanEndpoints,
			JSBeautify:     cfg.JSBeautify,
			SecretsEntropy: cfg.SecretsEntropy,
		}
		proc = processor.NewProcessor(processorCfg)

		// Process each result
		for _, result := range results {
			if err := proc.ProcessResult(*result, cfg.OutputDir); err != nil {
				if !cfg.Quiet {
					log.Printf("[WARN] Failed to process result for %s: %v", result.URL, err)
				}
			}
		}
		if !cfg.Quiet {
			ui.Success("Processing complete")
		}

		// Save secrets if requested
		if cfg.ScanSecrets && cfg.SecretsOutput != "" {
			if !cfg.Quiet {
				log.Printf("\n[5/7] Saving secrets...")
			}
			secretsPath := filepath.Join(cfg.OutputDir, cfg.SecretsOutput)
			if err := proc.SaveSecrets(secretsPath); err != nil {
				if !cfg.Quiet {
					log.Printf("[WARN] Failed to save secrets: %v", err)
				}
			} else {
				if !cfg.Quiet {
					ui.Success(fmt.Sprintf("Secrets saved to: %s", secretsPath))
				}
			}
		}

		// Save endpoints if requested
		if cfg.ScanEndpoints && cfg.EndpointsOutput != "" {
			if !cfg.Quiet {
				log.Printf("\n[6/7] Saving endpoints...")
			}
			endpointsPath := filepath.Join(cfg.OutputDir, cfg.EndpointsOutput)
			if err := proc.SaveEndpoints(endpointsPath); err != nil {
				if !cfg.Quiet {
					log.Printf("[WARN] Failed to save endpoints: %v", err)
				}
			} else {
				if !cfg.Quiet {
					ui.Success(fmt.Sprintf("Endpoints saved to: %s", endpointsPath))
				}
			}
		}
	}

	// Generate output in requested format
	stepNum := 4
	if proc != nil {
		stepNum = 7
	}
	if !cfg.Quiet {
		log.Printf("\n[%d/%d] Generating report...", stepNum, stepNum)
	}

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
		if !cfg.Quiet {
			ui.Success(fmt.Sprintf("Report saved to: %s", outputPath))
		}
		reportPath = outputPath
	} else {
		// Basic text report
		rep := reporter.New()
		// Convert []*Result to []Result
		plainResults := make([]models.DownloadResult, len(results))
		for i, r := range results {
			plainResults[i] = *r
		}
		rep.AddBatch(plainResults)

		reportPath = filepath.Join(cfg.OutputDir, "report.txt")
		if err := rep.Generate(reportPath); err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}
		if !cfg.Quiet {
			ui.Success(fmt.Sprintf("Report saved to: %s", reportPath))
		}
	}

	// Create tar.gz archive
	finalStep := stepNum + 1
	if !cfg.Quiet {
		log.Printf("\n[%d/%d] Creating tar.gz archive...", finalStep, finalStep)
	}
	archiver := storage.NewArchiver()
	tarPath := filepath.Join(cfg.OutputDir, "output.tar.gz")
	if err := archiver.CreateTarGz(cfg.OutputDir, tarPath); err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	if !cfg.Quiet {
		ui.Success(fmt.Sprintf("Archive created: %s", tarPath))
	}

	// Print enhanced summary
	elapsed := time.Since(startTime)
	if !cfg.Quiet {
		fmt.Println()
		// Convert []*Result to []Result for UI
		plainResults := make([]models.DownloadResult, len(results))
		for i, r := range results {
			plainResults[i] = *r
		}

		// Show results table
		table := ui.NewResultsTable(plainResults)
		fmt.Println(table.Render())

		// Show detailed summary
		summary := ui.RenderSummary(plainResults, elapsed, cfg.OutputDir)
		fmt.Print(summary)

		fmt.Printf("\nReport: %s\n", reportPath)
		fmt.Printf("Archive: %s\n", tarPath)
	}

	// Watch mode - keep running and watch for file changes
	// Only start watch/schedule on top-level run (not in recursive calls)
	if cfg.Watch && parentCtx == context.Background() {
		if cfg.InputFile == "" {
			return fmt.Errorf("--watch requires an input file (--input)")
		}
		fw := watcher.NewFileWatcher(cfg.InputFile, 5*time.Second, func() {
			log.Println("\n" + separator(60))
			log.Println("File changed, re-running download...")
			log.Println(separator(60))
			// Re-run with same context to avoid goroutine leak
			if err := runDownload(cfg, ctx); err != nil {
				log.Printf("Error during re-run: %v", err)
			}
		})
		return fw.Start(ctx)
	}

	// Schedule mode - run periodically
	// Only start watch/schedule on top-level run (not in recursive calls)
	if cfg.Schedule != "" && parentCtx == context.Background() {
		scheduler := watcher.NewScheduler(cfg.Schedule, func() error {
			log.Println("\n" + separator(60))
			log.Println("Running scheduled download...")
			log.Println(separator(60))
			// Use parent context to avoid creating nested contexts
			return runDownload(cfg, ctx)
		})
		return scheduler.Start(ctx)
	}

	return nil
}

func separator(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "="
	}
	return result
}
