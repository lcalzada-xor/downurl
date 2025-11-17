# 📚 Downurl Documentation Index

Complete documentation for Downurl - A high-performance concurrent file downloader.

---

## 🚀 Getting Started

**New to Downurl?** Start here:

- **[Getting Started Guide](user-guides/GETTING_STARTED.md)** - Quick start guide with basic usage
- **[README.md](../README.md)** - Project overview and quick reference

---

## 📖 User Guides

Documentation for using Downurl:

### Essential Guides
- **[Getting Started](user-guides/GETTING_STARTED.md)** - Installation and first steps
- **[Configuration Guide](user-guides/CONFIGURATION.md)** - Config files and environment variables
- **[Usage Guide](user-guides/USAGE.md)** *(Coming soon)* - Complete command-line reference
- **[Advanced Features](user-guides/ADVANCED.md)** *(Coming soon)* - Filtering, scanning, and automation

### Feature-Specific Guides
- **[Storage Modes](user-guides/GETTING_STARTED.md#storage-modes)** - File organization strategies
- **[Authentication](development/AUTH.md)** - Using Bearer, Basic, and custom auth
- **[Rate Limiting](user-guides/GETTING_STARTED.md)** - Control download speed
- **[Watch & Schedule](RELEASE_NOTES_v1.1.0.md#4-watch--schedule-modes)** - Automated downloads

---

## 🔧 Development Documentation

For contributors and developers:

### Architecture & Design
- **[Architecture](development/ARCHITECTURE.md)** - System design and components
- **[Features Implemented](development/FEATURES_IMPLEMENTED.md)** - Complete feature list
- **[Authentication Implementation](development/AUTH_IMPLEMENTATION.md)** - Auth system details

### Bug Bounty & Security
- **[Bug Bounty Features](development/BUGBOUNTY_FEATURES.md)** - Security research features
- **[Bug Fixes](development/BUGFIXES.md)** - Resolved issues
- **[Usability Improvements](development/USABILITY_IMPROVEMENTS.md)** - UX enhancements

### Planning Documents
- **[Bug Bounty Improvements Plan](development/BUGBOUNTY_IMPROVEMENTS_PLAN.md)** - Future security features
- **[Post-Crawling Features](development/POST_CRAWLING_FEATURES.md)** - Planned analysis features

---

## 🔄 Migration & Upgrades

Guides for upgrading between versions:

- **[Migration v0 to v1.0](migration/MIGRATION_v0_to_v1.0.md)** - Python to Go migration
- **[v1.1.0 Upgrade Guide](RELEASE_NOTES_v1.1.0.md#-upgrade-guide)** - Upgrading from v1.0.0

---

## 📋 Release Information

Release notes and planning:

### Current Release
- **[Release Notes v1.1.0](RELEASE_NOTES_v1.1.0.md)** - What's new in v1.1.0
- **[Release Plan v1.1.0](RELEASE_PLAN_v1.1.0.md)** - Release preparation checklist

### Previous Releases
- **[Release Notes v1.0.0](RELEASE_NOTES_v1.0.0.md)** - Initial Go release
- **[Changelog](../CHANGELOG.md)** - Complete version history

### Release Process
- **[Release Process](../RELEASE_PROCESS.md)** - How to prepare a new release

---

## 📚 Documentation by Topic

### Installation
- [Getting Started - Installation](user-guides/GETTING_STARTED.md#installation)
- [README - Installation](../README.md#installation)

### Basic Usage
- [Getting Started - Your First Download](user-guides/GETTING_STARTED.md#your-first-download)
- [Getting Started - Basic Usage Examples](user-guides/GETTING_STARTED.md#basic-usage-examples)

### Configuration
- [Configuration Guide](user-guides/CONFIGURATION.md)
- [Configuration File Format](user-guides/CONFIGURATION.md#configuration-file-format)
- [Environment Variables](user-guides/CONFIGURATION.md#environment-variables)

### Authentication
- [Auth Guide](development/AUTH.md)
- [Auth Implementation](development/AUTH_IMPLEMENTATION.md)

### Features
- [Storage Modes](user-guides/GETTING_STARTED.md#storage-modes)
- [Rate Limiting](user-guides/CONFIGURATION.md#ratelimit---rate-limiting)
- [Watch & Schedule](RELEASE_NOTES_v1.1.0.md#4-watch--schedule-modes)
- [Content Filtering](development/BUGBOUNTY_FEATURES.md)
- [Secret Scanning](development/BUGBOUNTY_FEATURES.md)

### Troubleshooting
- [Getting Started - Troubleshooting](user-guides/GETTING_STARTED.md#troubleshooting)
- [Known Issues](RELEASE_NOTES_v1.1.0.md#-known-issues)

---

## 🎯 Documentation by Use Case

### Bug Bounty / Security Research
- [Bug Bounty Features](development/BUGBOUNTY_FEATURES.md)
- [Security Research Config](user-guides/CONFIGURATION.md#example-1-bug-bounty-configuration)
- [Rate Limiting for Recon](user-guides/CONFIGURATION.md#ratelimit---rate-limiting)

### Web Archiving
- [Dated Storage Mode](user-guides/GETTING_STARTED.md#5-dated-mode)
- [Web Archiving Config](user-guides/CONFIGURATION.md#example-2-web-archiving)
- [Schedule Mode](RELEASE_NOTES_v1.1.0.md#schedule-mode-new-)

### API Data Collection
- [API Config Example](user-guides/CONFIGURATION.md#example-3-api-data-collection)
- [Authentication](development/AUTH.md)
- [Rate Limiting](user-guides/CONFIGURATION.md#ratelimit---rate-limiting)

### CDN Mirroring
- [Path Storage Mode](user-guides/GETTING_STARTED.md#2-path-mode)
- [High Concurrency](user-guides/GETTING_STARTED.md#increase-concurrency)

---

## 🔗 Quick Links

### Essential
- [Main README](../README.md)
- [Getting Started](user-guides/GETTING_STARTED.md)
- [Configuration](user-guides/CONFIGURATION.md)
- [Release Notes](RELEASE_NOTES_v1.1.0.md)

### Development
- [Architecture](development/ARCHITECTURE.md)
- [Contributing](development/ARCHITECTURE.md) *(Coming soon)*
- [Release Process](../RELEASE_PROCESS.md)

### Community
- [GitHub Issues](https://github.com/llvch/downurl/issues)
- [GitHub Discussions](https://github.com/llvch/downurl/discussions)
- [Changelog](../CHANGELOG.md)

---

## 📝 Documentation Structure

```
downurl/
├── README.md                      # Main project documentation
├── CHANGELOG.md                   # Version history
├── RELEASE_PROCESS.md            # Release preparation guide
│
├── docs/
│   ├── DOCUMENTATION_INDEX.md    # This file
│   ├── RELEASE_PLAN_v1.1.0.md   # Release planning
│   ├── RELEASE_NOTES_v1.1.0.md  # Release notes
│   ├── RELEASE_NOTES_v1.0.0.md  # Previous release
│   │
│   ├── user-guides/              # User-facing documentation
│   │   ├── GETTING_STARTED.md   # Quick start guide
│   │   ├── CONFIGURATION.md     # Config file guide
│   │   ├── USAGE.md             # Command reference (planned)
│   │   └── ADVANCED.md          # Advanced features (planned)
│   │
│   ├── development/              # Developer documentation
│   │   ├── ARCHITECTURE.md      # System architecture
│   │   ├── AUTH.md              # Authentication guide
│   │   ├── AUTH_IMPLEMENTATION.md
│   │   ├── BUGBOUNTY_FEATURES.md
│   │   ├── BUGBOUNTY_IMPROVEMENTS_PLAN.md
│   │   ├── BUGFIXES.md
│   │   ├── FEATURES_IMPLEMENTED.md
│   │   ├── POST_CRAWLING_FEATURES.md
│   │   └── USABILITY_IMPROVEMENTS.md
│   │
│   └── migration/                # Migration guides
│       └── MIGRATION_v0_to_v1.0.md
```

---

## 🤝 Contributing to Documentation

Found an error or want to improve the docs?

1. **Report Issues**: [GitHub Issues](https://github.com/llvch/downurl/issues)
2. **Suggest Improvements**: [GitHub Discussions](https://github.com/llvch/downurl/discussions)
3. **Submit PR**: Fork, edit, and submit a pull request

### Documentation Guidelines

- Use clear, concise language
- Include practical examples
- Test all code snippets
- Keep formatting consistent
- Add links between related docs

---

## 📊 Documentation Status

| Document | Status | Last Updated |
|----------|--------|--------------|
| README.md | ✅ Current | 2025-11-17 |
| Getting Started | ✅ Current | 2025-11-17 |
| Configuration | ✅ Current | 2025-11-17 |
| Release Notes v1.1.0 | ✅ Current | 2025-11-17 |
| Usage Guide | 📝 Planned | TBD |
| Advanced Features | 📝 Planned | TBD |
| Contributing Guide | 📝 Planned | TBD |

**Legend**:
- ✅ Current and up-to-date
- 📝 Planned (not yet created)
- ⚠️ Needs update
- ❌ Outdated

---

## 🔍 Search Tips

**Looking for specific topics?**

- **Installation**: See [Getting Started](user-guides/GETTING_STARTED.md#installation)
- **Command-line flags**: See [Getting Started - Essential Options](user-guides/GETTING_STARTED.md#essential-command-line-options)
- **Config file**: See [Configuration Guide](user-guides/CONFIGURATION.md)
- **Authentication**: See [Auth Guide](development/AUTH.md)
- **Bug fixes**: See [Changelog](../CHANGELOG.md) or [Release Notes](RELEASE_NOTES_v1.1.0.md)
- **Features**: See [Features Implemented](development/FEATURES_IMPLEMENTED.md)
- **Architecture**: See [Architecture Doc](development/ARCHITECTURE.md)

---

Need help? Check out [Getting Started](user-guides/GETTING_STARTED.md) or ask in [GitHub Discussions](https://github.com/llvch/downurl/discussions)!
