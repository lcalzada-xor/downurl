# Authentication Guide

This guide covers all authentication and custom header options available in `downurl`.

---

## 🔐 Authentication Methods

### 1. Bearer Token Authentication

Use this for APIs that require JWT tokens or OAuth bearer tokens.

#### Command Line

```bash
# Bearer token
./downurl -input urls.txt -auth-bearer "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Token with "Bearer" prefix (optional, will be added automatically)
./downurl -input urls.txt -auth-bearer "Bearer eyJhbGc..."
```

#### Environment Variable

```bash
export AUTH_BEARER="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
./downurl -input urls.txt
```

#### HTTP Request

```
GET /api/data HTTP/1.1
Host: example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

### 2. Basic Authentication

Use this for APIs requiring username and password.

#### Command Line

```bash
# Basic auth
./downurl -input urls.txt -auth-basic "username:password"

# Username only (no password)
./downurl -input urls.txt -auth-basic "username"

# Password with special characters (use quotes)
./downurl -input urls.txt -auth-basic "user:p@ss:word"
```

#### Environment Variable

```bash
export AUTH_BASIC="username:password"
./downurl -input urls.txt
```

#### HTTP Request

```
GET /api/data HTTP/1.1
Host: example.com
Authorization: Basic dXNlcm5hbWU6cGFzc3dvcmQ=
```

---

### 3. Custom Authorization Header

Use this for custom authentication schemes.

#### Command Line

```bash
# Custom auth header
./downurl -input urls.txt -auth-header "Token abc123xyz"

# API key in auth header
./downurl -input urls.txt -auth-header "ApiKey your-api-key-here"
```

#### HTTP Request

```
GET /api/data HTTP/1.1
Host: example.com
Authorization: Token abc123xyz
```

---

## 📋 Custom Headers

### Single Header via Command Line

```bash
# Custom User-Agent
./downurl -input urls.txt -user-agent "MyBot/1.0"
```

### Headers from File

Create a headers file (e.g., `headers.txt`):

```
# Format: Header-Name: value
Authorization: Bearer eyJhbGc...
X-API-Key: your-api-key-here
X-Custom-Header: custom-value
User-Agent: MyCustomBot/1.0
Accept: application/json
Referer: https://example.com
```

Use the file:

```bash
./downurl -input urls.txt -headers-file headers.txt
```

#### HTTP Request

```
GET /api/data HTTP/1.1
Host: example.com
Authorization: Bearer eyJhbGc...
X-API-Key: your-api-key-here
X-Custom-Header: custom-value
User-Agent: MyCustomBot/1.0
Accept: application/json
Referer: https://example.com
```

---

## 🍪 Cookies

### Cookie String

```bash
# Single cookie
./downurl -input urls.txt -cookie "session=abc123"

# Multiple cookies
./downurl -input urls.txt -cookie "session=abc123; token=xyz789; user_id=12345"
```

### Cookies from File

Create a cookies file (e.g., `cookies.txt`):

```
# Format: name=value
session=abc123def456xyz789
token=your-session-token-here
user_id=12345
remember_me=true
```

Use the file:

```bash
./downurl -input urls.txt -cookies-file cookies.txt
```

#### HTTP Request

```
GET /api/data HTTP/1.1
Host: example.com
Cookie: session=abc123def456xyz789; token=your-session-token-here; user_id=12345; remember_me=true
```

---

## 🔧 Combining Authentication Methods

You can combine different authentication options:

```bash
# Bearer auth + custom headers + cookies
./downurl -input urls.txt \
  -auth-bearer "eyJhbGc..." \
  -headers-file headers.txt \
  -cookie "session=abc123" \
  -user-agent "CustomBot/1.0"
```

**Important**: You can only use ONE of the following at a time:
- `-auth-bearer`
- `-auth-basic`
- `-auth-header`

But you CAN combine them with:
- `-headers-file`
- `-cookie` / `-cookies-file`
- `-user-agent`

---

## 📚 Common Use Cases

### Use Case 1: Authenticated API with JWT

```bash
# Download from API requiring JWT token
./downurl -input api_endpoints.txt \
  -auth-bearer "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -user-agent "MyApp/1.0"
```

### Use Case 2: Private CDN with Basic Auth

```bash
# Download from password-protected CDN
./downurl -input cdn_urls.txt \
  -auth-basic "user:password" \
  -workers 5
```

### Use Case 3: Web App with Session Cookie

```bash
# Download from authenticated web app
./downurl -input app_assets.txt \
  -cookie "session=abc123; XSRF-TOKEN=xyz789" \
  -headers-file headers.txt
```

### Use Case 4: API with Custom Headers

```bash
# API requiring multiple custom headers
cat > headers.txt << EOF
X-API-Key: your-api-key
X-Client-ID: your-client-id
X-Request-ID: unique-request-id
Accept: application/json
EOF

./downurl -input api_urls.txt -headers-file headers.txt
```

### Use Case 5: Bug Bounty Target with Auth

```bash
# Download JS files from authenticated target
./downurl -input js_urls.txt \
  -auth-bearer "your-session-token" \
  -cookie "session=abc; csrf=xyz" \
  -user-agent "Mozilla/5.0 (Windows NT 10.0; Win64; x64)" \
  -workers 10
```

---

## 🧪 Testing Authentication

### Test Bearer Token

```bash
# Create test URLs file
echo "https://httpbin.org/bearer" > test_urls.txt

# Test with bearer token
./downurl -input test_urls.txt -auth-bearer "test-token-123"

