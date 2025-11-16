# Plan de Funcionalidades para Bug Bounty

Este documento detalla las funcionalidades faltantes importantes para convertir `downurl` en una herramienta profesional de Bug Bounty.

---

## 🎯 ANÁLISIS DE LA HERRAMIENTA ACTUAL

### Funcionalidades Existentes
- ✅ Descarga concurrente de archivos
- ✅ Reintentos con backoff exponencial
- ✅ Validación de URLs (http/https)
- ✅ Protección contra DoS (límite 100MB)
- ✅ Detección de colisiones
- ✅ Organización por hostname
- ✅ Reportes básicos
- ✅ Archivado tar.gz

### Limitaciones para Bug Bounty
❌ Sin capacidad de crawling/recursión
❌ Sin filtrado por tipo de archivo
❌ Sin extracción de secrets/tokens
❌ Sin detección de endpoints/APIs
❌ Sin análisis de contenido
❌ Sin integración con pipelines
❌ Sin deduplicación por hash
❌ Sin soporte para autenticación
❌ Sin rate limiting inteligente
❌ Sin manejo de JavaScript dinámico

---

## 📋 PLAN DE FUNCIONALIDADES - FASE 1 (Core Features)

### 1. 🔍 Content Filtering & Analysis (CRÍTICO)

**Prioridad**: ALTA
**Impacto**: Bug Bounty requiere análisis de contenido, no solo descarga

#### 1.1 Filtrado por Tipo de Contenido
```go
// internal/filter/content.go
type ContentFilter struct {
    AllowedTypes    []string // ["text/javascript", "text/html", "application/json"]
    BlockedTypes    []string // ["image/*", "video/*"]
    AllowedExtensions []string // [".js", ".html", ".json", ".xml"]
    BlockedExtensions []string // [".png", ".jpg", ".mp4"]
}
```

**Flags CLI**:
```bash
--filter-type "text/javascript,text/html,application/json"
--filter-ext ".js,.html,.json,.xml,.txt"
--exclude-type "image/*,video/*,audio/*"
--exclude-ext ".png,.jpg,.gif,.mp4"
--min-size 1KB    # Ignorar archivos muy pequeños
--max-size 10MB   # Ajustable desde 100MB default
```

**Casos de Uso**:
- Solo descargar archivos JS para análisis de endpoints
- Excluir imágenes/videos que no aportan a recon
- Filtrar por tamaño para evitar binarios grandes

---

#### 1.2 Secret & Token Detection (CRÍTICO)

**Prioridad**: ALTA
**Impacto**: Detección automática de leaks es fundamental en bug bounty

```go
// internal/scanner/secrets.go
type SecretScanner struct {
    Patterns []SecretPattern
}

type SecretPattern struct {
    Name    string // "AWS Key", "JWT Token", "API Key"
    Regex   *regexp.Regexp
    Entropy float64 // Para detectar strings aleatorios
}

type SecretFinding struct {
    URL         string
    File        string
    LineNumber  int
    SecretType  string
    Match       string
    Context     string // Líneas alrededor
    Confidence  string // "high", "medium", "low"
}
```

**Patrones a Detectar**:
- AWS Access Keys: `AKIA[0-9A-Z]{16}`
- AWS Secret Keys: Alto entropy strings de 40 chars
- GitHub Tokens: `ghp_[a-zA-Z0-9]{36}`
- Google API Keys: `AIza[0-9A-Za-z-_]{35}`
- Slack Tokens: `xox[baprs]-[0-9a-zA-Z-]{10,48}`
- Private Keys: `-----BEGIN (RSA|DSA|EC|OPENSSH) PRIVATE KEY-----`
- JWT Tokens: `eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*`
- Generic API Keys: High entropy (Shannon entropy > 4.5)
- Database URLs: `mongodb://`, `postgres://`, `mysql://`
- Passwords en código: `password\s*=\s*['"][^'"]+['"]`

