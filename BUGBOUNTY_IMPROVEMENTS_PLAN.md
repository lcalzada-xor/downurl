# Plan de Mejoras - DownURL para Bug Bounty Hunters

## 📊 ANÁLISIS DE LA FUNCIONALIDAD ACTUAL

### ✅ Fortalezas Existentes
- Worker pool eficiente para descargas concurrentes
- Autenticación robusta (Bearer, Basic, Custom headers/cookies)
- Módulos de escaneo avanzados (secrets, endpoints)
- Filtrado de contenido con wildcards
- Análisis de JavaScript (beautify, detección de ofuscación)
- Output multi-formato (JSON, CSV, Markdown)

### ❌ Limitaciones Críticas Identificadas

#### 1. **MÓDULOS NO INTEGRADOS EN EL FLUJO PRINCIPAL**
**Problema**: Los módulos `processor`, `filter`, `scanner` existen pero NO se usan en `main.go`
- `internal/processor/processor.go` completo pero nunca se llama
- `internal/filter/content.go` completo pero no filtra durante descarga
- El escaneo de secrets/endpoints NO ocurre automáticamente

**Impacto**:
- ❌ No hay análisis automático post-descarga
- ❌ Se descargan archivos innecesarios (imágenes, videos)
- ❌ No se generan reportes con findings
- ❌ Funcionalidades implementadas pero inútiles

#### 2. **DESCARGA INEFICIENTE PARA BUG BOUNTY**
**Problema**: `internal/downloader/client.go:14-16`
```go
const MaxDownloadSize = 100 * 1024 * 1024 // 100 MB
```
- Límite hard-coded de 100MB
- Carga completo en memoria antes de guardar
- No hay streaming/descarga progresiva
- No se pueden descargar archivos grandes (logs, dumps, etc.)

**Impacto**:
- ❌ Memory exhaustion en scans grandes
- ❌ No se pueden analizar archivos > 100MB
- ❌ Descarga todo o nada (no progressive scan)

#### 3. **FALTA METADATA CRÍTICO PARA BUG BOUNTY**
**Problema**: No se capturan:
- Response headers (Server, X-Powered-By, X-Frame-Options, CSP, etc.)
- HTTP status codes de cada URL
- Redirect chains
- Certificados SSL/TLS
- DNS resolution info
- Technology stack indicators

**Impacto**:
- ❌ Pierdes información de reconnaissance valiosa
- ❌ No identificas tecnologías del target
- ❌ No detectas misconfigurations de headers

#### 4. **NO HAY DEDUPLICACIÓN**
**Problema**:
- Archivos con mismo contenido se guardan múltiples veces
- SHA256 se calcula pero no se usa para dedup
- URLs diferentes pueden servir mismo contenido

**Impacto**:
- ❌ Espacio en disco desperdiciado
- ❌ Findings duplicados en reportes
- ❌ Análisis redundante del mismo archivo

#### 5. **FALTA CAPACIDADES DE DISCOVERY**
**Problema**: Es un downloader puro, no extrae nuevas URLs
- No parsea HTML para extraer links
- No analiza JS imports/requires
- No descubre API schemas (OpenAPI, GraphQL introspection)
- No genera variaciones de URLs

**Impacto**:
- ❌ Se limita a URLs conocidas
- ❌ No descubre endpoints ocultos
- ❌ Requiere otro crawler externo siempre

---

## 🎯 PLAN DE MEJORAS PRIORIZADAS

### **FASE 1: INTEGRACIÓN DE MÓDULOS EXISTENTES** (CRÍTICO)
**Prioridad**: 🔴 CRÍTICA
**Esfuerzo**: Bajo
**Impacto**: Alto

#### Mejora 1.1: Integrar Processor en Main Workflow
**Archivo**: `cmd/downurl/main.go`

**Cambios**:
```go
// Después de la línea 112 (después de downloads)
log.Printf("\n[3.5/5] Processing downloaded files...")
processor := processor.NewProcessor(processor.Config{
    ScanSecrets:    cfg.ScanSecrets,
    ScanEndpoints:  cfg.ScanEndpoints,
    JSBeautify:     cfg.JSBeautify,
    SecretsEntropy: cfg.SecretsEntropy,
})

// Process all results
for _, result := range results {
    if err := processor.ProcessResult(result, cfg.OutputDir); err != nil {
        log.Printf("Warning: processing error: %v", err)
    }
}

// Save findings
if cfg.ScanSecrets {
    secretsPath := filepath.Join(cfg.OutputDir, "secrets.json")
    if err := processor.SaveSecrets(secretsPath); err != nil {
        log.Printf("Warning: failed to save secrets: %v", err)
    } else {
        log.Printf("Secrets saved to: %s", secretsPath)
    }
}

if cfg.ScanEndpoints {
    endpointsPath := filepath.Join(cfg.OutputDir, "endpoints.json")
    if err := processor.SaveEndpoints(endpointsPath); err != nil {
        log.Printf("Warning: failed to save endpoints: %v", err)
    } else {
        log.Printf("Endpoints saved to: %s", endpointsPath)
    }
}

// Generate comprehensive report
if cfg.OutputFile != "" {
    reporter := processor.GetReporter()
    switch cfg.OutputFormat {
    case "json":
        if err := reporter.GenerateJSON(cfg.OutputFile, cfg.PrettyJSON); err != nil {
            log.Printf("Warning: failed to generate JSON report: %v", err)
        }
    case "csv":
        if err := reporter.GenerateCSV(cfg.OutputFile); err != nil {
            log.Printf("Warning: failed to generate CSV report: %v", err)
        }
    case "markdown":
        if err := reporter.GenerateMarkdown(cfg.OutputFile); err != nil {
            log.Printf("Warning: failed to generate Markdown report: %v", err)
        }
    }
}
```

**Beneficio**:
- ✅ Análisis automático de todos los archivos descargados
- ✅ Generación de reportes con findings
- ✅ Exportación en formatos útiles (JSON/CSV/Markdown)

