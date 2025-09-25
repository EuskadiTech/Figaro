# Changelog

All notable changes to Figaro will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-09-25

### 🎉 Major Release - Go Rewrite

This is a complete rewrite of Figaro from PHP to Go, providing significant improvements in performance, security, and maintainability.

### Added
- **Go Implementation**: Complete rewrite in Go 1.21+ with Gin web framework
- **SQLite Database**: Replaced JSON file storage with SQLite database
- **Embedded Assets**: All static files (CSS, JS, images) are embedded in the binary
- **Single Binary Distribution**: Self-contained executable with no external dependencies
- **Multi-platform Builds**: Support for Linux (arm64, amd64), Windows (amd64), and macOS (amd64, arm64)
- **Docker Support**: Multi-platform Docker images built from scratch for minimal size
- **Database Migrations**: Automatic database schema management and migrations
- **PHP Compatibility**: Seamless migration support for existing PHP password hashes
- **Comprehensive Documentation**: Detailed README with setup, deployment, and development guides
- **Build Pipeline**: GitHub Actions workflow for automated multi-platform builds and releases

### Core Features
- **User Management**: Role-based access control with granular permissions
- **Authentication**: Secure login with username/password or QR code support
- **Center Management**: Multi-center support with classroom organization
- **Materials Inventory**: Track materials with quantities, units, and photos
- **Activities Management**: Schedule and manage educational activities
- **File Management**: Centralized document storage and access control
- **Admin Panel**: Complete administration interface for users and centers
- **Session Management**: Secure cookie-based sessions with proper expiration

### Technical Improvements
- **Performance**: Significant performance improvements over PHP version
- **Security**: Enhanced security with bcrypt password hashing and secure sessions
- **Maintainability**: Clean Go codebase with proper package structure
- **Deployment**: Single binary deployment with embedded assets
- **Monitoring**: Structured logging and error handling
- **Scalability**: Better concurrent request handling

### Infrastructure
- **GitHub Actions**: Automated CI/CD pipeline for builds and releases
- **Docker**: Multi-platform container images (linux/amd64, linux/arm64)
- **Binary Releases**: Pre-built binaries for multiple platforms
- **Documentation**: Comprehensive documentation and examples

### Compatibility
- **Data Migration**: Automatic migration from PHP JSON files to SQLite
- **Password Compatibility**: Existing PHP bcrypt password hashes work seamlessly
- **UI Consistency**: Maintains the same user interface and user experience
- **Feature Parity**: All features from PHP version are preserved

### Breaking Changes
- **Database**: Moved from JSON files to SQLite database (automatic migration provided)
- **Deployment**: New deployment model with single binary instead of PHP files
- **Configuration**: Environment variable based configuration instead of PHP config files

## [1.x.x] - Legacy PHP Version

The previous PHP-based versions of Figaro. See the `php-legacy` branch for historical releases.

### Legacy Features
- PHP 8.2+ with Apache
- JSON file-based data storage
- Traditional web application deployment
- Basic authentication and session management

---

## Migration Guide from PHP Version

### Automatic Migration
The Go version automatically migrates data from the existing PHP version:

1. **Users**: User accounts and permissions are imported from JSON files
2. **Centers**: Center configurations are migrated to the database
3. **Materials**: Material inventory is transferred to SQLite
4. **Activities**: Existing activities are imported and preserved
5. **Passwords**: PHP password hashes remain compatible

### Manual Steps
1. **Backup**: Always backup your existing `data/` directory
2. **Deploy**: Deploy the new Go binary or Docker image
3. **Verify**: Check that all data has been migrated correctly
4. **Test**: Test login and functionality with existing users

### Configuration Changes
- **Environment Variables**: Configuration now uses environment variables instead of PHP config files
- **Data Directory**: The `DATA_DIR` environment variable controls data location
- **Port Configuration**: Use `PORT` environment variable instead of Apache configuration

For detailed migration instructions, see the [Migration Guide](docs/MIGRATION.md).