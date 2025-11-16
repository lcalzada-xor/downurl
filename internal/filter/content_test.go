package filter

import (
	"testing"
)

func TestContentFilter_ShouldDownload_ContentType(t *testing.T) {
	cfg := FilterConfig{
		FilterType: "text/javascript,application/json",
	}
	filter := NewContentFilter(cfg)

	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{
			name:        "allowed javascript",
			contentType: "text/javascript",
			want:        true,
		},
		{
			name:        "allowed json",
			contentType: "application/json",
			want:        true,
		},
		{
			name:        "blocked html",
			contentType: "text/html",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := filter.ShouldDownload("http://example.com/file", tt.contentType, 1000)
			if got != tt.want {
				t.Errorf("ShouldDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentFilter_ShouldDownload_Extension(t *testing.T) {
	cfg := FilterConfig{
		FilterExt: ".js,.json",
	}
	filter := NewContentFilter(cfg)

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "allowed .js",
			url:  "http://example.com/app.js",
			want: true,
		},
		{
			name: "allowed .json",
			url:  "http://example.com/data.json",
			want: true,
		},
		{
			name: "blocked .html",
			url:  "http://example.com/index.html",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := filter.ShouldDownload(tt.url, "", 1000)
			if got != tt.want {
				t.Errorf("ShouldDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentFilter_ShouldDownload_Size(t *testing.T) {
	cfg := FilterConfig{
		MinSize: 100,
		MaxSize: 10000,
	}
	filter := NewContentFilter(cfg)

	tests := []struct {
		name   string
		size   int64
		want   bool
	}{
		{
			name: "too small",
			size: 50,
			want: false,
		},
		{
			name: "just right",
			size: 500,
			want: true,
		},
		{
			name: "too large",
			size: 20000,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := filter.ShouldDownload("http://example.com/file", "", tt.size)
			if got != tt.want {
				t.Errorf("ShouldDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentFilter_ShouldDownload_SkipEmpty(t *testing.T) {
	cfg := FilterConfig{
		SkipEmpty: true,
	}
	filter := NewContentFilter(cfg)

	got, reason := filter.ShouldDownload("http://example.com/file", "", 0)
	if got {
		t.Error("ShouldDownload() should reject empty file")
	}

	if reason != "file is empty" {
		t.Errorf("Reason = %s, want 'file is empty'", reason)
	}
}

func TestContentFilter_WildcardMatch(t *testing.T) {
	cfg := FilterConfig{
		ExcludeType: "image/*,video/*",
	}
	filter := NewContentFilter(cfg)

	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{
			name:        "blocked image/png",
			contentType: "image/png",
			want:        false,
		},
		{
			name:        "blocked image/jpeg",
			contentType: "image/jpeg",
			want:        false,
		},
		{
			name:        "blocked video/mp4",
			contentType: "video/mp4",
			want:        false,
		},
		{
			name:        "allowed text/html",
			contentType: "text/html",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := filter.ShouldDownload("http://example.com/file", tt.contentType, 1000)
			if got != tt.want {
				t.Errorf("ShouldDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsJavaScript(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{
			name:        "text/javascript",
			contentType: "text/javascript",
			want:        true,
		},
		{
			name:        "application/javascript",
			contentType: "application/javascript",
			want:        true,
		},
		{
			name:        "application/x-javascript",
			contentType: "application/x-javascript",
			want:        true,
		},
		{
			name:        "text/html",
			contentType: "text/html",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsJavaScript(tt.contentType)
			if got != tt.want {
				t.Errorf("IsJavaScript() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyContent(t *testing.T) {
	tests := []struct {
		contentType string
		want        string
	}{
		{"text/javascript", "JavaScript"},
		{"application/json", "JSON"},
		{"text/html", "HTML"},
		{"text/css", "CSS"},
		{"image/png", "Image"},
		{"video/mp4", "Video"},
		{"application/pdf", "PDF"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := ClassifyContent(tt.contentType)
			if got != tt.want {
				t.Errorf("ClassifyContent() = %s, want %s", got, tt.want)
			}
		})
	}
}