---

#### Mejora 1.2: Pre-Download Content Filtering
**Archivo**: `internal/downloader/downloader.go`

**Cambios**:
```go
// Modificar DownloadAll para aceptar ContentFilter
func (d *Downloader) DownloadAll(ctx context.Context, urls []string, contentFilter *filter.ContentFilter) []models.DownloadResult {
    // ... existing code ...

    for _, url := range urls {
        select {
        case <-ctx.Done():
            close(jobs)
            wg.Wait()
            close(results)
            return allResults
        case jobs <- downloadJob{url: url, filter: contentFilter}:
        }
    }
    // ... rest
}

// Worker checks filter before download
func (d *Downloader) worker(ctx context.Context, jobs <-chan downloadJob, results chan<- models.DownloadResult) {
    for job := range jobs {
        // HEAD request to get Content-Type and size
        contentType, contentLength := d.headRequest(ctx, job.url)

        // Check filter before downloading
        if job.filter != nil {
            shouldDownload, reason := job.filter.ShouldDownload(job.url, contentType, contentLength)
            if !shouldDownload {
                results <- models.DownloadResult{
                    URL:    job.url,
                    Errors: []string{"skipped: " + reason},
                }
                continue
            }
        }

        // Proceed with download
        result := d.downloadURL(ctx, job.url)
        results <- result
    }
}

// New method for HEAD request
func (d *Downloader) headRequest(ctx context.Context, url string) (string, int64) {
    req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
    if err != nil {
        return "", 0
    }

    resp, err := d.client.client.Do(req)
    if err != nil {
        return "", 0
    }
    defer resp.Body.Close()

    return resp.Header.Get("Content-Type"), resp.ContentLength
}
```

**Beneficio**:
- ✅ Filtra ANTES de descargar (ahorra ancho de banda)
- ✅ Salta archivos irrelevantes (imágenes, videos)
- ✅ Respeta min/max size limits
- ✅ Reduce uso de disco y tiempo de procesamiento

---

### **FASE 2: METADATA EXTRACTION** (ALTO)
**Prioridad**: 🟠 ALTA
**Esfuerzo**: Medio
**Impacto**: Alto

#### Mejora 2.1: Response Headers Capture
**Archivo nuevo**: `internal/metadata/collector.go`

**Implementación**:
```go
package metadata

import (
    "crypto/tls"
    "net/http"
    "time"
)

// ResponseMetadata contains comprehensive HTTP response metadata
type ResponseMetadata struct {
    // Request info
    URL            string    `json:"url"`
    Method         string    `json:"method"`
    RequestTime    time.Time `json:"request_time"`

    // Response info
    StatusCode     int                 `json:"status_code"`
    Status         string              `json:"status"`
    Headers        map[string][]string `json:"headers"`
    ContentType    string              `json:"content_type"`
    ContentLength  int64               `json:"content_length"`

    // Timing
    DNSTime        time.Duration `json:"dns_time"`
    ConnectTime    time.Duration `json:"connect_time"`
    TLSTime        time.Duration `json:"tls_time"`
    FirstByteTime  time.Duration `json:"first_byte_time"`
    TotalTime      time.Duration `json:"total_time"`

    // Redirects
    RedirectChain  []string `json:"redirect_chain,omitempty"`

    // TLS info
    TLSVersion     string   `json:"tls_version,omitempty"`
    CipherSuite    string   `json:"cipher_suite,omitempty"`
    ServerName     string   `json:"server_name,omitempty"`

    // Technology detection
    Server         string   `json:"server,omitempty"`
    PoweredBy      string   `json:"powered_by,omitempty"`
    Framework      string   `json:"framework,omitempty"`

    // Security headers
    SecurityHeaders SecurityHeaders `json:"security_headers"`
}

type SecurityHeaders struct {
    StrictTransportSecurity string `json:"strict_transport_security,omitempty"`
    ContentSecurityPolicy   string `json:"content_security_policy,omitempty"`
    XFrameOptions          string `json:"x_frame_options,omitempty"`
    XContentTypeOptions    string `json:"x_content_type_options,omitempty"`
    XXSSProtection         string `json:"x_xss_protection,omitempty"`
    ReferrerPolicy         string `json:"referrer_policy,omitempty"`
    PermissionsPolicy      string `json:"permissions_policy,omitempty"`
}

type Collector struct {
    captureHeaders  bool
    captureTiming   bool
    captureTLS      bool
}

func NewCollector(captureHeaders, captureTiming, captureTLS bool) *Collector {
    return &Collector{
        captureHeaders: captureHeaders,
        captureTiming:  captureTiming,
        captureTLS:     captureTLS,
    }
}

func (c *Collector) CollectMetadata(req *http.Request, resp *http.Response, timings *Timings) *ResponseMetadata {
    meta := &ResponseMetadata{
        URL:         req.URL.String(),
        Method:      req.Method,
        RequestTime: time.Now(),
        StatusCode:  resp.StatusCode,
        Status:      resp.Status,
        ContentType: resp.Header.Get("Content-Type"),
        ContentLength: resp.ContentLength,
    }

    if c.captureHeaders {
        meta.Headers = make(map[string][]string)
        for k, v := range resp.Header {
            meta.Headers[k] = v
        }

        // Extract technology indicators
        meta.Server = resp.Header.Get("Server")
        meta.PoweredBy = resp.Header.Get("X-Powered-By")
        meta.Framework = resp.Header.Get("X-AspNet-Version")

        // Extract security headers
        meta.SecurityHeaders = SecurityHeaders{
            StrictTransportSecurity: resp.Header.Get("Strict-Transport-Security"),
            ContentSecurityPolicy:   resp.Header.Get("Content-Security-Policy"),
            XFrameOptions:          resp.Header.Get("X-Frame-Options"),
            XContentTypeOptions:    resp.Header.Get("X-Content-Type-Options"),
            XXSSProtection:         resp.Header.Get("X-XSS-Protection"),
            ReferrerPolicy:         resp.Header.Get("Referrer-Policy"),
            PermissionsPolicy:      resp.Header.Get("Permissions-Policy"),
        }
    }

    if c.captureTiming && timings != nil {
        meta.DNSTime = timings.DNS
        meta.ConnectTime = timings.Connect
        meta.TLSTime = timings.TLS
        meta.FirstByteTime = timings.FirstByte
        meta.TotalTime = timings.Total
    }

    if c.captureTLS && resp.TLS != nil {
        meta.TLSVersion = tlsVersionString(resp.TLS.Version)
        meta.CipherSuite = tls.CipherSuiteName(resp.TLS.CipherSuite)
        meta.ServerName = resp.TLS.ServerName
    }

    return meta
}

func tlsVersionString(version uint16) string {
    switch version {
    case tls.VersionTLS10:
        return "TLS 1.0"
    case tls.VersionTLS11:
        return "TLS 1.1"
    case tls.VersionTLS12:
        return "TLS 1.2"
    case tls.VersionTLS13:
        return "TLS 1.3"
    default:
        return "Unknown"
    }
}

type Timings struct {
    DNS       time.Duration
    Connect   time.Duration
    TLS       time.Duration
    FirstByte time.Duration
    Total     time.Duration
}

// DetectTechnology analyzes headers and body to identify technologies
func DetectTechnology(headers http.Header, body []byte) []string {
    techs := []string{}

    // From headers
    if server := headers.Get("Server"); server != "" {
        techs = append(techs, "Server: "+server)
    }
    if poweredBy := headers.Get("X-Powered-By"); poweredBy != "" {
        techs = append(techs, "Powered-By: "+poweredBy)
    }

    // From body patterns (basic examples)
    bodyStr := string(body[:min(len(body), 10000)]) // Check first 10KB

    if strings.Contains(bodyStr, "wp-content") {
        techs = append(techs, "WordPress")
    }
    if strings.Contains(bodyStr, "Joomla!") {
        techs = append(techs, "Joomla")
    }
    if strings.Contains(bodyStr, "__VIEWSTATE") {
        techs = append(techs, "ASP.NET")
    }
    if strings.Contains(bodyStr, "ng-version") {
        techs = append(techs, "Angular")
    }
    if strings.Contains(bodyStr, "data-reactroot") || strings.Contains(bodyStr, "react") {
        techs = append(techs, "React")
    }

    return techs
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

**Archivo**: `internal/config/config.go` (añadir flags)
```go
// Metadata options
CaptureHeaders   bool
CaptureTiming    bool
CaptureTLS       bool
DetectTech       bool
MetadataOutput   string
```

**Beneficio**:
- ✅ Captura response headers completos
- ✅ Identifica tecnologías del target (WordPress, Angular, etc.)
- ✅ Detecta security headers missing/misconfigured
- ✅ Información de TLS/SSL para auditoría
- ✅ Timings para identificar endpoints lentos

---

#### Mejora 2.2: Enhanced Reporting with Metadata
**Archivo**: `internal/output/formats.go` (modificar)

**Cambios**:
```go
import "github.com/llvch/downurl/internal/metadata"

