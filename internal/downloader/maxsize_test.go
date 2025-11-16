package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPClient_Download_MaxSizeExceeded(t *testing.T) {
	// Create test server that returns large content
	largeContent := strings.Repeat("A", int(MaxDownloadSize)+1000)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeContent))
	}))
	defer server.Close()

	client := NewHTTPClient(5*time.Second, 0)
	ctx := context.Background()

	_, err := client.Download(ctx, server.URL)
	if err == nil {
		t.Error("Download() expected error for file exceeding max size")
	}

	if !strings.Contains(err.Error(), "maximum size") {
		t.Errorf("Error should mention maximum size, got: %v", err)
	}
}

func TestHTTPClient_Download_MaxSizeContentLength(t *testing.T) {
	// Create test server that advertises large content-length
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "999999999999") // 999GB
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(5*time.Second, 0)
	ctx := context.Background()

	_, err := client.Download(ctx, server.URL)
	if err == nil {
		t.Error("Download() expected error for large Content-Length")
	}

	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("Error should mention 'too large', got: %v", err)
	}
}

func TestHTTPClient_Download_NormalSize(t *testing.T) {
	// Create test server with normal-sized content
	normalContent := strings.Repeat("B", 1000) // 1KB
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(normalContent))
	}))
	defer server.Close()

	client := NewHTTPClient(5*time.Second, 0)
	ctx := context.Background()

	data, err := client.Download(ctx, server.URL)
	if err != nil {
		t.Fatalf("Download() error = %v, should succeed for normal-sized file", err)
	}

	if len(data) != 1000 {
		t.Errorf("Download() data length = %d, want 1000", len(data))
	}
}

func TestHTTPClient_Download_ExactlyMaxSize(t *testing.T) {
	// Create test server with content exactly at max size
	content := strings.Repeat("C", int(MaxDownloadSize)-1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	client := NewHTTPClient(5*time.Second, 0)
	ctx := context.Background()

	data, err := client.Download(ctx, server.URL)
	if err != nil {
		t.Fatalf("Download() error = %v, should succeed at max size-1", err)
	}

	if len(data) != int(MaxDownloadSize)-1 {
		t.Errorf("Download() data length = %d, want %d", len(data), int(MaxDownloadSize)-1)
	}
}
