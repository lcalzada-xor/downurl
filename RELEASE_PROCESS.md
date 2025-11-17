# 🚀 Proceso de Release para Downurl

## 📋 Plan de Pasos para Nueva Release

Este documento detalla el proceso completo para preparar y publicar una nueva versión de Downurl.

---

## 🎯 Fase 1: Preparación Pre-Release

### 1.1 Verificar Estado del Código
```bash
# Verificar que no hay cambios sin commitear
git status

# Verificar que estás en la rama principal
git checkout main
git pull origin main

# Verificar la última versión
git tag -l | tail -1
```

### 1.2 Ejecutar Tests Completos
```bash
# Tests unitarios
go test ./... -v

# Tests con race detector
go test ./... -race

# Análisis estático
go vet ./...

# Verificar cobertura
go test ./... -cover -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### 1.3 Verificar Compilación Multi-Plataforma
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o build/downurl-linux-amd64 cmd/downurl/main.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o build/downurl-linux-arm64 cmd/downurl/main.go

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o build/downurl-darwin-amd64 cmd/downurl/main.go

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o build/downurl-darwin-arm64 cmd/downurl/main.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o build/downurl-windows-amd64.exe cmd/downurl/main.go
```

---

## 📝 Fase 2: Documentación

### 2.1 Actualizar CHANGELOG.md
```bash
# Editar CHANGELOG.md
# - Cambiar fecha de "TBD" a fecha actual
# - Revisar que todas las características estén documentadas
# - Verificar que todos los bugs fixes estén listados
# - Añadir enlaces de comparación de versiones
```

**Checklist CHANGELOG:**
- [ ] Fecha de release actualizada
- [ ] Todas las nuevas características listadas
- [ ] Todos los bug fixes documentados
- [ ] Breaking changes identificados (si aplica)
- [ ] Enlaces de comparación añadidos
- [ ] Sección [Unreleased] actualizada

### 2.2 Actualizar README.md
```bash
# Actualizar README.md con:
# - Nuevas características de v1.1.0
# - Ejemplos de uso actualizados
# - Nuevos flags y opciones
# - Enlaces a nueva documentación
```

**Checklist README:**
- [ ] Features section actualizada
- [ ] Installation instructions correctas
- [ ] Usage examples con nuevas características
- [ ] Command-line flags actualizados
- [ ] Links a documentación verificados
- [ ] Badges de versión actualizados

### 2.3 Crear Release Notes
```bash
# Crear RELEASE_NOTES_v1.1.0.md
# - Resumen ejecutivo para usuarios
# - Características destacadas
# - Breaking changes (si aplica)
# - Guía de actualización
# - Known issues
```

### 2.4 Verificar Documentación en docs/
```bash
# Revisar y actualizar documentación en:
# - docs/user-guides/
# - docs/development/
# - docs/migration/
```

---

## 🧪 Fase 3: Testing y QA

### 3.1 Tests Funcionales
```bash
# Test básico de descarga
echo "https://example.com/test.js" | ./downurl

# Test con archivo
cat > test_urls.txt <<EOF
https://cdnjs.cloudflare.com/ajax/libs/lodash.js/4.17.21/lodash.min.js
https://cdnjs.cloudflare.com/ajax/libs/axios/0.27.2/axios.min.js
EOF
./downurl -input test_urls.txt

# Test watch mode
./downurl -input test_urls.txt --watch

# Test schedule mode
./downurl -input test_urls.txt --schedule "10s"

# Test rate limiting
./downurl -input test_urls.txt --rate-limit "5/second"
```

### 3.2 Tests de Configuración
```bash
# Test config file
cat > .downurlrc <<EOF
[defaults]
workers = 20
timeout = 30s

[ratelimit]
default = 10/minute
EOF

./downurl -input test_urls.txt

# Test save config
./downurl -input test_urls.txt --save-config my-config.ini
```

### 3.3 Tests de Storage Modes
```bash
# Probar cada modo de almacenamiento
for mode in flat path host type dated; do
    echo "Testing mode: $mode"
    ./downurl -input test_urls.txt --mode $mode -output "test_$mode"
done
```

### 3.4 Tests de Autenticación
```bash
# Test Bearer auth
./downurl -input urls.txt --auth-bearer "your_token"

# Test Basic auth
./downurl -input urls.txt --auth-basic "username:password"

# Test custom headers
cat > headers.txt <<EOF
X-API-Key: your-api-key
X-Custom-Header: value
EOF
./downurl -input urls.txt --headers-file headers.txt
```

### 3.5 Tests de Performance
```bash
# Test con muchas URLs (stress test)
# Generar 1000 URLs de prueba
for i in {1..1000}; do
    echo "https://example.com/file$i.js" >> large_test.txt
done

# Test con alta concurrencia
./downurl -input large_test.txt -workers 50

# Monitorear uso de memoria
/usr/bin/time -v ./downurl -input large_test.txt
```

