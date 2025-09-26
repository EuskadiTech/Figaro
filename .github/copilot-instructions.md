# GitHub Copilot Instructions for Figaró

## Project Overview

Figaró is a comprehensive educational center management system designed to manage students, activities, materials, and resources in educational institutions. This is a complete rewrite of the original PHP application in Go, featuring improved performance, security, and maintainability.

**Key Technologies:**
- **Backend**: Go 1.24+ with Gin web framework
- **Database**: SQLite with custom migrations
- **Frontend**: HTML templates with Bootstrap CSS
- **Legacy**: PHP files in `www/` directory (being migrated to Go)
- **Authentication**: Custom session-based auth with bcrypt compatibility
- **Deployment**: Docker with multi-platform support

## Architecture & Code Structure

### Directory Structure
```
.
├── cmd/figaro/          # Main application entry point
├── internal/            # Private application code
│   ├── auth/           # Authentication middleware and logic
│   ├── database/       # Database connection and migrations
│   ├── handlers/       # HTTP handlers (controllers)
│   ├── models/         # Data structures and business logic
│   └── handlers/templates/ # HTML templates
├── pkg/                # Public packages
│   ├── config/         # Configuration management
│   └── utils/          # Utility functions
├── www/                # Legacy PHP files (being phased out)
└── data/               # Runtime data (SQLite DB, uploads, backups)
```

### Key Models
- **User**: System users with role-based permissions
- **Center**: Educational centers with working hours
- **Classroom**: Classrooms within centers
- **Activity**: Scheduled educational activities
- **Material**: Inventory management with photos and stock tracking

## Development Guidelines

### Go Code Standards
- Follow standard Go conventions and `gofmt` formatting
- Use meaningful variable and function names in English
- Add comments for exported functions and complex logic
- Implement proper error handling with descriptive messages
- Use struct tags for JSON and database field mapping
- Prefer composition over inheritance

### Database Patterns
- All database migrations are in `internal/database/migrations/`
- Use up/down migration pairs (e.g., `001_name.up.sql`, `001_name.down.sql`)
- Always use prepared statements to prevent SQL injection
- Database queries should be in handler functions, not in models
- Use SQLite-compatible SQL syntax

### HTTP Handler Patterns
- Place all HTTP handlers in `internal/handlers/`
- Use Gin context for request/response handling
- Implement proper HTTP status codes and error responses
- Validate input data before processing
- Use middleware for common functionality (auth, logging, etc.)
- Return JSON for API endpoints, render HTML templates for web pages

### Template Guidelines
- HTML templates are in `internal/handlers/templates/`
- Use Go template syntax with proper escaping
- Base template (`base.html`) contains common layout
- Include Bootstrap CSS classes for consistent styling
- Use semantic HTML with proper accessibility attributes

### Authentication & Security
- Session-based authentication using secure cookies
- Password hashing compatible with PHP bcrypt for migration
- Permission-based access control (check user permissions)
- Validate all user inputs and sanitize outputs
- Use HTTPS in production (handled by reverse proxy)

## Common Development Tasks

### Adding a New Feature
1. **Models**: Define data structures in `internal/models/models.go`
2. **Database**: Create migration files in `internal/database/migrations/`
3. **Handlers**: Add HTTP handlers in `internal/handlers/`
4. **Templates**: Create HTML templates in `internal/handlers/templates/`
5. **Routes**: Register routes in `cmd/figaro/main.go`

### Creating Database Migrations
```bash
# Create new migration files
touch internal/database/migrations/00X_description.up.sql
touch internal/database/migrations/00X_description.down.sql
```

### Adding New Routes
```go
// In cmd/figaro/main.go
authGroup.GET("/new-feature", h.NewFeatureHandler)
authGroup.POST("/new-feature", h.NewFeatureHandler)
```

### HTML Template Pattern
```html
{{define "content"}}
<div class="container-fluid py-4">
    <h1>Page Title</h1>
    <!-- Content here -->
</div>
{{end}}
```

## Building & Testing

### Development Commands
```bash
# Run in development mode
export GIN_MODE=debug
go run ./cmd/figaro

# Build binary
go build -o figaro ./cmd/figaro

# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Docker Commands
```bash
# Build Docker image
docker build -t figaro .

# Run with Docker
docker run -d -p 8080:8080 -v $(pwd)/data:/data figaro
```

### Production Deployment
- Use provided Docker images: `ghcr.io/euskaditech/figaro:latest`
- Mount `/data` volume for persistent storage
- Set environment variables: `PORT`, `HOST`, `DATA_DIR`, `GIN_MODE`
- Default credentials: username `demo`, password `demo`

## Legacy PHP Integration

### Current Status
- PHP files in `www/` directory are legacy code being migrated
- Go handlers should replace PHP functionality gradually
- Maintain data compatibility between PHP and Go implementations
- Priority is on Go implementation for new features

### Migration Strategy
- Keep existing PHP endpoints working during transition
- Implement Go equivalents with same functionality
- Update routes to point to Go handlers when ready
- Remove PHP files only after Go replacement is tested

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `HOST` | `0.0.0.0` | HTTP server bind address |
| `DATA_DIR` | `./data` | Directory for SQLite database and file storage |
| `GIN_MODE` | `release` | Gin framework mode (release/debug) |

## Troubleshooting

### Common Issues
- **Database locked**: Ensure only one instance is running
- **Permission denied**: Check file permissions on data directory
- **Login issues**: Verify bcrypt password hashes are compatible
- **Static files not loading**: Check embedded assets in binary
- **Docker volume issues**: Ensure proper volume mounting for `/data`

### Debugging
- Set `GIN_MODE=debug` for detailed request logging
- Check SQLite database with: `sqlite3 data/figaro.db`
- Review application logs for error details
- Use Go's built-in profiling tools for performance issues

## Contributing

### Code Review Guidelines
- Ensure all tests pass before submitting
- Follow Go coding standards and run `gofmt`
- Add appropriate error handling and logging
- Update documentation for new features
- Test database migrations both up and down
- Verify HTML templates render correctly
- Check that new features work across different centers/users

### Performance Considerations
- SQLite is suitable for small to medium deployments
- Consider connection pooling for high-traffic scenarios
- Optimize database queries and use appropriate indexes
- Cache frequently accessed data when appropriate
- Monitor memory usage with embedded templates

Remember: This is an educational management system, so prioritize data integrity, user experience, and security in all implementations.