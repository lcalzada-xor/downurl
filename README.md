# Downurl

A high-performance, concurrent file downloader written in Go. Download files from URLs with worker pools, retry logic, and automatic archiving.

## Features

- **Concurrent Downloads**: Configurable worker pool for parallel downloads
- **Retry Logic**: Automatic retry with exponential backoff for failed downloads
- **Smart Filename Generation**: Safe filename extraction from URLs with hash fallback
- **Organized Storage**: Files organized by hostname in structured directories
- **Comprehensive Reporting**: Detailed reports with statistics and error tracking
- **Automatic Archiving**: Creates tar.gz archives of all downloaded content
- **Graceful Shutdown**: Handles interruption signals cleanly
- **Context-Aware**: Proper context cancellation throughout the application

## Architecture

```
downurl/
├── cmd/downurl/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── downloader/      # HTTP client and worker pool
│   ├── parser/          # URL parsing and validation
│   ├── storage/         # File system and archiving
│   └── reporter/        # Result reporting
└── pkg/models/          # Shared data structures
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
go install github.com/llvch/downurl/cmd/downurl@latest
```

## Usage

### Basic Usage

```bash
./downurl -input urls.txt
```

### Advanced Options

```bash
./downurl -input urls.txt \
          -output ./downloads \
          -workers 20 \
          -timeout 30s \
          -retry 5
```

### Command Line Flags

| Flag       | Description                      | Default   |
|------------|----------------------------------|-----------|
| `-input`   | Input file containing URLs       | *required*|
| `-output`  | Output directory                 | `output`  |
| `-workers` | Number of concurrent workers     | `10`      |
| `-timeout` | HTTP request timeout             | `15s`     |
| `-retry`   | Number of retry attempts         | `3`       |

### Environment Variables

You can also configure the application using environment variables:

```bash
export OUTPUT_DIR="./downloads"
export WORKERS=20
export TIMEOUT="30s"
export RETRY_ATTEMPTS=5

./downurl -input urls.txt
```

## Input File Format

Create a text file with one URL per line:

```
https://example.com/script.js
https://cdn.example.com/library.min.js
https://api.example.com/data.json

# Comments start with #
https://example.org/style.css
```

## Output Structure

```
output/
├── example.com/
│   └── js/
│       ├── script.js
│       └── library.min.js
├── api.example.com/
│   └── js/
│       └── data.json
├── report.txt           # Detailed download report
└── output.tar.gz        # Compressed archive
```

## Report Format

The generated report includes:

- Download statistics (successful, failed, total files)
- Average download duration
- Detailed results per URL
- Error messages for failed downloads

Example:

```
Download Report
Generated: 2025-11-16T10:00:00Z
Total URLs: 10
============================================================

Statistics:
  Successful: 8
  Failed: 2
  Total Downloaded: 8 files
  Total Errors: 2
  Average Duration: 1.2s
============================================================

Detailed Results:

[1] URL: https://example.com/script.js
    Host: example.com
    Duration: 1.5s
    Downloaded: 1 files
      - output/example.com/js/script.js
    Errors: 0
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover
```

### Project Structure

- **`cmd/downurl/`**: Main application entry point
- **`internal/config/`**: Configuration loading and validation
- **`internal/downloader/`**: HTTP client with retry logic and worker pool
- **`internal/parser/`**: URL parsing and filename generation
- **`internal/storage/`**: File system operations and tar.gz creation
- **`internal/reporter/`**: Result aggregation and report generation
- **`pkg/models/`**: Shared data structures

### Key Design Patterns

1. **Worker Pool**: Concurrent downloads using goroutines and channels
2. **Dependency Injection**: Components receive dependencies as parameters
3. **Context Propagation**: Proper cancellation handling throughout
4. **Error Wrapping**: Rich error context with `fmt.Errorf`
5. **Thread Safety**: Mutex protection for shared state

## Performance

- **Concurrency**: Default 10 workers, configurable up to system limits
- **Memory Efficient**: Streaming downloads without loading entire files in memory
- **Retry Logic**: Exponential backoff for transient failures
- **Resource Management**: Proper cleanup with defer statements

## Comparison with Python Version

| Feature                  | Python Version | Go Version |
|--------------------------|----------------|------------|
| Concurrency              | Single-threaded| Multi-threaded (goroutines) |
| Dependencies             | requests       | stdlib only |
| Type Safety              | Runtime        | Compile-time |
| Performance              | ~100 req/s     | ~1000+ req/s |
| Memory Usage             | Higher         | Lower |
| Binary Size              | N/A (script)   | ~8MB |
| Startup Time             | ~500ms         | <10ms |

## Examples

### Download JavaScript Files

```bash
# Create urls.txt
cat > urls.txt << EOF
https://cdn.jsdelivr.net/npm/vue@3/dist/vue.global.js
https://cdn.jsdelivr.net/npm/react@18/umd/react.production.min.js
https://unpkg.com/htmx.org@1.9.10
EOF

# Download with 5 workers
./downurl -input urls.txt -workers 5
```

### High-Performance Bulk Download

```bash
# Download with 50 workers and aggressive timeout
./downurl -input large-urls.txt \
          -workers 50 \
          -timeout 10s \
          -retry 5 \
          -output ./bulk-download
```

### Download with Environment Variables

```bash
# Set configuration
export OUTPUT_DIR="/var/downloads"
export WORKERS=30
export TIMEOUT="1m"

# Run downloader
./downurl -input production-urls.txt
```

## Error Handling

The application handles various error scenarios:

- **Network errors**: Automatic retry with backoff
- **HTTP errors**: 4xx errors are not retried, 5xx errors are retried
- **Timeout errors**: Configurable timeout with retry
- **File system errors**: Proper error reporting
- **Graceful shutdown**: SIGINT/SIGTERM handling

## Troubleshooting

### Problem: Downloads are slow

**Solution**: Increase the number of workers:
```bash
./downurl -input urls.txt -workers 50
```

### Problem: Timeouts on slow connections

**Solution**: Increase timeout duration:
```bash
./downurl -input urls.txt -timeout 1m
```

### Problem: Some URLs fail consistently

**Solution**: Check the report.txt file for error details. Increase retry attempts:
```bash
./downurl -input urls.txt -retry 10
```

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Roadmap

Future improvements:

- [ ] Progress bar for downloads
- [ ] Resume capability for interrupted downloads
- [ ] Support for authentication (Basic, Bearer tokens)
- [ ] Rate limiting per host
- [ ] Custom user agents per URL
- [ ] JSON output format for reports
- [ ] Prometheus metrics export
- [ ] Docker container support

## Author

Built with Go following best practices for scalability and maintainability.