### 3.6 Tests de Seguridad
```bash
# Test path traversal protection
cat > malicious_urls.txt <<EOF
https://example.com/../../../etc/passwd
https://example.com/test\x00.js
https://example.com//etc//hosts
EOF

./downurl -input malicious_urls.txt
# Verificar que los archivos se guarden en el directorio correcto
```

---

## 🏗️ Fase 4: Build y Package

### 4.1 Actualizar Versión en Código
```bash
# Actualizar constante de versión (si existe)
# Buscar en cmd/downurl/main.go o internal/config/config.go
grep -r "version" cmd/ internal/
```

### 4.2 Build para Todas las Plataformas
```bash
# Crear directorio de build
mkdir -p build/v1.1.0

# Script de build multi-plataforma
cat > build.sh <<'EOF'
#!/bin/bash
VERSION="1.1.0"
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT="build/v${VERSION}/downurl-${GOOS}-${GOARCH}"

    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi

    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$OUTPUT" cmd/downurl/main.go

    if [ $? -eq 0 ]; then
        echo "✓ Built: $OUTPUT"
    else
        echo "✗ Failed: $GOOS/$GOARCH"
        exit 1
    fi
done

echo "All builds completed successfully!"
EOF

chmod +x build.sh
./build.sh
```

### 4.3 Generar Checksums
```bash
# Generar SHA256 checksums
cd build/v1.1.0
sha256sum * > SHA256SUMS.txt
cat SHA256SUMS.txt
cd ../..
```

### 4.4 Comprimir Binarios
```bash
# Comprimir cada binario
cd build/v1.1.0
for file in downurl-*; do
    if [[ ! $file =~ \.tar\.gz$ ]] && [[ ! $file =~ \.txt$ ]]; then
        tar -czf "${file}.tar.gz" "$file"
        echo "Compressed: ${file}.tar.gz"
    fi
done
cd ../..
```

### 4.5 Verificar Binarios
```bash
# Probar cada binario (excepto Windows si estás en Linux/Mac)
./build/v1.1.0/downurl-linux-amd64 --version
./build/v1.1.0/downurl-darwin-amd64 --version
./build/v1.1.0/downurl-darwin-arm64 --version
```

---

## 🔖 Fase 5: Git Tag y Release

### 5.1 Commit Final
```bash
# Añadir todos los cambios de documentación
git add .
git status

# Crear commit de release
git commit -m "chore: prepare v1.1.0 release

- Update CHANGELOG with v1.1.0 changes
- Update README with new features
- Add release notes and documentation
- Organize documentation structure

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Push cambios
git push origin main
```

### 5.2 Crear Git Tag
```bash
# Crear tag anotado
git tag -a v1.1.0 -m "Release v1.1.0 - Usability Improvements

Major improvements to user experience:
✓ Animated progress bar with real-time updates
✓ Multiple input modes (stdin, single URL, file)
✓ Rate limiting with token bucket algorithm
✓ Watch & schedule modes
✓ Configuration file support
✓ Friendly error messages
✓ Storage organization modes

Critical bug fixes:
✓ Watch/scheduler recursion bug (no more leaks)
✓ Progress bar division by zero
✓ Path traversal security fixes

See RELEASE_NOTES_v1.1.0.md for full details."

# Verificar tag
git tag -l -n10 v1.1.0

# Push tag
git push origin v1.1.0
```

### 5.3 Crear GitHub Release
```bash
# Usar GitHub CLI para crear release
gh release create v1.1.0 \
    --title "v1.1.0 - Usability Improvements" \
    --notes-file RELEASE_NOTES_v1.1.0.md \
    build/v1.1.0/downurl-linux-amd64.tar.gz \
    build/v1.1.0/downurl-linux-arm64.tar.gz \
    build/v1.1.0/downurl-darwin-amd64.tar.gz \
    build/v1.1.0/downurl-darwin-arm64.tar.gz \
    build/v1.1.0/downurl-windows-amd64.exe.tar.gz \
    build/v1.1.0/SHA256SUMS.txt

# O crear release manualmente en GitHub:
# 1. Ir a https://github.com/llvch/downurl/releases/new
# 2. Seleccionar tag: v1.1.0
# 3. Título: "v1.1.0 - Usability Improvements"
# 4. Copiar contenido de RELEASE_NOTES_v1.1.0.md
# 5. Subir binarios y checksums
# 6. Marcar como "Latest release"
# 7. Publicar
```

---

## ✅ Fase 6: Verificación Post-Release

### 6.1 Verificar Release en GitHub
```bash
# Verificar que el release está visible
open https://github.com/llvch/downurl/releases/tag/v1.1.0

# Verificar que los binarios se pueden descargar
curl -LO https://github.com/llvch/downurl/releases/download/v1.1.0/downurl-linux-amd64.tar.gz

# Verificar checksum
sha256sum downurl-linux-amd64.tar.gz
```