**Flags CLI**:
```bash
--scan-secrets              # Activar escaneo
--secrets-output secrets.json
--secrets-strict            # Solo high confidence
--secrets-entropy 4.5       # Threshold de entropía
```

**Output JSON**:
```json
{
  "findings": [
    {
      "url": "https://example.com/config.js",
      "file": "output/example.com/js/config.js",
      "line": 42,
      "secret_type": "AWS Access Key",
      "match": "AKIAIOSFODNN7EXAMPLE",
      "context": "  aws_key: 'AKIAIOSFODNN7EXAMPLE',\n  aws_secret: 'wJalrX...'",
      "confidence": "high",
      "timestamp": "2025-11-16T10:00:00Z"
    }
  ],
  "summary": {
    "total_files_scanned": 150,
    "total_findings": 12,
    "high_confidence": 8,
    "medium_confidence": 3,
    "low_confidence": 1
  }
}
```

---

#### 1.3 Endpoint & API Discovery

**Prioridad**: ALTA
**Impacto**: Identificar endpoints es crucial para mapping de ataque

```go
// internal/scanner/endpoints.go
type EndpointScanner struct {
    Patterns []EndpointPattern
}

type EndpointFinding struct {
    URL          string
    File         string
    Endpoint     string
    Method       string   // GET, POST, etc.
    Parameters   []string
    LineNumber   int
    Context      string
}
```

**Detecciones**:
- Rutas de fetch/axios: `fetch\(['"]([^'"]+)['"]\)`
- AJAX calls: `\$.ajax\({.*url:\s*['"]([^'"]+)['"]`
- XMLHttpRequest: `xhr.open\(['"]([A-Z]+)['"],\s*['"]([^'"]+)['"]\)`
- REST endpoints: `/api/v[0-9]+/[a-zA-Z0-9/_-]+`
- GraphQL: `query.*{.*}`, endpoints `/graphql`
- WebSocket: `wss?://[^'"]+`
- URLs en strings: `https?://[a-zA-Z0-9][^'"\s]+`

**Flags CLI**:
```bash
--scan-endpoints
--endpoints-output endpoints.json
--endpoints-methods GET,POST,PUT,DELETE
```

**Output**:
```json
{
  "endpoints": [
    {
      "url": "https://example.com/app.js",
      "endpoint": "/api/v2/users/{id}",
      "method": "GET",
      "parameters": ["id"],
      "line": 156,
      "type": "rest_api"
    },
    {
      "url": "https://example.com/graphql.js",
      "endpoint": "/graphql",
      "method": "POST",
      "query": "query { user(id: $userId) { email } }",
      "type": "graphql"
    }
  ]
}
```

---

### 2. 🕷️ Recursive Crawling & Link Extraction (CRÍTICO)

**Prioridad**: ALTA
**Impacol**: Herramienta actual solo descarga lista fija. Bug bounty requiere descubrimiento.

```go
// internal/crawler/crawler.go
type Crawler struct {
    MaxDepth      int
    FollowExternal bool
    AllowedDomains []string
    Visited       map[string]bool
    Queue         chan CrawlJob
}

type CrawlJob struct {
    URL   string
    Depth int
}
```

**Funcionalidades**:
- Extracción de enlaces de HTML (`<a href>`, `<script src>`, `<link href>`)
- Extracción de imports en JS (`import './module.js'`)
- Seguir referencias en CSS (`@import`, `url()`)
- Extracción de URLs en JSON/XML
- Control de profundidad (max depth)
- Respeto de dominios permitidos
- Deduplicación de URLs visitadas

**Flags CLI**:
```bash
--crawl                          # Activar crawling
--crawl-depth 3                  # Profundidad máxima
--crawl-external                 # Seguir enlaces externos
--crawl-domains "*.example.com"  # Wildcard de dominios permitidos
--crawl-scope strict             # strict|medium|relaxed
```

