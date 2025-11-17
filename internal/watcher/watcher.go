package watcher

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// FileWatcher watches a file for changes
type FileWatcher struct {
	path     string
	interval time.Duration
	lastHash []byte
	onChange func()
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(path string, interval time.Duration, onChange func()) *FileWatcher {
	return &FileWatcher{
		path:     path,
		interval: interval,
		onChange: onChange,
	}
}

// Start starts watching the file
func (fw *FileWatcher) Start(ctx context.Context) error {
	// Get initial hash
	hash, err := fw.getFileHash()
	if err != nil {
		return fmt.Errorf("failed to read initial file: %w", err)
	}
	fw.lastHash = hash

	log.Printf("👀 Watching %s for changes (checking every %v)...", fw.path, fw.interval)
	log.Println("Press Ctrl+C to stop watching...")

	ticker := time.NewTicker(fw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("\nStopped watching file")
			return nil
		case <-ticker.C:
			if changed, err := fw.checkForChanges(); err != nil {
				log.Printf("Error checking file: %v", err)
			} else if changed {
				timestamp := time.Now().Format("15:04:05")
				log.Printf("\n[%s] File changed, triggering download...", timestamp)
				fw.onChange()
			}
		}
	}
}

// checkForChanges checks if file has changed
func (fw *FileWatcher) checkForChanges() (bool, error) {
	hash, err := fw.getFileHash()
	if err != nil {
		return false, err
	}

	if string(hash) != string(fw.lastHash) {
		fw.lastHash = hash
		return true, nil
	}

	return false, nil
}

// getFileHash returns SHA256 hash of file
func (fw *FileWatcher) getFileHash() ([]byte, error) {
	f, err := os.Open(fw.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// Scheduler handles scheduled downloads
type Scheduler struct {
	schedule string // cron expression
	runFunc  func() error
}

// NewScheduler creates a new scheduler
func NewScheduler(schedule string, runFunc func() error) *Scheduler {
	return &Scheduler{
		schedule: schedule,
		runFunc:  runFunc,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	log.Printf("📅 Scheduled download: %s", s.schedule)
	log.Println("Note: Full cron support requires external scheduler")
	log.Println("Consider using systemd timer or crontab:")
	log.Printf("  */5 * * * * /path/to/downurl -i urls.txt\n")

	// For now, we'll implement simple interval-based scheduling
	interval, err := s.parseSimpleSchedule()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately
	log.Println("Running initial download...")
	if err := s.runFunc(); err != nil {
		log.Printf("Error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("\nScheduler stopped")
			return nil
		case <-ticker.C:
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			log.Printf("\n[%s] Running scheduled download...", timestamp)
			if err := s.runFunc(); err != nil {
				log.Printf("Error: %v", err)
			}
		}
	}
}

// parseSimpleSchedule parses simple schedule formats
func (s *Scheduler) parseSimpleSchedule() (time.Duration, error) {
	// Support simple formats like "5m", "1h", "30s"
	if d, err := time.ParseDuration(s.schedule); err == nil {
		return d, nil
	}

	// TODO: Add full cron expression support
	return 0, fmt.Errorf("invalid schedule format: %s (use duration like 5m, 1h, etc.)", s.schedule)
}