type DownloadInfo struct {
    // ... existing fields ...
    Metadata      *metadata.ResponseMetadata `json:"metadata,omitempty"`
    Technologies  []string                   `json:"technologies,omitempty"`
}

// In GenerateMarkdown, add section:
func (r *Reporter) GenerateMarkdown(filepath string) error {
    // ... existing code ...

    // Add Technology Stack section
    if len(techMap) > 0 {
        md.WriteString("## 🔧 Technology Stack Detected\n\n")
        for tech, count := range techMap {
            md.WriteString(fmt.Sprintf("- %s: found in %d files\n", tech, count))
        }
        md.WriteString("\n")
    }

    // Add Security Headers Analysis
    md.WriteString("## 🔒 Security Headers Analysis\n\n")
    missingHeaders := analyzeMissingSecurityHeaders(r.report.Downloads)
    if len(missingHeaders) > 0 {
        md.WriteString("### ⚠️ Missing Security Headers\n\n")
        for header, urls := range missingHeaders {
            md.WriteString(fmt.Sprintf("- **%s**: missing in %d responses\n", header, len(urls)))
        }
    }

    // ... rest
}
```

**Beneficio**:
- ✅ Reportes con análisis de tecnologías
- ✅ Detección de headers de seguridad faltantes
- ✅ Información estructurada para siguiente fase de testing

---

### **FASE 3: STREAMING DOWNLOAD & PROGRESSIVE ANALYSIS** (ALTO)
**Prioridad**: 🟠 ALTA
**Esfuerzo**: Alto
**Impacto**: Muy Alto

#### Mejora 3.1: Streaming Download Architecture
**Archivo nuevo**: `internal/downloader/stream.go`

**Implementación**:
```go
package downloader

import (
    "bufio"
    "context"
    "crypto/sha256"
    "fmt"
    "io"
    "net/http"
    "os"
)

const (
    StreamChunkSize = 64 * 1024 // 64KB chunks
    MaxStreamSize   = 500 * 1024 * 1024 // 500MB for streaming
)

type StreamDownloader struct {
    client       *HTTPClient
    chunkSize    int64
    maxSize      int64
    onChunk      ChunkCallback
}

type ChunkCallback func(chunk []byte, offset int64, totalSize int64) error

type StreamProgress struct {
    URL           string
    BytesRead     int64
    TotalSize     int64
    ChunksRead    int
    SHA256Running string
}

func NewStreamDownloader(client *HTTPClient, chunkSize, maxSize int64) *StreamDownloader {
    return &StreamDownloader{
        client:    client,
        chunkSize: chunkSize,
        maxSize:   maxSize,
    }
}

