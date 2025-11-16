# Post-Crawling Features for Bug Bounty

Ya que `downurl` es una herramienta **post-crawling** (recibe una lista de URLs ya descubiertas), estas son las funcionalidades más relevantes para bug bounty.

---

## 🎯 CONTEXTO: Post-Crawling Tool

`downurl` NO es un crawler, sino una herramienta que:
1. **Recibe** URLs (de gospider, hakrawler, waybackurls, katana, etc.)
2. **Descarga** el contenido de esas URLs
3. **Analiza** el contenido descargado
4. **Extrae** información útil para bug bounty

Por lo tanto, las features deben enfocarse en **análisis de contenido descargado**, no en descubrimiento.

---

## 📋 FUNCIONALIDADES PRIORITARIAS (Post-Crawling)

### 1. 🔍 **Content Analysis & Secret Detection** (CRÍTICO)

Ya que tienes los archivos descargados, analizar su contenido es esencial.

#### 1.1 Secret Scanner

```go
// internal/scanner/secrets.go
type SecretScanner struct {
    Patterns []SecretPattern
}

func (s *SecretScanner) ScanFile(path string) []SecretFinding
func (s *SecretScanner) ScanBatch(files []string) []SecretFinding
```

**Detecciones**:
- AWS Keys: `AKIA[0-9A-Z]{16}`
- API Keys (high entropy strings)
- JWT Tokens: `eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*`
- Private Keys: `-----BEGIN PRIVATE KEY-----`
- Database credentials
- GitHub tokens, Slack tokens, etc.

**CLI Flags**:
```bash
--scan-secrets              # Scan downloaded files for secrets
--secrets-output secrets.json
--secrets-min-entropy 4.5   # Shannon entropy threshold
```

**Output**:
```json
{
  "scan_time": "2025-11-16T10:00:00Z",
  "files_scanned": 150,
  "findings": [
    {
      "file": "output/example.com/js/config.js",
      "url": "https://example.com/config.js",
      "line": 42,
      "type": "AWS Access Key",
      "value": "AKIAIOSFODNN7EXAMPLE",
      "context": "const aws_key = 'AKIAIOSFODNN7EXAMPLE'",
      "confidence": "high"
    }
  ]
}
```

---

#### 1.2 Endpoint Discovery

Extraer endpoints de código JS descargado.

```bash
--scan-endpoints
--endpoints-output endpoints.json
--endpoints-format json|txt|burp
```

**Detecciones**:
- Fetch/Axios calls: `fetch('/api/users')`
- AJAX: `$.ajax({url: '/api/data'})`
- REST patterns: `/api/v1/users/{id}`
- GraphQL: `query { user { email } }`
- WebSocket endpoints

**Output** (Burp Suite format):
```
GET https://example.com/api/v1/users
POST https://example.com/api/v1/login
PUT https://example.com/api/v1/users/123
DELETE https://example.com/api/v1/users/123
```

---

#### 1.3 Comment Extraction

Comentarios de desarrolladores suelen contener información sensible.

```bash
--extract-comments
--comments-output comments.txt
--comments-format text|json
```

**Extracciones**:
- JS comments: `// TODO: remove hardcoded API key`
- HTML comments: `<!-- Debug mode: enabled -->`
- CSS comments: `/* Old endpoint: /api/v1/legacy */`

---

### 2. 📁 **Advanced Filtering & Classification** (ALTO)

Filtrar y clasificar archivos descargados para reducir ruido.

#### 2.1 Content-Type Filtering

```bash
# Solo descargar ciertos tipos
--filter-type "text/javascript,application/json"
--exclude-type "image/*,video/*"

# Validar Content-Type vs extensión
--validate-content-type  # Detectar JS con extensión .txt, etc.
```

#### 2.2 Size-Based Filtering

```bash
--min-size 1KB     # Ignorar archivos muy pequeños (vacíos, redirects)
--max-size 10MB    # Ajustar límite (default 100MB)
--skip-empty       # Skip archivos vacíos
```

#### 2.3 Content Classification

Clasificar automáticamente archivos descargados.

```bash
--classify         # Auto-classify files
```

