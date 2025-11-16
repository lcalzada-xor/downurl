package storage

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Archiver handles tar.gz archive creation
type Archiver struct{}

// NewArchiver creates a new Archiver instance
func NewArchiver() *Archiver {
	return &Archiver{}
}

// CreateTarGz creates a tar.gz archive from a source directory
func (a *Archiver) CreateTarGz(sourceDir, destFile string) error {
	// Create destination file
	outFile, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk through source directory and add files to archive
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the archive file itself if it's in the source directory
		if path == destFile {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}

		// Update header name to be relative to source directory
		relPath, err := filepath.Rel(filepath.Dir(sourceDir), path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Normalize path separators for tar (always use forward slash)
		header.Name = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// If it's a file, write its content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}

			// Copy file content
			_, copyErr := io.Copy(tarWriter, file)
			file.Close() // Close immediately, not deferred

			if copyErr != nil {
				return fmt.Errorf("failed to write file content: %w", copyErr)
			}
		}

		return nil
	})
}