# Check the output
cat output/httpbin.org/*/bearer
```

### Test Basic Auth

```bash
# Create test URLs file
echo "https://httpbin.org/basic-auth/user/pass" > test_urls.txt

# Test with basic auth
./downurl -input test_urls.txt -auth-basic "user:pass"

# Check the output
cat output/httpbin.org/*/basic-auth_user_pass
```

### Test Custom Headers

```bash
# Create headers file
cat > headers.txt << EOF
X-Custom-Header: test-value
User-Agent: TestBot/1.0
EOF

# Create test URLs file
echo "https://httpbin.org/headers" > test_urls.txt

# Test with headers
./downurl -input test_urls.txt -headers-file headers.txt

# Check the output (should show your custom headers)
cat output/httpbin.org/*/headers
```

---

## ⚠️ Security Best Practices

### 1. Don't Hardcode Credentials

**Bad**:
```bash
./downurl -auth-bearer "my-secret-token" -input urls.txt
```

**Good**:
```bash
# Store in environment variable
export AUTH_BEARER=$(cat ~/.config/myapp/token)
./downurl -input urls.txt

# Or use a secure file
./downurl -auth-bearer "$(cat ~/.config/myapp/token)" -input urls.txt
```

### 2. Protect Headers and Cookies Files

```bash
# Create with restricted permissions
touch headers.txt
chmod 600 headers.txt  # Only owner can read/write
echo "Authorization: Bearer $(cat ~/.config/token)" > headers.txt
```

### 3. Use HTTPS URLs Only

```bash
# The tool validates that URLs use http:// or https://
# For authenticated requests, ALWAYS use https://
```

### 4. Don't Log Sensitive Data

```bash
# Redirect logs to a file with restricted permissions
./downurl -input urls.txt -auth-bearer "$TOKEN" 2>&1 | tee -a download.log
chmod 600 download.log
```

### 5. Clean Up After Use

```bash
# Remove sensitive files after download
./downurl -input urls.txt -headers-file headers.txt -cookies-file cookies.txt
rm -f headers.txt cookies.txt
```

---

## 🐛 Troubleshooting

### Problem: 401 Unauthorized

**Cause**: Invalid or expired token/credentials

**Solution**:
```bash
# Verify your token is valid
curl -H "Authorization: Bearer $TOKEN" https://api.example.com/test

# Check token expiration
# For JWT: decode at jwt.io

# Re-authenticate and get a new token
```

### Problem: 403 Forbidden

**Cause**: Valid auth but insufficient permissions

**Solution**:
```bash
# Check if you need additional headers
./downurl -headers-file headers.txt

# Verify user agent is not blocked
./downurl -user-agent "Mozilla/5.0 ..."
```

### Problem: Headers Not Being Applied

**Cause**: Invalid file format

**Solution**:
```bash
# Verify headers file format
cat headers.txt
# Should be: Header-Name: value (note the colon and space)

# Check for parsing errors
./downurl -headers-file headers.txt -input urls.txt
# Look for "failed to parse headers file" error
```

### Problem: Cookies Not Working

**Cause**: Invalid cookie format or expired cookies

**Solution**:
```bash
# Verify cookie format
cat cookies.txt
# Should be: name=value (one per line)

# Test cookie string format
./downurl -cookie "name=value; name2=value2" -input urls.txt
```

---

## 📖 Examples Directory

Check the `examples/` directory for template files:

- `examples/headers.txt` - Sample headers file
- `examples/cookies.txt` - Sample cookies file

Copy and modify these templates:

```bash
cp examples/headers.txt my-headers.txt
# Edit my-headers.txt with your values
./downurl -input urls.txt -headers-file my-headers.txt
```

---

## 🔗 Integration with Other Tools

### Extract Session from Browser

```bash
# Chrome/Firefox: Open DevTools > Application > Cookies
# Copy cookie values and create cookies.txt

# Or use browser extension to export cookies
# Then use with downurl
./downurl -input urls.txt -cookies-file exported-cookies.txt
```

### Get JWT Token from Login API

```bash
# Login and extract token
TOKEN=$(curl -s -X POST https://api.example.com/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}' \
  | jq -r '.token')

# Use token with downurl
./downurl -input urls.txt -auth-bearer "$TOKEN"
```

### Combine with Other Tools

```bash
# Get URLs from crawling tool, download with auth
gospider -s https://example.com | \
  grep -E '\.js$' | \
  ./downurl --stdin -auth-bearer "$TOKEN"
```

---

## 📝 Summary

| Authentication Type | Flag | Format | Example |
|---------------------|------|--------|---------|
| Bearer Token | `-auth-bearer` | Token string | `eyJhbGc...` |
| Basic Auth | `-auth-basic` | `username:password` | `user:pass` |
| Custom Auth | `-auth-header` | Auth value | `Token abc123` |
| Custom Headers | `-headers-file` | File path | `headers.txt` |
| User Agent | `-user-agent` | UA string | `MyBot/1.0` |
| Cookies (string) | `-cookie` | `name=value; ...` | `session=abc` |
| Cookies (file) | `-cookies-file` | File path | `cookies.txt` |

**Environment Variables**:
- `AUTH_BEARER` - Bearer token
- `AUTH_BASIC` - Basic auth credentials
- `AUTH_HEADER` - Custom auth header
- `COOKIE` - Cookie string
- `USER_AGENT` - User agent

---

**Document created**: 2025-11-16
**Last updated**: 2025-11-16
