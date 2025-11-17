# 🚀 Getting Started with Downurl

## Quick Start Guide

Downurl is a high-performance concurrent file downloader written in Go. This guide will help you get up and running in minutes.

---

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Visit the [releases page](https://github.com/llvch/downurl/releases) and download the appropriate binary for your system:

**Linux (AMD64)**:
```bash
curl -LO https://github.com/llvch/downurl/releases/latest/download/downurl-linux-amd64.tar.gz
tar -xzf downurl-linux-amd64.tar.gz
chmod +x downurl-linux-amd64
sudo mv downurl-linux-amd64 /usr/local/bin/downurl
```

**macOS (Apple Silicon)**:
```bash
curl -LO https://github.com/llvch/downurl/releases/latest/download/downurl-darwin-arm64.tar.gz
tar -xzf downurl-darwin-arm64.tar.gz
chmod +x downurl-darwin-arm64
sudo mv downurl-darwin-arm64 /usr/local/bin/downurl
```

**Windows**:
Download `downurl-windows-amd64.exe.tar.gz`, extract, and add to your PATH.

### Option 2: Build from Source

```bash
# Clone repository
git clone https://github.com/llvch/downurl.git
cd downurl

# Build
go build -o downurl cmd/downurl/main.go

# Optional: Install globally
sudo mv downurl /usr/local/bin/
```

### Verify Installation

```bash
downurl --version
```

---

## Your First Download

### 1. Single URL Mode (Quickest)

Download a single file without creating an input file:

```bash
downurl "https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.21/lodash.min.js"
```

### 2. Stdin Mode (Pipe URLs)

Pipe URLs directly from any command:

```bash
echo "https://example.com/script.js" | downurl
```

```bash
cat urls.txt | downurl
```

### 3. File Mode (Traditional)

Create a file with URLs (one per line):

```bash
cat > urls.txt <<EOF
https://cdnjs.cloudflare.com/ajax/libs/axios/0.27.2/axios.min.js
https://cdnjs.cloudflare.com/ajax/libs/vue/3.2.31/vue.global.min.js
https://unpkg.com/htmx.org@1.9.10
EOF

downurl -input urls.txt
```

---

## Basic Usage Examples

### Download with Progress Bar

```bash
downurl -input urls.txt
```

Output:
```
⣾ Downloading... [==================>     ] 75% | 3.2 MB/s | ETA: 5s

✓ Successfully downloaded 8 files
✗ Failed: 2 files
⚠ Total time: 12.5s
```

### Increase Concurrency

Download faster with more workers:

```bash
downurl -input urls.txt -workers 20
```

### Organize by Hostname

Group downloaded files by their hostname:

```bash
downurl -input urls.txt --mode host
```

Output structure:
```
output/
├── cdnjs.cloudflare.com/
│   ├── axios.min.js
│   └── vue.global.min.js
└── unpkg.com/
    └── htmx.org
```

---

## Essential Command-Line Options

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-input` | Input file with URLs | (stdin) | `-input urls.txt` |
| `-output` | Output directory | `output` | `-output downloads` |
| `-workers` | Concurrent workers | `10` | `-workers 20` |
| `-timeout` | Request timeout | `15s` | `-timeout 30s` |
| `--mode` | Storage mode | `flat` | `--mode host` |
| `--quiet` | Suppress output | `false` | `--quiet` |

---

## Storage Modes

Downurl offers 5 ways to organize downloaded files:

### 1. Flat Mode (Default)
All files in one directory:
```bash
downurl -input urls.txt --mode flat
```
```
output/
├── file1.js
├── file2.css
└── file3.json
```

### 2. Path Mode
Replicate URL directory structure:
```bash
downurl -input urls.txt --mode path
```
```
output/
├── api/
│   └── v1/
│       └── data.json
└── assets/
    └── script.js
```

### 3. Host Mode
Group by hostname:
```bash
downurl -input urls.txt --mode host
```
```
output/
├── example.com/
│   └── script.js
└── cdn.example.com/
    └── library.js
```

### 4. Type Mode
Organize by file extension:
```bash
downurl -input urls.txt --mode type
```
```
output/
├── js/
│   ├── script.js
│   └── library.js
├── css/
│   └── styles.css
└── json/
    └── data.json
```

### 5. Dated Mode
Group by download date:
```bash
downurl -input urls.txt --mode dated
```
```
output/
├── 2025-11-17/
│   ├── file1.js
│   └── file2.css
└── 2025-11-18/
    └── file3.json
```

---

## Common Use Cases

### Web Scraping / Bug Bounty

Download JavaScript files from a target:
```bash
cat discovered_js_urls.txt | downurl --mode host -workers 30
```

### Asset Archiving

Archive website assets with date organization:
```bash
downurl -input assets.txt --mode dated -output archive/
```

### CDN Mirror

Mirror CDN libraries locally:
```bash
cat cdn_urls.txt | downurl --mode path -output cdn-mirror/
```

### API Response Collection

Download API responses with authentication:
```bash
downurl -input api_endpoints.txt \
        --auth-bearer "your_token" \
        --mode host
```

---

## Next Steps

Now that you know the basics:

1. **Learn Advanced Features**: Check out the [Usage Guide](USAGE.md) for:
   - Rate limiting
   - Authentication
   - Content filtering
   - Watch & schedule modes

2. **Configure Downurl**: See [Configuration Guide](CONFIGURATION.md) for:
   - Config file setup
   - Environment variables
   - Custom defaults

3. **Explore Filtering**: Read [Advanced Features](ADVANCED.md) for:
   - Secret scanning
   - Endpoint discovery
   - JavaScript beautification

---

## Troubleshooting

### Downloads are slow
**Solution**: Increase workers
```bash
downurl -input urls.txt -workers 50
```

### Timeouts on slow connections
**Solution**: Increase timeout
```bash
downurl -input urls.txt -timeout 1m
```

### Too much output
**Solution**: Use quiet mode
```bash
downurl -input urls.txt --quiet
```

### Want to see errors only
**Solution**: Disable progress bar
```bash
downurl -input urls.txt --no-progress
```

---

## Getting Help

- **Full documentation**: See [Usage Guide](USAGE.md)
- **Report bugs**: [GitHub Issues](https://github.com/llvch/downurl/issues)
- **View examples**: [docs/user-guides/](../user-guides/)

---

**Quick Reference Card**

```bash
# Download single URL
downurl "https://example.com/file.js"

# Download from stdin
cat urls.txt | downurl

# Download from file
downurl -input urls.txt

# Download with 20 workers
downurl -input urls.txt -workers 20

# Download and organize by host
downurl -input urls.txt --mode host

# Quiet mode
downurl -input urls.txt --quiet

# Save configuration
downurl -input urls.txt --save-config .downurlrc
```

---

Happy downloading! 🚀
