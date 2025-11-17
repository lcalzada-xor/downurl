# ⚙️ Configuration Guide

## Configuration File Support

Downurl supports configuration files to avoid repeating command-line flags every time.

---

## Configuration File Format

Downurl uses **INI-style** configuration files with multiple sections:

```ini
[defaults]
# Basic settings
workers = 20
timeout = 30s
output = ./downloads
mode = path

[auth]
# Authentication settings
bearer = ${API_TOKEN}
basic = username:password

[filters]
# Content filtering
extensions = js,css,json
max_size = 50MB
min_size = 1KB

[ratelimit]
# Rate limiting
default = 10/minute

[ui]
# User interface
quiet = false
no_progress = false
```

---

## Configuration File Discovery

Downurl automatically looks for configuration files in this order:

1. **Current directory**: `./.downurlrc`
2. **Home directory**: `~/.downurlrc`
3. **Explicit path**: `--config /path/to/config`

**Priority**: Command-line flags > Config file > Default values

---

## Configuration Sections

### [defaults] - Basic Settings

```ini
[defaults]
# Number of concurrent workers
workers = 20

# Request timeout (s, m, h format)
timeout = 30s

# Output directory
output = ./downloads

# Storage organization mode (flat, path, host, type, dated)
mode = host

# Retry attempts
retry = 3
```

**Available options**:
- `workers` - Number of concurrent downloads (default: 10)
- `timeout` - HTTP request timeout (default: 15s)
- `output` - Output directory (default: "output")
- `mode` - Storage mode: flat, path, host, type, dated (default: flat)
- `retry` - Retry attempts for failed downloads (default: 3)

---

### [auth] - Authentication

```ini
[auth]
# Bearer token authentication
bearer = ${API_TOKEN}

# Basic authentication (username:password)
basic = myuser:mypass

# Custom authorization header
header = X-API-Key ${MY_API_KEY}

# User agent
user_agent = Downurl/1.1.0
```

**Environment variable expansion**:
- Use `${VAR_NAME}` to reference environment variables
- Example: `bearer = ${GITHUB_TOKEN}`

**Available options**:
- `bearer` - Bearer token for Authorization header
- `basic` - Basic auth credentials (username:password)
- `header` - Custom authorization header
- `user_agent` - Custom User-Agent string

---

### [filters] - Content Filtering

```ini
[filters]
# Include only these extensions
extensions = js,jsx,ts,tsx,json

# Exclude these extensions
exclude_extensions = min.js,bundle.js

# Include only these content types
content_types = application/javascript,text/javascript

# Exclude these content types
exclude_types = image/png,image/jpeg

# File size limits
min_size = 1KB
max_size = 50MB

# Skip empty files
skip_empty = true
```

**Size formats**:
- Bytes: `1024`, `2048`
- Kilobytes: `1KB`, `500KB`
- Megabytes: `1MB`, `10MB`
- Gigabytes: `1GB`

**Available options**:
- `extensions` - Comma-separated list of allowed extensions
- `exclude_extensions` - Comma-separated list of excluded extensions
- `content_types` - Allowed MIME types
- `exclude_types` - Excluded MIME types
- `min_size` - Minimum file size
- `max_size` - Maximum file size (default: 100MB)
- `skip_empty` - Skip zero-byte files (default: false)

---

### [ratelimit] - Rate Limiting

```ini
[ratelimit]
# Global rate limit
default = 10/minute

# Per-domain rate limits (future feature)
# example.com = 5/second
# api.github.com = 60/hour
```

**Rate formats**:
- Per second: `10/second`, `10/s`
- Per minute: `100/minute`, `100/m`
- Per hour: `1000/hour`, `1000/h`

**Available options**:
- `default` - Global rate limit for all downloads

---

### [ui] - User Interface

```ini
[ui]
# Suppress all output
quiet = false

# Disable progress bar (keep logs)
no_progress = false

# Enable color output
color = true
```

**Available options**:
- `quiet` - Suppress all progress and UI output
- `no_progress` - Disable progress bar but keep logs
- `color` - Enable/disable colored output (default: true)

---

## Complete Configuration Example

### Example 1: Bug Bounty Configuration

```ini
# .downurlrc for bug bounty hunting

[defaults]
workers = 30
timeout = 20s
output = ./targets
mode = host
retry = 5

[auth]
bearer = ${BUGCROWD_TOKEN}
user_agent = Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36

[filters]
extensions = js,json,xml,txt
min_size = 100
max_size = 10MB
skip_empty = true

[ratelimit]
default = 5/second

[ui]
quiet = false
no_progress = false
```

### Example 2: Web Archiving

```ini
# .downurlrc for web archiving

[defaults]
workers = 50
timeout = 1m
output = ./archive
mode = dated
retry = 10

[filters]
max_size = 100MB

[ratelimit]
default = 20/second

[ui]
quiet = false
```

### Example 3: API Data Collection