// DownloadStream downloads file in chunks with progressive processing
func (s *StreamDownloader) DownloadStream(ctx context.Context, url string, outputPath string, callback ChunkCallback) (*StreamProgress, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    resp, err := s.client.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status}
    }

    // Check total size
    totalSize := resp.ContentLength
    if totalSize > s.maxSize {
        return nil, fmt.Errorf("file too large: %d bytes (max: %d)", totalSize, s.maxSize)
    }

    // Create output file
    outFile, err := os.Create(outputPath)
    if err != nil {
        return nil, fmt.Errorf("failed to create output file: %w", err)
    }
    defer outFile.Close()

    // Stream with progressive hashing
    hasher := sha256.New()
    multiWriter := io.MultiWriter(outFile, hasher)

    reader := bufio.NewReaderSize(resp.Body, int(s.chunkSize))
    chunk := make([]byte, s.chunkSize)

    progress := &StreamProgress{
        URL:       url,
        TotalSize: totalSize,
    }

    for {
        select {
        case <-ctx.Done():
            return progress, ctx.Err()
        default:
            n, err := reader.Read(chunk)
            if n > 0 {
                // Write to file and hasher
                if _, writeErr := multiWriter.Write(chunk[:n]); writeErr != nil {
                    return progress, fmt.Errorf("write error: %w", writeErr)
                }

                progress.BytesRead += int64(n)
                progress.ChunksRead++

                // Call chunk callback for progressive analysis
                if callback != nil {
                    if callbackErr := callback(chunk[:n], progress.BytesRead-int64(n), totalSize); callbackErr != nil {
                        return progress, fmt.Errorf("callback error: %w", callbackErr)
                    }
                }

                // Check size limit
                if progress.BytesRead > s.maxSize {
                    return progress, fmt.Errorf("exceeded max size: %d", s.maxSize)
                }
            }

            if err != nil {
                if err == io.EOF {
                    // Complete
                    progress.SHA256Running = fmt.Sprintf("%x", hasher.Sum(nil))
                    return progress, nil
                }
                return progress, fmt.Errorf("read error: %w", err)
            }
        }
    }
}

// ProgressiveSecretScanner scans during download
type ProgressiveScanner struct {
    buffer       []byte
    bufferOffset int64
    secretScanner *scanner.SecretScanner
    findings     []scanner.SecretFinding
}

func NewProgressiveScanner(secretScanner *scanner.SecretScanner) *ProgressiveScanner {
    return &ProgressiveScanner{
        buffer:        make([]byte, 0, StreamChunkSize*2),
        secretScanner: secretScanner,
        findings:      []scanner.SecretFinding{},
    }
}

// ProcessChunk analyzes chunk as it's downloaded
func (ps *ProgressiveScanner) ProcessChunk(chunk []byte, offset int64, totalSize int64) error {
    // Append to buffer
    ps.buffer = append(ps.buffer, chunk...)

    // Scan buffer when it reaches threshold
    if len(ps.buffer) >= StreamChunkSize {
        // Scan current buffer
        secrets := ps.secretScanner.ScanBytes(ps.buffer, offset)
        ps.findings = append(ps.findings, secrets...)

        // Keep last 8KB for potential patterns spanning chunks
        keepSize := 8 * 1024
        if len(ps.buffer) > keepSize {
            ps.buffer = ps.buffer[len(ps.buffer)-keepSize:]
            ps.bufferOffset = offset + int64(len(chunk)) - int64(keepSize)
        }
    }

    return nil
}

func (ps *ProgressiveScanner) GetFindings() []scanner.SecretFinding {
    // Final scan of remaining buffer
    if len(ps.buffer) > 0 {
        secrets := ps.secretScanner.ScanBytes(ps.buffer, ps.bufferOffset)
        ps.findings = append(ps.findings, secrets...)
    }
    return ps.findings
}
```

**Modificar**: `internal/scanner/secrets.go` (añadir método)
```go
// ScanBytes scans byte slice directly (for streaming)
func (s *SecretScanner) ScanBytes(data []byte, baseOffset int64) []SecretFinding {
    findings := []SecretFinding{}

    content := string(data)
    lines := strings.Split(content, "\n")

    for lineNum, line := range lines {
        for _, pattern := range s.patterns {
            matches := pattern.Regex.FindAllString(line, -1)
            for _, match := range matches {
                findings = append(findings, SecretFinding{
                    Line:       int(baseOffset) + lineNum + 1,
                    SecretType: pattern.Name,
                    Match:      match,
                    Confidence: pattern.Confidence,
                    Context:    line,
                })
            }
        }

        // Entropy check
        if entropy := s.calculateEntropy(line); entropy >= s.entropyThreshold {
            // ... entropy-based detection
        }
    }

    return findings
}
```

**Beneficio**:
- ✅ Descarga archivos > 100MB (hasta 500MB configurable)
- ✅ Análisis progresivo durante descarga (encuentra secrets más rápido)
- ✅ Menor uso de memoria (streaming vs load completo)
- ✅ Puede cancelar download si encuentra algo crítico
- ✅ SHA256 calculado durante descarga (no re-lectura)

---

### **FASE 4: DEDUPLICATION & SMART STORAGE** (MEDIO)
**Prioridad**: 🟡 MEDIA
**Esfuerzo**: Medio
**Impacto**: Medio

#### Mejora 4.1: Content-Based Deduplication
**Archivo nuevo**: `internal/dedup/deduplicator.go`

**Implementación**:
```go
package dedup

import (
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sync"
)

type Deduplicator struct {
    seen     map[string]string // hash -> original file path
    mu       sync.RWMutex
    indexPath string
}

type DedupIndex struct {
    Entries map[string]DedupEntry `json:"entries"`
}

type DedupEntry struct {
    Hash         string   `json:"hash"`
    OriginalPath string   `json:"original_path"`
    OriginalURL  string   `json:"original_url"`
    DuplicateURLs []string `json:"duplicate_urls"`
    Size         int64    `json:"size"`
}

func NewDeduplicator(indexPath string) *Deduplicator {
    d := &Deduplicator{
        seen:      make(map[string]string),
        indexPath: indexPath,
    }

    // Load existing index
    d.loadIndex()

    return d
}