### 6.2 Probar Instalación Fresca
```bash
# Simular instalación de usuario nuevo
mkdir /tmp/downurl-test
cd /tmp/downurl-test

# Descargar y extraer
curl -LO https://github.com/llvch/downurl/releases/download/v1.1.0/downurl-linux-amd64.tar.gz
tar -xzf downurl-linux-amd64.tar.gz

# Probar binario
./downurl-linux-amd64 --version
echo "https://example.com/test.js" | ./downurl-linux-amd64
```

### 6.3 Actualizar README Badges
```bash
# Actualizar badges en README.md (si aplica)
# - Version badge
# - Release date badge
# - Download count (automático en GitHub)
```

### 6.4 Announcement
```bash
# Opcional: Anunciar en redes sociales, foros, etc.
# - Twitter/X
# - Reddit (r/golang, r/hacking, etc.)
# - Hacker News
# - Discord/Slack communities
```

---

## 🔄 Fase 7: Post-Release Monitoring

### 7.1 Monitorear Issues (Primeras 48 horas)
- Revisar GitHub Issues regularmente
- Responder a reportes de bugs
- Documentar problemas conocidos
- Preparar hotfixes si es necesario

### 7.2 Preparar Hotfix (Si es necesario)
```bash
# Si se encuentra un bug crítico:
# 1. Crear rama de hotfix
git checkout -b hotfix/v1.1.1

# 2. Aplicar fix
# ... hacer cambios ...

# 3. Test rápido
go test ./...

# 4. Commit y tag
git commit -m "fix: critical bug in feature X"
git tag -a v1.1.1 -m "Hotfix v1.1.1"

# 5. Release v1.1.1
git push origin hotfix/v1.1.1
git push origin v1.1.1
```

### 7.3 Actualizar Documentación Post-Release
```bash
# Si se encuentran errores o mejoras en docs:
# - Actualizar en main
# - No requiere nueva release
# - Los usuarios verán cambios en GitHub
```

---

## 📊 Checklist Completo de Release

### Pre-Release
- [ ] Todos los tests pasan (unit, integration, race)
- [ ] go vet sin warnings
- [ ] Builds multi-plataforma exitosos
- [ ] Tests de seguridad pasados
- [ ] Tests de performance aceptables

### Documentación
- [ ] CHANGELOG.md actualizado con fecha
- [ ] README.md actualizado con nuevas features
- [ ] RELEASE_NOTES_v{version}.md creado
- [ ] docs/ organizado y actualizado
- [ ] Ejemplos probados y funcionando
- [ ] Links verificados

### Build
- [ ] Versión actualizada en código
- [ ] Binarios compilados para todas las plataformas
- [ ] Checksums generados (SHA256)
- [ ] Binarios comprimidos (.tar.gz)
- [ ] Binarios probados en plataformas objetivo

### Git
- [ ] Commit final con cambios de release
- [ ] Tag anotado creado (v{version})
- [ ] Push a origin/main
- [ ] Push del tag a origin

### GitHub Release
- [ ] Release creado en GitHub
- [ ] Título y descripción correctos
- [ ] Release notes completas
- [ ] Binarios subidos
- [ ] SHA256SUMS.txt subido
- [ ] Marcado como "Latest release"

### Post-Release
- [ ] Release verificado en GitHub
- [ ] Instalación fresca probada
- [ ] Documentación accesible
- [ ] Monitoring activo (48h)
- [ ] Issues respondidos

---

## 🚨 Rollback Plan

Si se descubre un bug crítico después del release:

### Opción 1: Hotfix Rápido (Preferido)
1. Crear rama `hotfix/v1.1.1`
2. Aplicar fix mínimo
3. Test rápido
4. Release v1.1.1
5. Comunicar a usuarios

### Opción 2: Rollback (Sólo si es crítico)
1. Despublicar release v1.1.0 en GitHub
2. Comunicar problema a usuarios
3. Recomendar volver a v1.0.0
4. Trabajar en fix
5. Re-release como v1.1.1

---

## 📞 Comunicación

### Durante Release
- Informar en GitHub Discussions (si aplica)
- Tweet/post sobre nuevo release
- Actualizar documentación

### Si hay problemas
- GitHub Issue inmediato
- Etiqueta "critical" o "bug"
- Comunicar ETA de fix
- Mantener transparencia

---

## 🎯 Métricas de Éxito

Una release exitosa debe tener:

- ✅ **0 bugs críticos** en primeras 48 horas
- ✅ **< 5 bugs menores** reportados
- ✅ **100% tests passing** en todas plataformas
- ✅ **Feedback positivo** de usuarios
- ✅ **Documentación clara** (pocas preguntas de uso)
- ✅ **Instalación sin problemas** en todas plataformas

---

## 📚 Recursos Útiles

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [GitHub Releases Guide](https://docs.github.com/en/repositories/releasing-projects-on-github)
- [Go Release Process](https://go.dev/doc/contribute#release)

---

**Última actualización**: 2025-11-17
**Versión del documento**: 1.0
**Propietario**: Release Manager