**Ejemplo**:
```bash
# Crawlear example.com hasta profundidad 2, solo mismo dominio
./downurl --crawl --crawl-depth 2 --crawl-domains "example.com" -input seed_urls.txt
```

---

### 3. 🔐 Authentication & Custom Headers (ALTO)

**Prioridad**: ALTA
**Impacto**: Muchos targets requieren autenticación

```go
// internal/config/auth.go
type AuthConfig struct {
    Type    string // "bearer", "basic", "cookie", "header"
    Token   string
    Username string
    Password string
    Headers  map[string]string
    Cookies  map[string]string
}
```

**Flags CLI**:
```bash
# Bearer token
--auth-bearer "eyJhbGc..."

# Basic auth
--auth-basic "user:pass"

# Custom headers
--header "Authorization: Bearer token"
--header "X-API-Key: abc123"
--header "Cookie: session=xyz"

# Headers desde archivo
--headers-file headers.txt

# User agent personalizado
--user-agent "Mozilla/5.0 BugBountyBot/1.0"

# Cookie jar
--cookie-jar cookies.txt
```

**Headers file format**:
```
Authorization: Bearer eyJhbGc...
X-Custom-Header: value
Cookie: session=xyz; token=abc
```

---

### 4. 📊 Advanced Reporting & Output Formats (MEDIO)

**Prioridad**: MEDIA
**Impacto**: Integración con pipelines y herramientas

#### 4.1 Multiple Output Formats

```bash
--output-format json           # JSON estructurado
--output-format csv            # CSV para análisis
--output-format markdown       # Markdown legible
--output-format html           # HTML con highlight
--output-format nuclei         # Template para Nuclei
```

#### 4.2 JSON Output Structure
```json
{
  "scan_info": {
    "start_time": "2025-11-16T10:00:00Z",
    "end_time": "2025-11-16T10:05:30Z",
    "duration_seconds": 330,
    "target": "example.com",
    "seed_urls": 5,
    "total_urls": 147
  },
  "downloads": {
    "total": 147,
    "successful": 145,
    "failed": 2,
    "by_type": {
      "text/javascript": 89,
      "text/html": 34,
      "application/json": 12,
      "text/css": 10
    },
    "total_size_bytes": 15728640
  },
  "findings": {
    "secrets": 12,
    "endpoints": 78,
    "subdomains": 5
  },
  "files": [
    {
      "url": "https://example.com/app.js",
      "path": "output/example.com/js/app.js",
      "size": 45632,
      "content_type": "text/javascript",
      "sha256": "abc123...",
      "downloaded_at": "2025-11-16T10:01:23Z"
    }
  ]
}
```

#### 4.3 Integration Outputs

**Nuclei Template Generation**:
```yaml
# Auto-generated from endpoints
id: example-com-endpoints
info:
  name: Example.com API Endpoints
  author: downurl
  severity: info

requests:
  - method: GET
    path:
      - "{{BaseURL}}/api/v2/users"
      - "{{BaseURL}}/api/v2/products"
    matchers:
      - type: status
        status:
          - 200
```

---

### 5. 🎛️ Rate Limiting & Respectful Crawling (MEDIO)

**Prioridad**: MEDIA
**Impacto**: Evitar bans y ser respetuoso con el target

```go
// internal/ratelimit/limiter.go
type RateLimiter struct {
    RequestsPerSecond int
    BurstSize         int
    PerHostLimits     map[string]*rate.Limiter
}
```

**Flags CLI**:
```bash
--rate-limit 10              # 10 requests/segundo global
--rate-limit-host 5          # 5 requests/segundo por host
--rate-burst 20              # Burst size
--delay 100ms                # Delay entre requests
--random-delay 50-200ms      # Delay aleatorio
--respect-robots-txt         # Respetar robots.txt
```

**Features**:
- Rate limiting global y per-host
- Delays configurables y aleatorios
- Respect robots.txt parsing
- Backoff automático en 429 (Too Many Requests)

---

### 6. 📦 Deduplication & Hashing (MEDIO)