func (d *Deduplicator) loadIndex() error {
    data, err := os.ReadFile(d.indexPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // New index
        }
        return err
    }

    var index DedupIndex
    if err := json.Unmarshal(data, &index); err != nil {
        return err
    }

    d.mu.Lock()
    defer d.mu.Unlock()

    for hash, entry := range index.Entries {
        d.seen[hash] = entry.OriginalPath
    }

    return nil
}

func (d *Deduplicator) saveIndex() error {
    // Build index from seen map
    index := DedupIndex{
        Entries: make(map[string]DedupEntry),
    }

    d.mu.RLock()
    for hash, path := range d.seen {
        // Load entry info
        info, _ := os.Stat(path)
        index.Entries[hash] = DedupEntry{
            Hash:         hash,
            OriginalPath: path,
            Size:         info.Size(),
        }
    }
    d.mu.RUnlock()

    data, err := json.MarshalIndent(index, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(d.indexPath, data, 0644)
}

// CheckDuplicate returns (isDuplicate, originalPath, hash)
func (d *Deduplicator) CheckDuplicate(filePath string) (bool, string, string, error) {
    // Calculate hash
    hash, err := calculateFileHash(filePath)
    if err != nil {
        return false, "", "", err
    }

    d.mu.RLock()
    originalPath, exists := d.seen[hash]
    d.mu.RUnlock()

    if exists {
        return true, originalPath, hash, nil
    }

    // Not duplicate, register
    d.mu.Lock()
    d.seen[hash] = filePath
    d.mu.Unlock()

    // Save index
    d.saveIndex()

    return false, "", hash, nil
}

// CreateSymlink creates symlink for duplicate instead of copying
func (d *Deduplicator) CreateSymlink(targetPath, linkPath string) error {
    // Ensure directory exists
    dir := filepath.Dir(linkPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    // Create relative symlink
    relPath, err := filepath.Rel(filepath.Dir(linkPath), targetPath)
    if err != nil {
        return err
    }

    return os.Symlink(relPath, linkPath)
}

func calculateFileHash(filePath string) (string, error) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }

    hash := sha256.Sum256(data)
    return fmt.Sprintf("%x", hash), nil
}

// GetStats returns deduplication statistics
func (d *Deduplicator) GetStats() (totalFiles int, uniqueFiles int, duplicates int, spaceSaved int64) {
    d.mu.RLock()
    defer d.mu.RUnlock()

    uniqueFiles = len(d.seen)

    // Load index for full stats
    data, _ := os.ReadFile(d.indexPath)
    var index DedupIndex
    if json.Unmarshal(data, &index) == nil {
        for _, entry := range index.Entries {
            totalFiles += len(entry.DuplicateURLs) + 1
            duplicates += len(entry.DuplicateURLs)
            spaceSaved += entry.Size * int64(len(entry.DuplicateURLs))
        }
    }

    return
}
```

**Integrar en**: `internal/storage/filesystem.go`
```go
type FileStorage struct {
    // ... existing fields ...
    deduplicator *dedup.Deduplicator
}

func (fs *FileStorage) Save(url string, data []byte) (string, error) {
    // ... existing save logic ...

    // Check deduplication
    if fs.deduplicator != nil {
        isDup, originalPath, hash, err := fs.deduplicator.CheckDuplicate(finalPath)
        if err == nil && isDup {
            // Create symlink instead
            symlinkPath := finalPath + ".link"
            if err := fs.deduplicator.CreateSymlink(originalPath, symlinkPath); err == nil {
                return symlinkPath, nil
            }
        }
    }

    return finalPath, nil
}
```

**Beneficio**:
- ✅ Ahorra espacio en disco (symlinks para duplicados)
- ✅ Identifica contenido duplicado servido en URLs diferentes
- ✅ Reporta estadísticas de deduplicación
- ✅ Index persistente para runs posteriores

---

### **FASE 5: URL DISCOVERY & GENERATION** (MEDIO)
**Prioridad**: 🟡 MEDIA
**Esfuerzo**: Alto
**Impacto**: Muy Alto

#### Mejora 5.1: URL Extraction from Downloaded Content
**Archivo nuevo**: `internal/discovery/extractor.go`

**Implementación**:
```go
package discovery

import (
    "net/url"
    "regexp"
    "strings"
)

type URLExtractor struct {
    baseURL        *url.URL
    seenURLs       map[string]bool
    patterns       []*regexp.Regexp
    extractFromJS  bool
    extractFromHTML bool
}

type ExtractedURL struct {
    URL        string `json:"url"`
    Source     string `json:"source"`      // "href", "src", "fetch", etc.
    SourceFile string `json:"source_file"` // File where found
    Type       string `json:"type"`        // "absolute", "relative", "protocol-relative"
}

func NewURLExtractor(baseURL string, extractJS, extractHTML bool) (*URLExtractor, error) {
    base, err := url.Parse(baseURL)
    if err != nil {
        return nil, err
    }

    patterns := []*regexp.Regexp{
        // Absolute URLs
        regexp.MustCompile(`https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=%]+`),

        // Relative URLs in quotes
        regexp.MustCompile(`["'](/[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=%]*)["']`),

        // Protocol-relative URLs
        regexp.MustCompile(`["'](//[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=%]*)["']`),

        // HTML attributes
        regexp.MustCompile(`(?:href|src|action|data)=["']([^"']+)["']`),

        // JS fetch/axios/etc
        regexp.MustCompile(`(?:fetch|axios\.get|axios\.post|xhr\.open)\s*\(\s*["'\x60]([^"'\x60]+)["'\x60]`),

        // Import statements
        regexp.MustCompile(`import\s+.*?\s+from\s+["']([^"']+)["']`),
        regexp.MustCompile(`require\s*\(\s*["']([^"']+)["']\s*\)`),
    }

    return &URLExtractor{
        baseURL:        base,
        seenURLs:       make(map[string]bool),
        patterns:       patterns,
        extractFromJS:  extractJS,
        extractFromHTML: extractHTML,
    }, nil
}