```ini
# .downurlrc for API scraping

[defaults]
workers = 10
timeout = 30s
output = ./api-data
mode = path
retry = 3

[auth]
bearer = ${API_TOKEN}
header = X-API-Version 2024-01-01

[filters]
content_types = application/json,application/xml
min_size = 1

[ratelimit]
default = 10/minute

[ui]
no_progress = true
```

---

## Using Configuration Files

### Auto-discovery

Place `.downurlrc` in your current directory:

```bash
# Create config
cat > .downurlrc <<EOF
[defaults]
workers = 20
mode = host
EOF

# Run downurl (will auto-load .downurlrc)
downurl -input urls.txt
```

### Explicit Path

Specify a custom config file:

```bash
downurl -input urls.txt --config my-custom-config.ini
```

### Save Current Configuration

Save your current command-line settings to a config file:

```bash
# Run with desired settings
downurl -input urls.txt \
        -workers 30 \
        -timeout 1m \
        --mode host \
        --rate-limit "10/second" \
        --save-config my-config.ini

# Next time, just load the config
downurl -input urls.txt --config my-config.ini
```

---

## Environment Variables

All configuration options support environment variable expansion.

### Setting Environment Variables

**Linux/macOS**:
```bash
export API_TOKEN="ghp_xxxxxxxxxxxx"
export OUTPUT_DIR="/var/downloads"
export WORKERS=30
```

**Windows (PowerShell)**:
```powershell
$env:API_TOKEN="ghp_xxxxxxxxxxxx"
$env:OUTPUT_DIR="C:\Downloads"
$env:WORKERS=30
```

### Using in Config File

```ini
[defaults]
workers = ${WORKERS}
output = ${OUTPUT_DIR}

[auth]
bearer = ${API_TOKEN}
```

### Direct Environment Variables (No Config File)

Downurl also reads these environment variables directly:

```bash
export DOWNURL_WORKERS=20
export DOWNURL_TIMEOUT=30s
export DOWNURL_OUTPUT=./downloads
export DOWNURL_MODE=host
```

---

## Priority Order

When the same setting is specified in multiple places:

1. **Command-line flags** (highest priority)
2. **Explicit config file** (`--config path/to/file`)
3. **Current directory config** (`./.downurlrc`)
4. **Home directory config** (`~/.downurlrc`)
5. **Environment variables** (`DOWNURL_*`)
6. **Default values** (lowest priority)

**Example**:
```bash
# Config file has: workers = 20
# Environment has: DOWNURL_WORKERS=30
# Command-line has: -workers 50

# Result: 50 workers (command-line wins)
downurl -input urls.txt -workers 50
```

---

## Common Configuration Patterns

### Personal Default Config

Place in `~/.downurlrc` for your personal defaults:

```ini
[defaults]
workers = 20
timeout = 30s
mode = host

[ui]
quiet = false
```

### Project-Specific Config

Place in project directory `./.downurlrc`:

```ini
[defaults]
output = ./project-downloads
mode = dated

[auth]
bearer = ${PROJECT_API_TOKEN}

[filters]
extensions = js,json
```

### Multiple Configs for Different Tasks

```bash
# High-speed bulk download
downurl -input bulk.txt --config configs/high-speed.ini

# Careful rate-limited scraping
downurl -input targets.txt --config configs/rate-limited.ini

# Authenticated API access
downurl -input api.txt --config configs/api-auth.ini
```

---

## Validation and Debugging

### Check Current Configuration

```bash
# Save current effective config to see what's active
downurl --save-config current.ini

# View the generated config
cat current.ini
```

### Validate Config File

```bash
# Try a dry-run with the config
downurl -input urls.txt --config test.ini --dry-run
```

### Debug Configuration Loading

```bash
# Run with verbose logging (future feature)
downurl -input urls.txt --config test.ini --verbose
```

---

## Tips and Best Practices

1. **Use `.downurlrc` in project root** - Each project can have its own config
2. **Store credentials in env vars** - Never commit secrets to config files
3. **Use `--save-config` to document your workflow** - Save working configs
4. **Override with flags when needed** - Config is default, flags are explicit
5. **Version control `.downurlrc`** - Track project-specific settings (without secrets!)

---

## Examples by Use Case

### For Bug Bounty Hunters

```ini
[defaults]
workers = 20
mode = host
output = ./recon

[auth]
user_agent = Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)

[filters]
extensions = js,json,xml
min_size = 100

[ratelimit]
default = 5/second
```

### For Web Archivists

```ini
[defaults]
workers = 30
mode = dated
output = ./archive
retry = 10

[filters]
max_size = 500MB

[ui]
quiet = false
```

### For API Researchers

```ini
[defaults]
workers = 5
mode = path
timeout = 1m

[auth]
bearer = ${API_KEY}

[filters]
content_types = application/json

[ratelimit]
default = 2/second
```

---

## Related Documentation

- [Getting Started](GETTING_STARTED.md) - Basic usage
- [Usage Guide](USAGE.md) - Complete command-line reference
- [Advanced Features](ADVANCED.md) - Filtering, scanning, and more

---

Need help? Check out the [GitHub Issues](https://github.com/llvch/downurl/issues)!
