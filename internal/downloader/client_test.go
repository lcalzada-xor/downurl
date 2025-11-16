package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClient_Download_Success(t *testing.T) {
	// Create test server
	expectedContent := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(expectedContent)
	}))
	defer server.Close()

	client := NewHTTPClient(5*time.Second, 2)
	ctx := context.Background()

	data, err := client.Download(ctx, server.URL)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	if string(data) != string(expectedContent) {
		t.Errorf("Download() data = %s, want %s", data, expectedContent)
	}
}

func TestHTTPClient_Download_404(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewHTTPClient(5*time.Second, 2)
	ctx := context.Background()

	_, err := client.Download(ctx, server.URL)
	if err == nil {
		t.Error("Download() expected error for 404 response")
	}

	// Error message should contain information about the failure
	if err.Error() == "" {
		t.Error("Download() error message is empty")
	}
}

func TestHTTPClient_Download_Timeout(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("delayed"))
	}))
	defer server.Close()

	client := NewHTTPClient(100*time.Millisecond, 0)
	ctx := context.Background()

	_, err := client.Download(ctx, server.URL)
	if err == nil {
		t.Error("Download() expected timeout error")
	}
}

func TestHTTPClient_Download_ContextCancellation(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Write([]byte("data"))
	}))
	defer server.Close()

	client := NewHTTPClient(5*time.Second, 2)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	_, err := client.Download(ctx, server.URL)
	if err == nil {
		t.Error("Download() expected error for cancelled context")
	}
}
