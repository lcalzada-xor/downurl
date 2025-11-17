# ✅ Release Checklist v1.1.0

Quick reference checklist for releasing Downurl v1.1.0.

---

## 📋 Pre-Release

### Code Quality
- [x] All unit tests passing
- [x] Race detector clean (`go test -race`)
- [x] `go vet` clean
- [x] No memory leaks
- [x] Code reviewed
- [ ] Test coverage > 70%

### Features Complete
- [x] Enhanced UI (progress bar, colors, tables)
- [x] Multiple input modes (stdin, single URL, file)
- [x] Rate limiting
- [x] Watch & schedule modes
- [x] Configuration file support
- [x] Storage organization modes
- [x] Friendly error messages

### Bug Fixes
- [x] Watch/scheduler recursion bug fixed
- [x] Progress bar division by zero fixed
- [x] Path traversal vulnerability fixed (v1.0.0)
- [x] Hostname sanitization fixed

---

## 📝 Documentation

### Main Documentation
- [ ] README.md updated with v1.1.0 features
- [x] CHANGELOG.md updated with release date
- [x] RELEASE_NOTES_v1.1.0.md created
- [x] RELEASE_PROCESS.md created
- [x] DOCUMENTATION_INDEX.md created

### User Guides
- [x] GETTING_STARTED.md created
- [x] CONFIGURATION.md created
- [ ] USAGE.md created (optional)
- [ ] ADVANCED.md created (optional)

### Verification
- [ ] All examples tested and working
- [ ] All links verified
- [ ] Code snippets executable
- [ ] Badges updated

---

## 🧪 Testing

### Functional Tests
- [ ] Basic download from file
- [ ] Download from stdin
- [ ] Single URL download
- [ ] Watch mode (30+ minutes)
- [ ] Schedule mode
- [ ] Rate limiting accuracy
- [ ] All 5 storage modes
- [ ] Configuration file loading
- [ ] Authentication (Bearer, Basic, Custom)

### Platform Tests
- [ ] Linux AMD64
- [ ] Linux ARM64
- [ ] macOS Intel
- [ ] macOS Apple Silicon
- [ ] Windows 10/11

### Security Tests
- [x] Path traversal attempts blocked
- [x] Malicious URLs sanitized
- [x] Null byte injection handled
- [ ] Fuzzing tests (optional)

### Performance Tests
- [ ] 1000+ URLs download
- [ ] High concurrency (50+ workers)
- [ ] Memory profiling (long-running)
- [ ] CPU profiling

---

## 🏗️ Build

### Preparation
- [ ] Update version in code (if applicable)
- [ ] Create `build/v1.1.0/` directory
- [ ] Verify Go version (1.24.9)

### Build All Platforms
- [ ] Linux AMD64 (`GOOS=linux GOARCH=amd64`)
- [ ] Linux ARM64 (`GOOS=linux GOARCH=arm64`)
- [ ] macOS AMD64 (`GOOS=darwin GOARCH=amd64`)
- [ ] macOS ARM64 (`GOOS=darwin GOARCH=arm64`)
- [ ] Windows AMD64 (`GOOS=windows GOARCH=amd64`)

### Packaging
- [ ] Generate SHA256 checksums
- [ ] Compress binaries (tar.gz)
- [ ] Test each binary
- [ ] Verify checksums

---

## 🔖 Git & GitHub

### Git Operations
- [ ] All documentation changes committed
- [ ] Commit message follows convention
- [ ] Push to `origin/main`
- [ ] Create annotated tag `v1.1.0`
- [ ] Push tag to `origin`

### GitHub Release
- [ ] Create release on GitHub
- [ ] Title: "v1.1.0 - Usability Improvements"
- [ ] Copy release notes from RELEASE_NOTES_v1.1.0.md
- [ ] Upload all binaries (.tar.gz)
- [ ] Upload SHA256SUMS.txt
- [ ] Mark as "Latest release"
- [ ] Publish release

---

## ✅ Post-Release

### Verification
- [ ] Release visible on GitHub
- [ ] Binaries downloadable
- [ ] Checksums correct
- [ ] Fresh install test (Linux)
- [ ] Fresh install test (macOS)
- [ ] Fresh install test (Windows)

### Documentation
- [ ] Documentation accessible
- [ ] Links working
- [ ] README badges updated

### Announcement (Optional)
- [ ] GitHub Discussions post
- [ ] Twitter/X announcement
- [ ] Reddit post (r/golang)
- [ ] Hacker News submission

### Monitoring
- [ ] Watch GitHub Issues (48 hours)
- [ ] Respond to bug reports
- [ ] Document known issues
- [ ] Prepare hotfix if needed

---

## 🎯 Success Criteria

Release is successful if:
- [x] All tests passing
- [ ] 0 critical bugs in 48 hours
- [ ] < 5 minor bugs reported
- [ ] Positive user feedback
- [ ] Documentation clear
- [ ] All platforms working

---

## 🚨 Rollback Triggers

Immediately rollback or hotfix if:
- Critical security vulnerability discovered
- Data loss bug found
- Application crashes on startup
- Major functionality broken
- Cannot install on any platform

---

## 📞 Emergency Contacts

**Release Manager**: [Your Name]
**Backup**: [Backup Name]

**Communication Channels**:
- GitHub Issues: https://github.com/llvch/downurl/issues
- Email: [Your Email]

---

## 📊 Timeline

| Phase | Duration | Status |
|-------|----------|--------|
| Pre-Release Prep | 2 days | ✅ Complete |
| Documentation | 1 day | 🔄 In Progress |
| Testing & QA | 2 days | ⏳ Pending |
| Build & Package | 1 day | ⏳ Pending |
| Release | 1 day | ⏳ Pending |
| Post-Release Monitor | 3 days | ⏳ Pending |

**Total**: 7-10 days

---

## 🔄 Current Status

**Last Updated**: 2025-11-17
**Current Phase**: Documentation
**Next Steps**:
1. Update README.md
2. Complete testing
3. Build binaries

**Blockers**: None

**Notes**:
- All critical features implemented
- Critical bugs fixed
- Documentation in progress

---

## 📝 Quick Commands

### Run Full Test Suite
```bash
go test ./... -v -race -cover
go vet ./...
```

### Build All Platforms
```bash
./build.sh  # See RELEASE_PROCESS.md
```

### Create Release
```bash
# Commit
git add .
git commit -m "chore: prepare v1.1.0 release"
git push origin main

# Tag
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# GitHub Release
gh release create v1.1.0 \
    --title "v1.1.0 - Usability Improvements" \
    --notes-file docs/RELEASE_NOTES_v1.1.0.md \
    build/v1.1.0/*.tar.gz \
    build/v1.1.0/SHA256SUMS.txt
```

---

For detailed instructions, see [RELEASE_PROCESS.md](RELEASE_PROCESS.md)
