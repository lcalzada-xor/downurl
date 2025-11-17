# Downurl

A high-performance, concurrent file downloader written in Go. Download files from URLs with enhanced UI, multiple input modes, rate limiting, and automatic archiving.

[![Version](https://img.shields.io/badge/version-1.1.0-blue.svg)](https://github.com/llvch/downurl/releases)
[![Go Version](https://img.shields.io/badge/go-1.24.9-00ADD8.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## ✨ Features

### Core Features
- **Concurrent Downloads**: Configurable worker pool for parallel downloads
- **Multiple Input Modes**: stdin, single URL, or file-based input
- **Animated Progress Bar**: Real-time progress with speed and ETA
- **Color-Coded Output**: Visual feedback with colored status messages
- **Smart Filename Generation**: Safe filename extraction from URLs with hash fallback
- **Storage Modes**: 5 ways to organize files (flat, path, host, type, dated)
- **Comprehensive Reporting**: Detailed reports with statistics and error tracking
- **Automatic Archiving**: Creates tar.gz archives of all downloaded content

### Advanced Features
- **Rate Limiting**: Token bucket algorithm with flexible rate configuration
- **Watch Mode**: Monitor input file for changes and auto-download
- **Schedule Mode**: Periodic downloads at specified intervals
- **Configuration File**: INI-style config with environment variable expansion
- **Authentication**: Bearer, Basic, Custom headers, Cookies
- **Content Filtering**: By type, extension, size
- **Security Scanning**: Secret detection and endpoint discovery
- **Friendly Errors**: Helpful error messages with suggestions

### Reliability
- **Retry Logic**: Automatic retry with exponential backoff
- **Graceful Shutdown**: Handles interruption signals cleanly
- **Context-Aware**: Proper context cancellation throughout
- **Thread-Safe**: No race conditions, verified with `-race` flag

## 🚀 Quick Start

### Installation

**Download Pre-built Binary**:
```bash
# Linux (AMD64)
curl -LO https://github.com/llvch/downurl/releases/latest/download/downurl-linux-amd64.tar.gz
tar -xzf downurl-linux-amd64.tar.gz
sudo mv downurl-linux-amd64 /usr/local/bin/downurl

# macOS (Apple Silicon)
curl -LO https://github.com/llvch/downurl/releases/latest/download/downurl-darwin-arm64.tar.gz
tar -xzf downurl-darwin-arm64.tar.gz
sudo mv downurl-darwin-arm64 /usr/local/bin/downurl
```

**From Source**:
```bash
git clone https://github.com/llvch/downurl.git
cd downurl
go build -o downurl cmd/downurl/main.go
```

### Basic Usage

**Single URL** (v1.1.0+):
```bash
downurl "https://example.com/script.js"
```

**From stdin** (v1.1.0+):
```bash
cat urls.txt | downurl
echo "https://example.com/file.js" | downurl
```

**From file**:
```bash
downurl -input urls.txt
```

**With progress bar and colors** (v1.1.0+):
```
⣾ Downloading... [==================>     ] 75% | 3.2 MB/s | ETA: 5s

✓ Successfully downloaded 8 files
✗ Failed: 2 files
⚠ Total time: 12.5s
```

## 📖 Documentation

- **[Getting Started Guide](docs/user-guides/GETTING_STARTED.md)** - Installation and first steps
- **[Configuration Guide](docs/user-guides/CONFIGURATION.md)** - Config files and settings
- **[Release Notes v1.1.0](docs/RELEASE_NOTES_v1.1.0.md)** - What's new
- **[Documentation Index](docs/DOCUMENTATION_INDEX.md)** - Complete documentation

## 💡 Usage Examples

### Multiple Input Modes (v1.1.0+)

```bash
# Single URL (no file needed)
downurl "https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.21/lodash.min.js"

# Pipe from stdin
cat urls.txt | downurl
curl -s https://api.example.com/urls | jq -r '.urls[]' | downurl

# Traditional file mode
downurl -input urls.txt
```

### Storage Organization

```bash
# Flat: All files in one directory
downurl -input urls.txt --mode flat

# Path: Replicate URL structure
downurl -input urls.txt --mode path

# Host: Group by hostname (great for recon)
downurl -input urls.txt --mode host

# Type: Organize by extension
downurl -input urls.txt --mode type

# Dated: Group by download date
downurl -input urls.txt --mode dated
```

### Rate Limiting (v1.1.0+)

```bash
# 10 requests per second
downurl -input urls.txt --rate-limit "10/second"

# 100 requests per minute
downurl -input urls.txt --rate-limit "100/minute"

# 1000 requests per hour
downurl -input urls.txt --rate-limit "1000/hour"
```

### Watch & Schedule (v1.1.0+)

```bash
# Watch mode: Auto-download when file changes
downurl -input urls.txt --watch

# Schedule mode: Download every 5 minutes
downurl -input urls.txt --schedule "5m"

# Every hour
downurl -input urls.txt --schedule "1h"
```

### Configuration File (v1.1.0+)

```bash
# Create config file
cat > .downurlrc <<EOF
[defaults]
workers = 20
timeout = 30s
mode = host

[ratelimit]
default = 10/minute
EOF

# Run with config (auto-discovered)
downurl -input urls.txt

# Save current settings to config
downurl -input urls.txt -workers 30 --save-config my-config.ini
```

### Authentication

```bash
# Bearer token
downurl -input urls.txt --auth-bearer "your_token_here"

# Basic auth
downurl -input urls.txt --auth-basic "username:password"

# Custom headers
cat > headers.txt <<EOF
X-API-Key: your-api-key
Authorization: Bearer token123
EOF
downurl -input urls.txt --headers-file headers.txt

# Cookies
downurl -input urls.txt --cookie "session=abc123; token=xyz789"
```

### Content Filtering

```bash
# Filter by extension
downurl -input urls.txt --filter-ext "js,json,xml"

# Exclude extensions
downurl -input urls.txt --exclude-ext "min.js,bundle.js"

# Size limits
downurl -input urls.txt --min-size 1KB --max-size 10MB

# Content type
downurl -input urls.txt --filter-type "application/javascript"
```

### Security Research

```bash
# Secret scanning
downurl -input urls.txt --scan-secrets --secrets-output secrets.json

# Endpoint discovery
downurl -input js_files.txt --scan-endpoints --endpoints-output endpoints.json

# JavaScript beautification
downurl -input urls.txt --js-beautify

# Extract strings
downurl -input urls.txt --extract-strings --strings-min-length 10
```

### High Performance

```bash
# 50 concurrent workers
downurl -input urls.txt -workers 50

# Aggressive timeout
downurl -input urls.txt -timeout 10s -retry 5

# Quiet mode (no UI)
downurl -input urls.txt --quiet
```

## 🎯 Use Cases

### Bug Bounty / Security Research
```bash
# Monitor targets continuously with rate limiting
downurl -input recon_urls.txt --watch --mode host --rate-limit "5/second"

# Quick JS analysis from stdin
cat discovered_js.txt | downurl --mode host --scan-secrets
```

### Web Archiving
```bash
# Schedule periodic archiving
downurl -input archive_urls.txt --schedule "1h" --mode dated
```

### API Data Collection
```bash
# Rate-limited API scraping with auth
downurl -input api_endpoints.txt \
        --auth-bearer "$TOKEN" \
        --rate-limit "10/minute" \
        --mode path
```

### CDN Mirroring
```bash
# Mirror with high concurrency
cat cdn_urls.txt | downurl --mode path -workers 50
```

## 🔧 Command-Line Reference

### Essential Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-input` | Input file with URLs | (stdin) | `-input urls.txt` |
| `-output` | Output directory | `output` | `-output downloads` |
| `-workers` | Concurrent workers | `10` | `-workers 20` |
| `-timeout` | Request timeout | `15s` | `-timeout 30s` |
| `--mode` | Storage mode | `flat` | `--mode host` |

### Input Modes (v1.1.0+)

| Mode | Description | Example |
|------|-------------|---------|
| Single URL | Quick download | `downurl "https://example.com/file.js"` |
| Stdin | Pipe URLs | `cat urls.txt \| downurl` |
| File | Traditional | `downurl -input urls.txt` |

### Storage Modes

| Mode | Description | Structure |
|------|-------------|-----------|
| `flat` | All files in one dir | `output/file.js` |
| `path` | Replicate URL path | `output/api/v1/data.json` |
| `host` | Group by hostname | `output/example.com/file.js` |
| `type` | By file extension | `output/js/file.js` |
| `dated` | By download date | `output/2025-11-17/file.js` |

### New Flags (v1.1.0)

| Flag | Description | Example |
|------|-------------|---------|
| `--rate-limit` | Rate limit | `--rate-limit "10/second"` |
| `--watch` | Monitor file changes | `--watch` |
| `--schedule` | Periodic downloads | `--schedule "5m"` |
| `--config` | Config file path | `--config .downurlrc` |
| `--save-config` | Export config | `--save-config my.ini` |
| `--quiet` | Suppress output | `--quiet` |
| `--no-progress` | Disable progress bar | `--no-progress` |

### Authentication

| Flag | Description | Example |
|------|-------------|---------|
| `--auth-bearer` | Bearer token | `--auth-bearer "token123"` |
| `--auth-basic` | Basic auth | `--auth-basic "user:pass"` |
| `--auth-header` | Custom auth | `--auth-header "X-Key value"` |
| `--headers-file` | Headers from file | `--headers-file headers.txt` |
| `--cookie` | Cookie string | `--cookie "session=abc"` |
| `--cookies-file` | Cookies from file | `--cookies-file cookies.txt` |

### Filtering

| Flag | Description | Example |
|------|-------------|---------|
| `--filter-ext` | Include extensions | `--filter-ext "js,json"` |
| `--exclude-ext` | Exclude extensions | `--exclude-ext "min.js"` |
| `--filter-type` | Include content types | `--filter-type "application/json"` |
| `--exclude-type` | Exclude content types | `--exclude-type "image/png"` |
| `--min-size` | Minimum file size | `--min-size 1KB` |
| `--max-size` | Maximum file size | `--max-size 50MB` |
| `--skip-empty` | Skip empty files | `--skip-empty` |

### Security Scanning

| Flag | Description | Example |
|------|-------------|---------|
| `--scan-secrets` | Enable secret scanning | `--scan-secrets` |
| `--secrets-output` | Secrets output file | `--secrets-output secrets.json` |
| `--secrets-entropy` | Entropy threshold | `--secrets-entropy 3.5` |
| `--scan-endpoints` | Discover endpoints | `--scan-endpoints` |
| `--endpoints-output` | Endpoints output | `--endpoints-output endpoints.json` |

### JavaScript Analysis

| Flag | Description | Example |
|------|-------------|---------|
| `--js-beautify` | Beautify JavaScript | `--js-beautify` |
| `--extract-strings` | Extract JS strings | `--extract-strings` |
| `--strings-min-length` | Min string length | `--strings-min-length 10` |
| `--strings-pattern` | String regex pattern | `--strings-pattern "api.*"` |

### Output Formats

| Flag | Description | Example |
|------|-------------|---------|
| `--output-format` | Report format | `--output-format json` |
| `--output-file` | Output file path | `--output-file report.json` |
| `--pretty-json` | Pretty-print JSON | `--pretty-json` |

Formats: `text`, `json`, `csv`, `markdown`

## 📊 Performance

### Benchmarks

- **Speed**: ~1000+ requests/second with 50 workers
- **Memory**: Stable at ~25MB in long-running mode
- **Startup**: <10ms cold start
- **Binary Size**: ~9MB (compressed)

### Comparison with Python Version

| Metric | Python | Go (Downurl) |
|--------|--------|--------------|
| Requests/sec | ~100 | ~1000+ |
| Memory | Higher | ~25MB |
| Dependencies | requests | stdlib only |
| Startup | ~500ms | <10ms |
| Binary | N/A | ~9MB |

## 🏗️ Architecture

```
downurl/
├── cmd/downurl/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── downloader/      # HTTP client and worker pool
│   ├── parser/          # URL parsing (file, stdin)
│   ├── storage/         # File system and archiving
│   ├── reporter/        # Result reporting
│   ├── ui/              # Progress bar, colors, tables (v1.1.0)
│   ├── ratelimit/       # Token bucket rate limiter (v1.1.0)
│   └── watcher/         # File watching & scheduling (v1.1.0)
└── pkg/models/          # Shared data structures
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# With race detector
go test ./... -race

# With coverage
go test ./... -cover

# Static analysis
go vet ./...
```

**Test Coverage**: 28.5%
**Race Detector**: Clean ✓
**go vet**: Clean ✓

## 🛡️ Security

### Features
- Path traversal protection (sanitizes `../`, null bytes)
- Maximum file size enforcement (default 100MB)
- Timeout and retry configuration
- Secret detection with entropy analysis
- Safe hostname sanitization

### Security Testing
- 100+ security test cases
- Path traversal attack prevention
- Malicious URL handling
- Null byte injection protection

## 🔄 What's New in v1.1.0

### New Features
- ✨ **Enhanced UI**: Animated progress bar, colors, tables
- 📥 **Multiple Input Modes**: stdin, single URL, file
- ⚡ **Rate Limiting**: Token bucket algorithm
- 👀 **Watch Mode**: Auto-download on file changes
- ⏰ **Schedule Mode**: Periodic downloads
- ⚙️ **Config File**: INI-style `.downurlrc` support
- 💬 **Friendly Errors**: Helpful messages with suggestions

### Critical Bug Fixes
- 🐛 Watch/scheduler recursion bug (memory leaks)
- 🐛 Progress bar division by zero
- 🛡️ Path traversal vulnerability (v1.0.0)
- 🛡️ Hostname sanitization improvements

See [RELEASE_NOTES_v1.1.0.md](docs/RELEASE_NOTES_v1.1.0.md) for full details.

## 🗺️ Roadmap

### Planned for v1.2.0
- [ ] Resume capability for interrupted downloads
- [ ] Per-host rate limiting
- [ ] Custom user agents per URL pattern
- [ ] Prometheus metrics export
- [ ] Docker container support
- [ ] WebSocket support
- [ ] GraphQL endpoint discovery

### Under Consideration
- Web UI for monitoring downloads
- Plugin system for extensibility
- Database backend for large-scale tracking
- Distributed download coordination
- Cloud storage integration (S3, GCS, Azure)

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass (`go test ./... -race`)
5. Submit a pull request

## 📝 License

MIT License - See [LICENSE](LICENSE) file for details.

## 📚 Documentation

- [Getting Started](docs/user-guides/GETTING_STARTED.md) - Quick start guide
- [Configuration](docs/user-guides/CONFIGURATION.md) - Config file reference
- [Release Notes](docs/RELEASE_NOTES_v1.1.0.md) - What's new in v1.1.0
- [Architecture](docs/development/ARCHITECTURE.md) - System design
- [Documentation Index](docs/DOCUMENTATION_INDEX.md) - Complete docs

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/llvch/downurl/issues)
- **Discussions**: [GitHub Discussions](https://github.com/llvch/downurl/discussions)
- **Documentation**: [docs/](docs/)

## 🙏 Acknowledgments

Built with Go 1.24.9 using only the standard library.

---

**Download the latest release**: [v1.1.0](https://github.com/llvch/downurl/releases/tag/v1.1.0)

Made with ❤️ for the security research and web archiving communities.