**Output**:
```
Classification Report:
  JavaScript (minified): 45 files
  JavaScript (beautified): 12 files
  HTML: 23 files
  JSON (config): 8 files
  JSON (data): 15 files
  CSS: 10 files
  XML: 3 files
  Plain text: 5 files
  Binary (skipped): 2 files
```

---

### 3. 🔬 **JavaScript Analysis** (ALTO)

Analizar archivos JS para extraer información útil.

#### 3.1 Beautify/Deobfuscate

```bash
--js-beautify           # Beautify minified JS
--js-deobfuscate        # Basic deobfuscation
--js-sourcemap          # Try to find & apply sourcemaps
```

**Ejemplo**:
```bash
# Input: app.min.js (1 línea, 500KB)
# Output: app.beautified.js (15,000 líneas, formateado)
```

#### 3.2 String Extraction

```bash
--extract-strings
--strings-min-length 10     # Ignorar strings muy cortos
--strings-pattern "api|key|token|password"
```

**Output**:
```json
{
  "file": "app.js",
  "strings": [
    "https://api.internal.example.com",
    "TEMP_API_KEY_REMOVE_BEFORE_PROD",
    "/admin/debug/endpoints"
  ]
}
```

#### 3.3 Function/Variable Analysis

```bash
--extract-functions      # List all function names
--extract-variables      # List all var/let/const
```

**Uso**: Encontrar funciones sospechosas como `debugMode()`, `adminAccess()`, etc.

---

### 4. 📊 **Enhanced Output Formats** (MEDIO)

Formatos de salida para integración con otras herramientas.

#### 4.1 JSON Output (Structured)

```bash
--output-format json
--output-file results.json
```

```json
{
  "metadata": {
    "start_time": "2025-11-16T10:00:00Z",
    "duration_seconds": 120,
    "total_urls": 150,
    "successful": 145
  },
  "downloads": [
    {
      "url": "https://example.com/app.js",
      "path": "output/example.com/js/app.js",
      "size_bytes": 45632,
      "content_type": "text/javascript",
      "sha256": "abc123...",
      "downloaded_at": "2025-11-16T10:01:00Z"
    }
  ],
  "findings": {
    "secrets": [...],
    "endpoints": [...],
    "comments": [...]
  }
}
```

#### 4.2 CSV Export

```bash
--output-format csv
--output-file downloads.csv
```

```csv
URL,Path,Size,ContentType,SHA256,Status
https://example.com/app.js,output/.../app.js,45632,text/javascript,abc...,success
```

#### 4.3 Markdown Report

```bash
--output-format markdown
--output-file REPORT.md
```

Genera reporte legible con:
- Estadísticas
- Findings organizados por severidad
- Links a archivos descargados
- Recomendaciones

---

### 5. 🗂️ **Smart Deduplication & Hashing** (MEDIO)

Evitar descargar/analizar duplicados.

#### 5.1 Content-Based Dedup

```bash
--dedup-hash                # Dedup por SHA256 del contenido
--dedup-database dedup.db   # SQLite cache persistente
```

**Ventajas**:
- Detecta mismo archivo con URLs diferentes
- Evita analizar duplicados
- Reporta ratio de deduplicación

#### 5.2 URL Normalization

```bash
--normalize-urls    # Normalizar URLs antes de descargar
```

**Ejemplo**:
```
Input:
  https://example.com/app.js?v=1
  https://example.com/app.js?v=2
  https://example.com/app.js?cache=123

Output (normalized):
  https://example.com/app.js  (downloaded once)
```

#### 5.3 Hash Database

```bash
--hash-database hashes.json  # Store all file hashes
```

**Uso posterior**:
```bash
# Comparar con scan anterior
./downurl --compare-hashes old_hashes.json new_hashes.json
# Output: Added files, removed files, modified files
```

---

### 6. 🔗 **Link & Reference Extraction** (MEDIO)

Extraer referencias de archivos descargados (sin descargarlas).

```bash
--extract-links
--links-output links.txt
--links-depth internal    # internal|external|all
```

**Extracciones**:
- URLs en JS strings
- Script src, link href, img src
- CSS @import, url()
- JSON endpoints

