# Downurl v1.0.0 - Initial Release

**Release Date:** 2025-11-16

We're excited to announce the first stable release of Downurl, a high-performance concurrent file downloader written in Go!

## What is Downurl?

Downurl is a powerful command-line tool designed for downloading files from URLs with advanced features like:
- Concurrent downloads with worker pools
- Automatic retry logic
- Content filtering
- Secret scanning
- Endpoint discovery
- Multiple output formats

Perfect for bug bounty hunters, security researchers, web developers, and DevOps engineers.

## Highlights

### Performance
- **10x faster** than the Python version (~1000+ req/s vs ~100 req/s)
- **Low memory footprint** with streaming downloads
- **Fast startup** (<10ms)
- **Efficient concurrency** with configurable worker pools

### Key Features

#### 1. Concurrent Downloads
```bash
downurl -i urls.txt -w 50  # 50 concurrent workers
```

#### 2. Authentication Support
```bash
# Bearer token
downurl -i urls.txt --auth-bearer "your-token"

# Basic auth
downurl -i urls.txt --auth-basic "user:pass"

# Custom headers
downurl -i urls.txt --headers-file headers.txt
```

#### 3. Content Filtering
```bash
# Only download JavaScript files
downurl -i urls.txt --filter-ext "js,jsx"

# Skip files larger than 10MB
downurl -i urls.txt --max-size 10485760
```

#### 4. Security Scanning
```bash
# Scan for secrets
downurl -i urls.txt --scan-secrets --secrets-output secrets.json

# Discover API endpoints
downurl -i urls.txt --scan-endpoints --endpoints-output endpoints.json
```

#### 5. Multiple Output Formats
```bash
# JSON output
downurl -i urls.txt --output-format json --output-file report.json

# CSV output
downurl -i urls.txt --output-format csv --output-file report.csv

# Markdown output
downurl -i urls.txt --output-format markdown --output-file report.md
```

## Installation

### From Source
```bash
git clone https://github.com/llvch/downurl.git
cd downurl
go build -o downurl cmd/downurl/main.go
```

### Using Go Install
```bash
go install github.com/llvch/downurl/cmd/downurl@v1.0.0
```

## Quick Start

1. Create a file with URLs (one per line):
```
https://example.com/app.js
https://example.com/config.json
https://api.example.com/data
```

2. Run downurl:
```bash
./downurl -i urls.txt
```

3. Check the output:
```
output/
├── example.com/
│   ├── js/app.js
│   └── json/config.json
├── api.example.com/
│   └── data
├── report.txt
└── output.tar.gz
```

## Use Cases

### Bug Bounty Hunting
```bash
# Download all JS files and scan for secrets
downurl -i targets.txt \
  --filter-ext "js" \
  --scan-secrets \
  --scan-endpoints \
  --secrets-output secrets.json \
  --endpoints-output endpoints.json
```

### Asset Discovery
```bash
# Download with authentication and custom headers
downurl -i urls.txt \
  --auth-bearer "your-token" \
  --headers-file headers.txt \
  --user-agent "MyScanner/1.0"
```

### Bulk Download
```bash
# High-performance bulk download
downurl -i large-list.txt \
  --workers 100 \
  --timeout 30s \
  --retry 5
```

## Documentation

- [README.md](README.md) - Main documentation
- [ARCHITECTURE.md](ARCHITECTURE.md) - Technical architecture
- [AUTH.md](AUTH.md) - Authentication guide
- [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) - Migrating from Python version
- [CHANGELOG.md](CHANGELOG.md) - Full changelog

## Technical Details

- **Language:** Go 1.24.9
- **Dependencies:** None (pure stdlib)
- **Binary Size:** ~9MB
- **Test Coverage:** 28.5%
- **License:** MIT

## Known Issues

Before using in production, please be aware of these issues:

1. **Path traversal vulnerability** in hostname sanitization (CRITICAL)
2. **Potential goroutine leak** on context cancellation (HIGH)
3. **Race condition** in downloadAndSaveStream variables (HIGH)

We're working on fixes for these issues in v1.0.1. For detailed information, see the [CHANGELOG.md](CHANGELOG.md).

## Testing

All tests pass:
```bash
go test ./...  # All tests pass
go test ./... -race  # No race conditions in tests
go vet ./...  # Static analysis clean
```

## Performance Comparison

| Feature | Python Version | Go Version (v1.0.0) |
|---------|----------------|---------------------|
| Concurrency | Single-threaded | Multi-threaded (goroutines) |
| Dependencies | requests | stdlib only |
| Performance | ~100 req/s | ~1000+ req/s |
| Memory | Higher | Lower |
| Binary Size | N/A (script) | ~9MB |
| Startup Time | ~500ms | <10ms |

## What's Next?

Planned for future releases:
- Progress bar for downloads
- Resume capability
- Rate limiting per host
- Prometheus metrics
- Docker support
- WebSocket support

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Support

- **Issues:** https://github.com/llvch/downurl/issues
- **Discussions:** https://github.com/llvch/downurl/discussions

## Credits

Built by Lucas Calzada (@llvch)

## License

MIT License - See [LICENSE](LICENSE) file for details

---

Thank you for using Downurl! If you find it useful, please consider giving it a star on GitHub.