**Prioridad**: MEDIA
**Impacto**: Evitar descargar duplicados y reducir almacenamiento

```go
// internal/dedup/deduplicator.go
type Deduplicator struct {
    SeenHashes map[string]bool // SHA256 -> bool
    SeenURLs   map[string]bool
}
```

**Funcionalidades**:
- Hash SHA256 de contenido descargado
- Detección de duplicados por hash
- Detección de duplicados por URL canónica
- Skip descarga si ya existe con mismo hash
- Report de duplicados encontrados

**Flags CLI**:
```bash
--dedup-hash                 # Dedup por hash
--dedup-url                  # Dedup por URL
--dedup-database dedup.db    # SQLite cache persistente
```

**Output**:
```
Statistics:
  Unique files: 89
  Duplicate files: 34
  Deduplication ratio: 27.6%
  Storage saved: 12.3 MB
```

---

## 📋 PLAN DE FUNCIONALIDADES - FASE 2 (Advanced)

### 7. 🔬 JavaScript Analysis & Beautification (ALTO)

**Prioridad**: MEDIA-ALTA
**Impacto**: Análisis de JS ofuscado es común en bug bounty

```go
// internal/jsanalyzer/analyzer.go
type JSAnalyzer struct {
    Beautifier   *JSBeautifier
    DeObfuscator *Deobfuscator
}
```

**Features**:
- Beautify minified JavaScript
- Detección de ofuscación (eval, Function constructor)
- Source map extraction y aplicación
- Análisis de webpack bundles
- Extracción de strings interesantes

**Flags CLI**:
```bash
--js-beautify               # Beautify JS files
--js-extract-strings        # Extract all strings
--js-min-string-length 10   # Ignorar strings cortos
--js-sourcemap              # Download & apply sourcemaps
```

---

### 8. 🌐 Subdomain & URL Discovery (MEDIO)

**Prioridad**: MEDIA
**Impacto**: Descubrir assets adicionales

```go
// internal/discovery/subdomains.go
type SubdomainDiscovery struct {
    Found map[string]bool
}
```

**Sources**:
- URLs en archivos JS/HTML/CSS
- CSP headers (`Content-Security-Policy`)
- CORS headers (`Access-Control-Allow-Origin`)
- Certificate Transparency (si integrado)
- DNS prefetch hints (`<link rel="dns-prefetch">`)

**Output**:
```json
{
  "subdomains_discovered": [
    "api.example.com",
    "cdn.example.com",
    "admin.example.com"
  ],
  "urls_discovered": [
    "https://api.example.com/v2/users",
    "https://cdn.example.com/assets/app.js"
  ]
}
```

---

### 9. 🔌 Pipeline Integration (MEDIO)

**Prioridad**: MEDIA
**Impacto**: Integración con workflow de recon

```bash
# Input desde stdin (pipe de otras herramientas)
cat urls.txt | ./downurl --stdin

# Output a stdout para pipeline
./downurl --stdout --format json | jq '.findings.secrets'

# Webhooks
--webhook "https://slack.com/webhook/..."
--webhook-on "secret_found,endpoint_found"

# Database export
--export-db postgres://localhost/bounty
--export-table downloads
```

**Ejemplo Pipeline**:
```bash
# Subfinder -> httpx -> downurl -> secrets scan
subfinder -d example.com | \
  httpx -mc 200 | \
  ./downurl --stdin --crawl --scan-secrets --format json | \
  jq '.findings.secrets[] | select(.confidence=="high")'
```

---

### 10. 📈 Progress & Status Monitoring (BAJO)

**Prioridad**: BAJA
**Impacto**: UX mejorado

```bash
--progress                  # Barra de progreso
--status-server :8080       # HTTP status endpoint
--metrics-prometheus        # Prometheus metrics
```

**Progress Bar**:
```
Downloading: [████████████░░░░░░░░] 62% (89/145)
Current: https://example.com/app.js
Secrets: 12 | Endpoints: 78 | Failed: 2
ETA: 2m 34s | Speed: 15 files/s
```