**Output**:
```
# Internal links (same domain)
https://example.com/api/v2/users
https://example.com/admin/dashboard

# External links
https://cdn.example.net/lib.js
https://analytics.google.com/ga.js
```

**Uso**: Feed a otro crawler o análisis.

---

### 7. 🎭 **Response Analysis** (MEDIO)

Analizar respuestas HTTP, no solo contenido.

#### 7.1 Header Analysis

```bash
--analyze-headers
--headers-output headers.json
```

**Detecta**:
- CORS misconfiguration
- Security headers faltantes
- Información sensible en headers
- Server fingerprinting

```json
{
  "url": "https://example.com/api",
  "headers": {
    "Server": "Apache/2.4.1 (Ubuntu)",
    "X-Powered-By": "PHP/7.4.3",
    "Access-Control-Allow-Origin": "*"
  },
  "findings": [
    {
      "type": "CORS Misconfiguration",
      "severity": "medium",
      "description": "ACAO set to wildcard"
    }
  ]
}
```

#### 7.2 Status Code Tracking

```bash
--track-status-codes
```

**Output**:
```
Status Code Distribution:
  200 OK: 145 files
  301 Moved: 3 files
  404 Not Found: 2 files

Interesting Responses:
  403 Forbidden: https://example.com/admin/config.js
  500 Server Error: https://example.com/debug/info
```

---

### 8. 📦 **Diff Mode** (BAJO-MEDIO)

Comparar dos scans del mismo target.

```bash
# Scan 1
./downurl -input urls.txt -output scan1/

# Scan 2 (una semana después)
./downurl -input urls.txt -output scan2/

# Diff
./downurl --diff scan1/ scan2/ -output diff_report.md
```

**Output**:
```markdown
# Diff Report: scan1 vs scan2

## New Files (5)
- example.com/js/new_feature.js
- example.com/api/v2/endpoint

## Removed Files (2)
- example.com/js/old_lib.js

## Modified Files (12)
- example.com/js/app.js
  - Size changed: 45KB -> 52KB
  - SHA256 changed
  - New secrets found: 2
  - New endpoints found: 5
```

---

### 9. 🚀 **Performance Optimization** (BAJO)

Optimizaciones para scans grandes.

```bash
--resume-from scan.state    # Resume interrupted download
--checkpoint-interval 100   # Save state every N files
--skip-existing             # Skip already downloaded files
```

---

### 10. 🔌 **Pipeline Integration** (MEDIO)

Integración con workflows de recon.

#### 10.1 Stdin Support

```bash
# Recibir URLs desde stdin
cat urls.txt | ./downurl --stdin
gospider -s https://example.com | grep '\.js$' | ./downurl --stdin
```

#### 10.2 Stdout JSON Mode

```bash
# Output JSON a stdout para pipelines
./downurl --stdin --stdout-json | jq '.findings.secrets'
```

#### 10.3 Webhooks

```bash
--webhook "https://slack.com/webhook/xxx"
--webhook-on "secret_found,high_severity"
```

Envía notificación cuando encuentra algo crítico.

---

## 🎯 ROADMAP RECOMENDADO (Post-Crawling Focus)

### Sprint 1 (1-2 semanas): Secret Detection
1. **Secret Scanner** (5 días) - CRÍTICO
   - Patrones comunes (AWS, GitHub, JWT)
   - Shannon entropy para detección genérica
   - JSON output
   - Tests
2. **Endpoint Discovery** (3 días) - ALTO
   - Regex patterns para fetch/axios
   - REST API patterns
   - Output en formato Burp/Nuclei

### Sprint 2 (1-2 semanas): Content Analysis
3. **Content Filtering** (3 días) - ALTO
   - Filter by Content-Type
   - Min/max size filtering
   - Empty file detection
4. **JS Analysis** (4 días) - ALTO
   - Beautify minified JS
   - String extraction
   - Basic deobfuscation

### Sprint 3 (1 semana): Output & Integration
5. **JSON Output** (2 días) - MEDIO
   - Structured JSON format
   - CSV export
   - Markdown reports
6. **Pipeline Integration** (3 días) - MEDIO
   - Stdin support
   - Stdout JSON mode
   - Webhooks