func (e *URLExtractor) ExtractFromFile(filePath string, content []byte) []ExtractedURL {
    contentStr := string(content)
    extracted := []ExtractedURL{}

    // Check file type
    isJS := strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".mjs")
    isHTML := strings.HasSuffix(filePath, ".html") || strings.HasSuffix(filePath, ".htm")

    if (isJS && !e.extractFromJS) || (isHTML && !e.extractFromHTML) {
        return extracted
    }

    // Apply all patterns
    for _, pattern := range e.patterns {
        matches := pattern.FindAllStringSubmatch(contentStr, -1)
        for _, match := range matches {
            if len(match) < 2 {
                continue
            }

            rawURL := match[1]
            if rawURL == "" || strings.HasPrefix(rawURL, "#") {
                continue
            }

            // Normalize URL
            normalizedURL := e.normalizeURL(rawURL)
            if normalizedURL == "" {
                continue
            }

            // Check if already seen
            if e.seenURLs[normalizedURL] {
                continue
            }
            e.seenURLs[normalizedURL] = true

            extracted = append(extracted, ExtractedURL{
                URL:        normalizedURL,
                SourceFile: filePath,
                Type:       classifyURL(rawURL),
            })
        }
    }

    return extracted
}

func (e *URLExtractor) normalizeURL(rawURL string) string {
    // Protocol-relative
    if strings.HasPrefix(rawURL, "//") {
        return e.baseURL.Scheme + ":" + rawURL
    }

    // Absolute URL
    if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
        parsed, err := url.Parse(rawURL)
        if err != nil {
            return ""
        }
        // Only include same domain (avoid external links)
        if parsed.Host != e.baseURL.Host {
            return ""
        }
        return rawURL
    }

    // Relative URL
    if strings.HasPrefix(rawURL, "/") {
        return e.baseURL.Scheme + "://" + e.baseURL.Host + rawURL
    }

    // Relative to current path
    basePath := e.baseURL.Path
    if !strings.HasSuffix(basePath, "/") {
        basePath = basePath[:strings.LastIndex(basePath, "/")+1]
    }
    return e.baseURL.Scheme + "://" + e.baseURL.Host + basePath + rawURL
}

func classifyURL(rawURL string) string {
    if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
        return "absolute"
    }
    if strings.HasPrefix(rawURL, "//") {
        return "protocol-relative"
    }
    return "relative"
}

// ExtractAPIPaths finds API endpoints patterns
func (e *URLExtractor) ExtractAPIPaths(content []byte) []string {
    apis := []string{}
    seenAPIs := make(map[string]bool)

    contentStr := string(content)

    // API patterns
    apiPatterns := []*regexp.Regexp{
        regexp.MustCompile(`['"\x60](/api/[a-zA-Z0-9/_\-{}:]+)['"\x60]`),
        regexp.MustCompile(`['"\x60](/v[0-9]/[a-zA-Z0-9/_\-{}:]+)['"\x60]`),
        regexp.MustCompile(`['"\x60](/rest/[a-zA-Z0-9/_\-{}:]+)['"\x60]`),
        regexp.MustCompile(`['"\x60](/graphql)['"\x60]`),
    }

    for _, pattern := range apiPatterns {
        matches := pattern.FindAllStringSubmatch(contentStr, -1)
        for _, match := range matches {
            if len(match) >= 2 {
                api := match[1]
                if !seenAPIs[api] {
                    apis = append(apis, api)
                    seenAPIs[api] = true
                }
            }
        }
    }

    return apis
}
```

**Archivo nuevo**: `internal/discovery/generator.go`

**Implementación**:
```go
package discovery

import (
    "fmt"
    "net/url"
    "strconv"
    "strings"
)

type URLGenerator struct {
    baseURL     *url.URL
    wordlist    []string
    extensions  []string
    maxDepth    int
}

type GenerationRule struct {
    Type       string   // "extension", "parameter", "path", "numeric"
    Values     []string
    Ranges     []int    // For numeric generation
}

func NewURLGenerator(baseURL string, wordlist []string, extensions []string) (*URLGenerator, error) {
    base, err := url.Parse(baseURL)
    if err != nil {
        return nil, err
    }

    return &URLGenerator{
        baseURL:    base,
        wordlist:   wordlist,
        extensions: extensions,
        maxDepth:   3,
    }, nil
}

// GenerateFromTemplate generates URLs from template with parameters
// Template: /api/users/{id} -> /api/users/1, /api/users/2, ...
func (g *URLGenerator) GenerateFromTemplate(template string, rules []GenerationRule) []string {
    generated := []string{}

    for _, rule := range rules {
        switch rule.Type {
        case "numeric":
            if len(rule.Ranges) >= 2 {
                start, end := rule.Ranges[0], rule.Ranges[1]
                for i := start; i <= end; i++ {
                    url := strings.ReplaceAll(template, "{id}", strconv.Itoa(i))
                    url = strings.ReplaceAll(url, ":id", strconv.Itoa(i))
                    generated = append(generated, g.baseURL.Scheme+"://"+g.baseURL.Host+url)
                }
            }

        case "parameter":
            for _, value := range rule.Values {
                // Parse existing query params
                u, _ := url.Parse(template)
                q := u.Query()
                for k, v := range rule.Values {
                    q.Set(fmt.Sprintf("param%d", k), v)
                }
                u.RawQuery = q.Encode()
                generated = append(generated, g.baseURL.Scheme+"://"+g.baseURL.Host+u.String())
            }

        case "extension":
            basePath := strings.TrimSuffix(template, filepath.Ext(template))
            for _, ext := range g.extensions {
                generated = append(generated, g.baseURL.Scheme+"://"+g.baseURL.Host+basePath+ext)
            }
        }
    }

    return generated
}