**Status Endpoint** (`http://localhost:8080/status`):
```json
{
  "status": "running",
  "progress": 0.62,
  "total_urls": 145,
  "downloaded": 89,
  "failed": 2,
  "current_url": "https://example.com/app.js",
  "findings": {
    "secrets": 12,
    "endpoints": 78
  }
}
```

---

## 🎯 PRIORIZACIÓN - ROADMAP RECOMENDADO

### Sprint 1 (2-3 semanas): Core Bug Bounty Features
1. **Secret Scanner** (5 días) - CRÍTICO
2. **Endpoint Discovery** (3 días) - CRÍTICO
3. **Content Filtering** (3 días) - ALTO
4. **JSON Output** (2 días) - ALTO

### Sprint 2 (2-3 semanas): Discovery & Crawling
5. **Recursive Crawling** (5 días) - CRÍTICO
6. **Deduplication** (3 días) - MEDIO
7. **Subdomain Discovery** (3 días) - MEDIO

### Sprint 3 (2 semanas): Auth & Integration
8. **Authentication Support** (4 días) - ALTO
9. **Rate Limiting** (3 días) - MEDIO
10. **Pipeline Integration** (3 días) - MEDIO

### Sprint 4 (1-2 semanas): Advanced Analysis
11. **JS Beautification** (5 días) - MEDIO
12. **Advanced Reporting** (3 días) - BAJO

---

## 🏗️ ARQUITECTURA DE IMPLEMENTACIÓN

### Estructura de Paquetes Propuesta

```
downurl/
├── internal/
│   ├── scanner/           # NUEVO
│   │   ├── secrets.go
│   │   ├── endpoints.go
│   │   └── patterns.go
│   ├── crawler/           # NUEVO
│   │   ├── crawler.go
│   │   ├── extractor.go
│   │   └── queue.go
│   ├── filter/            # NUEVO
│   │   ├── content.go
│   │   └── size.go
│   ├── dedup/             # NUEVO
│   │   ├── deduplicator.go
│   │   └── database.go
│   ├── auth/              # NUEVO
│   │   ├── provider.go
│   │   └── headers.go
│   ├── ratelimit/         # NUEVO
│   │   └── limiter.go
│   ├── jsanalyzer/        # NUEVO (Fase 2)
│   │   ├── beautify.go
│   │   └── extractor.go
│   └── exporter/          # NUEVO
│       ├── json.go
│       ├── csv.go
│       └── nuclei.go
```

---

## 📏 CRITERIOS DE ÉXITO

### Métricas de Funcionalidad
- ✅ Detectar 95%+ de secrets conocidos (usar test suite de truffleHog)
- ✅ Extraer 90%+ de endpoints en aplicaciones JS modernas
- ✅ Crawling sin pérdida de memoria en 10,000+ URLs
- ✅ Rate limiting preciso (±5% del target)

### Métricas de Performance
- ✅ Scaneo de secrets < 100ms por archivo
- ✅ Crawling > 100 URLs/segundo
- ✅ Memory footprint < 500MB con 10K URLs

### Métricas de Usabilidad
- ✅ JSON output válido y bien estructurado
- ✅ Documentación completa de todos los flags
- ✅ Ejemplos de uso para casos comunes

---

## 🔧 CONSIDERACIONES TÉCNICAS

### Dependencies Sugeridas

```go
// go.mod additions
require (
    golang.org/x/time v0.5.0           // Rate limiting
    golang.org/x/net v0.20.0           // HTML parsing
    github.com/PuerkitoBio/goquery v1.8.1 // DOM traversal
    golang.org/x/sync v0.6.0           // errgroup, semaphore
)

// Optional (considerar trade-off de zero-dependency)
github.com/tdewolff/minify v2.20.9  // JS beautify (alternativa: escribir propio)
```