### Sprint 4 (1 semana): Advanced Features
7. **Deduplication** (3 días) - MEDIO
   - Hash-based dedup
   - Persistent database
8. **Link Extraction** (2 días) - MEDIO
   - Extract URLs from downloaded files
   - Feed to other tools

---

## 💡 CASOS DE USO REALES - Bug Bounty

### Caso 1: JS Analysis Pipeline

```bash
# 1. Crawl target
gospider -s https://target.com -d 3 -c 10 | grep '\.js$' > js_urls.txt

# 2. Download & analyze JS files
./downurl -input js_urls.txt \
  --filter-ext ".js" \
  --scan-secrets \
  --scan-endpoints \
  --js-beautify \
  --extract-strings \
  --output-format json \
  -o js_analysis.json

# 3. Extract high-value findings
jq '.findings.secrets[] | select(.confidence=="high")' js_analysis.json

# 4. Feed endpoints to fuzzer
jq -r '.findings.endpoints[].url' js_analysis.json | ffuf -w - ...
```

### Caso 2: Authenticated App Analysis

```bash
# 1. Get JS files from authenticated app
cat authenticated_js_urls.txt | \
  ./downurl --stdin \
  --auth-bearer "$TOKEN" \
  --filter-type "text/javascript" \
  --scan-secrets \
  --secrets-output secrets.json

# 2. Review secrets found
jq '.findings[] | select(.type | contains("API"))' secrets.json
```

### Caso 3: Config File Hunt

```bash
# 1. Download potential config files
cat urls.txt | grep -E '\.(json|xml|yaml|config)$' | \
  ./downurl --stdin \
  --scan-secrets \
  --extract-strings \
  --strings-pattern "key|token|password|secret"

# 2. Review extracted strings
cat output/report.txt | grep -i "password\|secret\|key"
```

### Caso 4: Continuous Monitoring

```bash
#!/bin/bash
# daily_scan.sh

# Download current state
./downurl -input production_urls.txt \
  -output "scans/$(date +%Y%m%d)/" \
  --scan-secrets \
  --scan-endpoints \
  --dedup-hash

# Compare with yesterday
./downurl --diff \
  scans/$(date -d yesterday +%Y%m%d)/ \
  scans/$(date +%Y%m%d)/ \
  --webhook "$SLACK_WEBHOOK"
```

---

## 📊 COMPARACIÓN: Features vs Herramientas Existentes

| Feature | downurl | truffleHog | meg | gospider | gf |
|---------|---------|------------|-----|----------|----|
| Download files | ✅ | ❌ | ✅ | ❌ | ❌ |
| Secret scanning | 🔜 | ✅ | ❌ | ❌ | ✅ |
| Endpoint extraction | 🔜 | ❌ | ❌ | ❌ | ✅ |
| JS beautify | 🔜 | ❌ | ❌ | ❌ | ❌ |
| Authentication | ✅ | ❌ | ❌ | ✅ | ❌ |
| Concurrent | ✅ | ✅ | ✅ | ✅ | ❌ |
| JSON output | 🔜 | ✅ | ❌ | ✅ | ✅ |

✅ = Implementado | 🔜 = Planificado | ❌ = No disponible

---

## ✅ RESUMEN EJECUTIVO

### Top 5 Features (Post-Crawling)

1. **Secret Scanner** - Detección automática de credentials
2. **Endpoint Discovery** - Extraer APIs de JS
3. **Content Filtering** - Reducir ruido, solo archivos relevantes
4. **JS Beautify** - Analizar código minificado
5. **JSON Output** - Integración con pipelines

### Ventaja Competitiva

`downurl` sería único en combinar:
- ✅ **Download + Analysis** en una herramienta
- ✅ **Authentication** built-in
- ✅ **High Performance** (Go concurrency)
- ✅ **Bug Bounty Focus** (secrets, endpoints, configs)
- ✅ **Pipeline-Friendly** (JSON, stdin/stdout)

### Next Steps

Empezar por **Secret Scanner** (Sprint 1) ya que:
- Alto impacto para bug bounty
- Standalone (no depende de otras features)
- Relativamente simple de implementar
- Casos de uso claros

---

**Documento creado**: 2025-11-16
**Versión**: 1.0