// GenerateCommonPaths generates common paths from wordlist
func (g *URLGenerator) GenerateCommonPaths(basePath string) []string {
    generated := []string{}

    for _, word := range g.wordlist {
        path := basePath
        if !strings.HasSuffix(path, "/") {
            path += "/"
        }
        path += word

        // Generate with various extensions
        for _, ext := range g.extensions {
            generated = append(generated, g.baseURL.Scheme+"://"+g.baseURL.Host+path+ext)
        }
    }

    return generated
}

// ParameterFuzz generates variations of query parameters
func (g *URLGenerator) ParameterFuzz(baseURL string, params []string, values []string) []string {
    generated := []string{}

    u, err := url.Parse(baseURL)
    if err != nil {
        return generated
    }

    // Test each parameter
    for _, param := range params {
        for _, value := range values {
            q := u.Query()
            q.Set(param, value)
            u.RawQuery = q.Encode()
            generated = append(generated, u.String())
        }
    }

    return generated
}
```

**Beneficio**:
- ✅ Descubre URLs embebidas en JS/HTML descargados
- ✅ Genera variaciones de URLs con parámetros
- ✅ Fuzzing de paths con wordlists
- ✅ Expande descubrimiento sin crawling activo
- ✅ Identifica APIs no documentadas

---

### **FASE 6: RESUMABLE DOWNLOADS** (BAJO)
**Prioridad**: 🟢 BAJA
**Esfuerzo**: Medio
**Impacto**: Bajo

#### Mejora 6.1: Checkpoint System
**Archivo nuevo**: `internal/checkpoint/manager.go`

**Implementación**:
```go
package checkpoint

import (
    "encoding/json"
    "os"
    "sync"
    "time"
)

type Checkpoint struct {
    Timestamp      time.Time          `json:"timestamp"`
    TotalURLs      int                `json:"total_urls"`
    CompletedURLs  []string           `json:"completed_urls"`
    FailedURLs     []string           `json:"failed_urls"`
    PendingURLs    []string           `json:"pending_urls"`
    Statistics     Stats              `json:"statistics"`
}

type Stats struct {
    Downloaded     int   `json:"downloaded"`
    Failed         int   `json:"failed"`
    BytesDownloaded int64 `json:"bytes_downloaded"`
}

type Manager struct {
    checkpointPath string
    checkpoint     *Checkpoint
    mu             sync.RWMutex
    autoSave       bool
    saveInterval   time.Duration
}

func NewManager(path string, autoSave bool, saveInterval time.Duration) *Manager {
    m := &Manager{
        checkpointPath: path,
        checkpoint:     &Checkpoint{},
        autoSave:       autoSave,
        saveInterval:   saveInterval,
    }

    m.Load()

    if autoSave {
        go m.autoSaveLoop()
    }

    return m
}

func (m *Manager) Load() error {
    data, err := os.ReadFile(m.checkpointPath)
    if err != nil {
        if os.IsNotExist(err) {
            m.checkpoint = &Checkpoint{
                CompletedURLs: []string{},
                FailedURLs:    []string{},
                PendingURLs:   []string{},
            }
            return nil
        }
        return err
    }

    m.mu.Lock()
    defer m.mu.Unlock()

    return json.Unmarshal(data, &m.checkpoint)
}

func (m *Manager) Save() error {
    m.mu.RLock()
    m.checkpoint.Timestamp = time.Now()
    data, err := json.MarshalIndent(m.checkpoint, "", "  ")
    m.mu.RUnlock()

    if err != nil {
        return err
    }

    return os.WriteFile(m.checkpointPath, data, 0644)
}

func (m *Manager) autoSaveLoop() {
    ticker := time.NewTicker(m.saveInterval)
    defer ticker.Stop()

    for range ticker.C {
        m.Save()
    }
}

func (m *Manager) MarkCompleted(url string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.checkpoint.CompletedURLs = append(m.checkpoint.CompletedURLs, url)
    m.checkpoint.Statistics.Downloaded++

    // Remove from pending
    m.removePending(url)
}

func (m *Manager) MarkFailed(url string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.checkpoint.FailedURLs = append(m.checkpoint.FailedURLs, url)
    m.checkpoint.Statistics.Failed++

    m.removePending(url)
}

func (m *Manager) removePending(url string) {
    for i, pending := range m.checkpoint.PendingURLs {
        if pending == url {
            m.checkpoint.PendingURLs = append(
                m.checkpoint.PendingURLs[:i],
                m.checkpoint.PendingURLs[i+1:]...,
            )
            break
        }
    }
}

func (m *Manager) GetPendingURLs() []string {
    m.mu.RLock()
    defer m.mu.RUnlock()

    return append([]string{}, m.checkpoint.PendingURLs...)
}

func (m *Manager) IsCompleted(url string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()

    for _, completed := range m.checkpoint.CompletedURLs {
        if completed == url {
            return true
        }
    }
    return false
}

func (m *Manager) InitializeURLs(urls []string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.checkpoint.TotalURLs = len(urls)
    m.checkpoint.PendingURLs = append([]string{}, urls...)
}
```

**Integrar en**: `cmd/downurl/main.go`
```go
// Load checkpoint
checkpointPath := filepath.Join(cfg.OutputDir, "checkpoint.json")
checkpointMgr := checkpoint.NewManager(checkpointPath, true, 30*time.Second)

// Filter out already completed URLs
filteredURLs := []string{}
for _, url := range urls {
    if !checkpointMgr.IsCompleted(url) {
        filteredURLs = append(filteredURLs, url)
    }
}

if len(filteredURLs) < len(urls) {
    log.Printf("Resuming: %d URLs already completed, %d remaining", len(urls)-len(filteredURLs), len(filteredURLs))
}

checkpointMgr.InitializeURLs(filteredURLs)