### Trade-offs
- **Stdlib-only vs Dependencies**: Secret scanner y endpoint discovery pueden hacerse con stdlib (regex), pero libraries especializadas serían más robustas
- **Performance vs Features**: Scanear contenido reduce throughput ~30-40%, hacer opcional con flags
- **Memory vs Speed**: Cache de dedup puede consumir RAM, ofrecer modo disk-based

---

## 🎓 CASOS DE USO - Bug Bounty

### Caso 1: Recon Inicial de Target
```bash
# Seed con subdomain discovery previo
subfinder -d target.com | httpx -mc 200 > seeds.txt

# Crawl profundo con análisis completo
./downurl \
  -input seeds.txt \
  --crawl --crawl-depth 3 \
  --scan-secrets --scan-endpoints \
  --filter-type "text/javascript,text/html,application/json" \
  --dedup-hash \
  --output-format json \
  -o target_recon.json
```

### Caso 2: JS Analysis para API Discovery
```bash
# Solo JS files, beautify y extract endpoints
./downurl \
  -input js_urls.txt \
  --filter-ext ".js" \
  --js-beautify \
  --scan-endpoints \
  --endpoints-output apis.json \
  --rate-limit-host 10
```

### Caso 3: Secret Hunting en Repos Públicos
```bash
# CDN crawl con secret detection
./downurl \
  -input cdn_urls.txt \
  --crawl --crawl-depth 2 \
  --scan-secrets --secrets-strict \
  --secrets-output secrets.json \
  --webhook "https://slack.com/bounty-alerts"
```

### Caso 4: Authenticated Target Scan
```bash
# Con autenticación y headers custom
./downurl \
  -input private_app.txt \
  --auth-bearer "eyJhbGc..." \
  --header "X-CSRF-Token: abc123" \
  --crawl --crawl-depth 2 \
  --scan-endpoints \
  --rate-limit-host 5
```

---

## ✅ RESUMEN EJECUTIVO

### Top 5 Features Faltantes (Must-Have)

1. **Secret Scanner** - Detección automática de leaks (AWS keys, tokens, passwords)
2. **Endpoint Discovery** - Extracción de APIs y rutas de aplicación
3. **Recursive Crawling** - Descubrimiento automático de assets vinculados
4. **Content Filtering** - Filtrar por tipo/extensión para reducir ruido
5. **JSON Output** - Formato estructurado para integración con pipelines

### Impacto Esperado

Con estas features, `downurl` se convierte en:
- 🎯 **Herramienta de Recon**: Crawling + discovery automático
- 🔍 **Secret Hunter**: Detección automática de sensitive data
- 🗺️ **API Mapper**: Extracción de endpoints para fuzzing
- 🔗 **Pipeline Component**: JSON output para integración
- ⚡ **High Performance**: Manteniendo concurrencia y speed

### Diferenciación

Herramientas actuales:
- **meg/hakrawler**: Solo crawling, sin análisis de contenido
- **gau/waybackurls**: Solo histórico, sin descarga
- **truffleHog**: Solo secrets, no descarga/crawl
- **gospider**: Crawling pero limited analysis

`downurl` con estas features sería:
- ✅ All-in-one: Download + Crawl + Analyze
- ✅ High Performance: Go concurrency
- ✅ Flexible: Modular pipeline integration
- ✅ Actionable: Secrets + Endpoints + Output formats

---

## 📚 REFERENCIAS & INSPIRACIÓN

### Herramientas Similares a Estudiar
- **truffleHog**: Secret scanning patterns
- **gospider**: Crawling implementation
- **meg**: Concurrent downloading
- **hakrawler**: Link extraction
- **nuclei**: Template output format
- **ffuf**: Rate limiting implementation

### Recursos
- OWASP Top 10 para detection patterns
- HackerOne disclosed reports para casos de uso
- Bug Bounty Forum discussions sobre tooling needs

---

**Documento creado**: 2025-11-16
**Última actualización**: 2025-11-16
**Autor**: Claude Code
**Versión**: 1.0
