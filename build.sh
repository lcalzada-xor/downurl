#!/bin/bash

# Build script for Downurl v1.1.0
# Builds binaries for all supported platforms

set -e  # Exit on error

VERSION="1.1.0"
BUILD_DIR="build/v${VERSION}"
CMD_PATH="cmd/downurl/main.go"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Platform configurations: GOOS/GOARCH/output_name
PLATFORMS=(
    "linux/amd64/downurl-linux-amd64"
    "linux/arm64/downurl-linux-arm64"
    "darwin/amd64/downurl-darwin-amd64"
    "darwin/arm64/downurl-darwin-arm64"
    "windows/amd64/downurl-windows-amd64.exe"
)

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}  Downurl v${VERSION} Build Script${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Error: Go is not installed${NC}"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✓${NC} Go version: ${GO_VERSION}"
echo ""

# Create build directory
echo -e "${YELLOW}→${NC} Creating build directory: ${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH OUTPUT <<< "$PLATFORM"
    OUTPUT_PATH="${BUILD_DIR}/${OUTPUT}"

    echo -e "${YELLOW}→${NC} Building for ${GOOS}/${GOARCH}..."

    # Build with optimizations
    if GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-s -w" \
        -o "$OUTPUT_PATH" \
        "$CMD_PATH"; then

        # Get file size
        SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
        echo -e "${GREEN}✓${NC} Built: ${OUTPUT} (${SIZE})"
    else
        echo -e "${RED}✗${NC} Failed: ${GOOS}/${GOARCH}"
        exit 1
    fi
done

echo ""
echo -e "${YELLOW}→${NC} Generating SHA256 checksums..."
cd "${BUILD_DIR}"
sha256sum * > SHA256SUMS.txt
cd - > /dev/null
echo -e "${GREEN}✓${NC} SHA256SUMS.txt created"

echo ""
echo -e "${YELLOW}→${NC} Compressing binaries..."
cd "${BUILD_DIR}"
COMPRESSED=0
for file in downurl-*; do
    # Skip if already compressed or is the checksum file
    if [[ $file =~ \.tar\.gz$ ]] || [[ $file == "SHA256SUMS.txt" ]]; then
        continue
    fi

    # Compress
    if tar -czf "${file}.tar.gz" "$file"; then
        SIZE=$(du -h "${file}.tar.gz" | cut -f1)
        echo -e "${GREEN}✓${NC} Compressed: ${file}.tar.gz (${SIZE})"
        ((COMPRESSED++))
    else
        echo -e "${RED}✗${NC} Failed to compress: ${file}"
    fi
done
cd - > /dev/null

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}  Build Complete!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo -e "Built ${GREEN}${#PLATFORMS[@]}${NC} binaries"
echo -e "Compressed ${GREEN}${COMPRESSED}${NC} archives"
echo -e "Output directory: ${BLUE}${BUILD_DIR}/${NC}"
echo ""
echo -e "Files created:"
ls -lh "${BUILD_DIR}" | tail -n +2 | awk '{printf "  %s  %s\n", $5, $9}'
echo ""

# Optional: Verify builds
echo -e "${YELLOW}→${NC} Verifying checksums..."
cd "${BUILD_DIR}"
if sha256sum -c SHA256SUMS.txt > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} All checksums verified"
else
    echo -e "${YELLOW}⚠${NC} Checksum verification skipped (compressed files added)"
fi
cd - > /dev/null

echo ""
echo -e "${BLUE}Next steps:${NC}"
echo -e "  1. Test binaries: ${YELLOW}./build/v${VERSION}/downurl-linux-amd64 --version${NC}"
echo -e "  2. Commit changes: ${YELLOW}git add . && git commit -m 'chore: prepare v${VERSION} release'${NC}"
echo -e "  3. Create tag: ${YELLOW}git tag -a v${VERSION} -m 'Release v${VERSION}'${NC}"
echo -e "  4. Push: ${YELLOW}git push origin main && git push origin v${VERSION}${NC}"
echo -e "  5. Create release: ${YELLOW}gh release create v${VERSION} build/v${VERSION}/*.tar.gz${NC}"
echo ""