// Download with checkpoint updates
// ... in download loop
checkpointMgr.MarkCompleted(url)
```

**Beneficio**:
- ✅ Resume interrupted scans
- ✅ No re-download de URLs ya completadas
- ✅ Auto-save cada 30 segundos
- ✅ Útil para scans largos (1000s de URLs)

---

## 📋 RESUMEN EJECUTIVO

### Problemas Críticos Actuales
1. ❌ **Módulos no integrados**: Processor, Filter, Scanner existen pero no se usan
2. ❌ **Límite 100MB**: No descarga archivos grandes
3. ❌ **Sin metadata**: Pierdes info de reconnaissance (headers, tech stack)
4. ❌ **Sin deduplicación**: Desperdicias espacio y tiempo
5. ❌ **Sin discovery**: Solo descarga URLs conocidas

### Impacto de las Mejoras

| Fase | Mejora | Esfuerzo | Impacto | Prioridad |
|------|--------|----------|---------|-----------|
| 1.1 | Integrar Processor | Bajo | Alto | 🔴 CRÍTICA |
| 1.2 | Pre-Download Filtering | Bajo | Alto | 🔴 CRÍTICA |
| 2.1 | Response Metadata | Medio | Alto | 🟠 ALTA |
| 2.2 | Enhanced Reporting | Medio | Alto | 🟠 ALTA |
| 3.1 | Streaming Download | Alto | Muy Alto | 🟠 ALTA |
| 4.1 | Deduplication | Medio | Medio | 🟡 MEDIA |
| 5.1 | URL Extraction | Alto | Muy Alto | 🟡 MEDIA |
| 5.2 | URL Generation | Medio | Alto | 🟡 MEDIA |
| 6.1 | Checkpoints | Medio | Bajo | 🟢 BAJA |

### Roadmap Recomendado

**Semana 1**: Implementar Fase 1 (Integración de módulos existentes)
- Máximo impacto con mínimo esfuerzo
- Desbloquea funcionalidades ya construidas
- Análisis automático funcional

**Semana 2-3**: Implementar Fase 2 (Metadata extraction)
- Añade información crítica para bug bounty
- Mejora reportes significativamente
- Identifica tecnologías automáticamente

**Semana 4-5**: Implementar Fase 3 (Streaming)
- Permite archivos > 100MB
- Análisis progresivo más rápido
- Mejor uso de recursos

**Semana 6**: Implementar Fase 4-5 (Dedup + Discovery)
- Optimiza almacenamiento
- Expande superficie de ataque descubierta
- Reduce dependencia de crawlers externos

**Opcional**: Implementar Fase 6 (Checkpoints)
- Solo si haces scans muy largos (1000s de URLs)
- Útil pero no crítico

### Flags CLI Propuestos (Nuevos)

```bash
# Metadata
--capture-headers     # Capture response headers
--capture-timing      # Capture request timing
--capture-tls         # Capture TLS info
--detect-tech         # Detect technologies
--metadata-output     # Path for metadata JSON

# Streaming
--stream-chunk-size   # Chunk size (default: 64KB)
--max-stream-size     # Max size for streaming (default: 500MB)
--progressive-scan    # Scan during download

# Deduplication
--dedup               # Enable deduplication
--dedup-index         # Path to dedup index

# Discovery
--extract-urls        # Extract URLs from content
--extract-apis        # Extract API paths
--url-output          # Path for discovered URLs
--generate-variations # Generate URL variations
--wordlist            # Wordlist for path generation

# Checkpoints
--checkpoint          # Enable checkpointing
--checkpoint-file     # Checkpoint file path
--resume              # Resume from checkpoint
```

### Ejemplo de Uso Completo (Post-Mejoras)

```bash
# Download con análisis completo
downurl \
  -input urls.txt \
  -output results/ \
  -workers 20 \
  \
  # Auth
  -auth-bearer "YOUR_TOKEN" \
  -headers-file headers.txt \
  \
  # Scanning
  -scan-secrets \
  -scan-endpoints \
  -secrets-entropy 4.5 \
  \
  # Filtering
  -filter-type "text/*,application/javascript,application/json" \
  -exclude-type "image/*,video/*" \
  -min-size 100 \
  -max-size 10485760 \
  -skip-empty \
  \
  # JS Analysis
  -js-beautify \
  -extract-strings \
  \
  # Metadata
  -capture-headers \
  -capture-timing \
  -detect-tech \
  \
  # Streaming
  -max-stream-size 524288000 \
  -progressive-scan \
  \
  # Deduplication
  -dedup \
  \
  # Discovery
  -extract-urls \
  -extract-apis \
  -url-output discovered_urls.txt \
  \
  # Output
  -output-format json \
  -output-file report.json \
  -pretty-json \
  \
  # Checkpoint
  -checkpoint \
  -resume

# Output structure:
results/
├── downloads/           # Downloaded files
│   ├── example.com/
│   │   ├── file1.js
│   │   ├── file1.beautified.js
│   │   └── file2.json
├── secrets.json         # All secrets found
├── endpoints.json       # All endpoints discovered
├── discovered_urls.txt  # New URLs extracted
├── report.json          # Comprehensive report
├── report.md            # Human-readable report
├── metadata.json        # Response metadata
├── dedup_index.json     # Deduplication index
├── checkpoint.json      # Resume checkpoint
└── output.tar.gz        # Archive
```

---

## 🎯 CONCLUSIÓN

**Estado Actual**: Tool sólido pero sub-utilizado. Módulos valiosos implementados pero no integrados.

**Prioridad Inmediata**:
1. ✅ Integrar Processor en main workflow (2-3 horas)
2. ✅ Implementar pre-download filtering (3-4 horas)
3. ✅ Añadir metadata capture (1-2 días)

**Objetivo**: Convertir downurl en una herramienta **completa** para bug bounty que:
- Descarga inteligentemente (filtering)
- Analiza automáticamente (secrets, endpoints, tech stack)
- Reporta comprehensivamente (JSON, Markdown con findings)
- Escala eficientemente (streaming, dedup, checkpoints)
- Descubre proactivamente (URL extraction, generation)

**ROI**: Con Fase 1 + Fase 2 implementadas, tendrás 80% del valor con 20% del esfuerzo.
