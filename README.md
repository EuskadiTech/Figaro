# Figaró - Educational Center Management System

![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Build Status](https://github.com/EuskadiTech/Figaro/actions/workflows/build.yml/badge.svg)

**Figaró** is a comprehensive educational center management system designed to help manage students, activities, materials, and resources in educational institutions. This is a complete rewrite of the original PHP application in Go, featuring improved performance, security, and maintainability.

![Figaro Homepage](https://github.com/user-attachments/assets/e486fe66-a46f-4589-8826-70e9d05204ae)

## ✨ Features

### 🏫 Center Management
- **Multi-center support** - Manage multiple educational centers from a single installation
- **Classroom organization** - Organize classrooms within each center
- **Working hours configuration** - Set operating hours per center and day of the week
- **Location-based access control** - Users can select their center and classroom

### 👥 User Management  
- **Role-based access control** - Granular permissions for different user roles
- **Admin panel** - Complete user administration interface
- **Authentication** - Secure login with username/password or QR code
- **Session management** - Persistent user sessions with proper security

### 📚 Materials Inventory
- **Inventory tracking** - Track available quantities and minimum stock levels
- **Photo management** - Visual identification of materials with images
- **Multi-unit support** - Different units of measurement (pieces, packages, etc.)
- **Center-specific inventory** - Separate inventories per center

### 📅 Activities Management
- **Event scheduling** - Create and manage educational activities
- **Global and center-specific activities** - Support for both types
- **Time conflict detection** - Automatic validation against working hours
- **Activity descriptions** - Rich text descriptions and details

### 🗃️ File Management
- **Document storage** - Centralized file storage and management
- **Access controls** - Permission-based file access
- **File organization** - Structured file hierarchy

### 🔒 Security Features
- **PHP password compatibility** - Seamless migration from PHP bcrypt hashes
- **QR code authentication** - Alternative login method
- **Session security** - Secure cookie-based sessions
- **Permission checks** - Every action verified against user permissions

## 🚀 Quick Start

### Using Docker (Recommended)

```bash
# Pull and run the latest image
docker run -d \
  --name figaro \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  ghcr.io/euskaditech/figaro:latest
```

### Using Pre-built Binaries

1. Download the latest binary for your platform from [Releases](https://github.com/EuskadiTech/Figaro/releases)
2. Make it executable (Linux/macOS): `chmod +x figaro-*`
3. Run the binary: `./figaro-linux-amd64` (or appropriate for your platform)
4. Open http://localhost:8080 in your browser

### Building from Source

```bash
# Clone the repository
git clone https://github.com/EuskadiTech/Figaro.git
cd Figaro

# Build the application
go build -o figaro ./cmd/figaro

# Run the application
./figaro
```

## 📖 Usage

### First Login
- **Username:** `demo`
- **Password:** `demo`

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `HOST` | `0.0.0.0` | HTTP server bind address |
| `DATA_DIR` | `./data` | Directory for SQLite database and file storage |
| `GIN_MODE` | `release` | Gin framework mode (release/debug) |

### Data Directory Structure

```
data/
├── figaro.db          # SQLite database
├── uploads/           # User uploaded files
└── backups/           # Database backups (if enabled)
```

## 🏗️ Architecture

### Technology Stack
- **Backend:** Go 1.21+ with Gin web framework
- **Database:** SQLite with proper migrations
- **Frontend:** Server-side rendered HTML templates
- **Assets:** Embedded static files (CSS, JS, images)
- **Authentication:** bcrypt password hashing with PHP compatibility
- **Build:** Multi-stage Docker builds for minimal image size

### Project Structure

```
.
├── cmd/figaro/                 # Application entry point
├── internal/
│   ├── auth/                  # Authentication and authorization
│   ├── database/              # Database connection and migrations
│   ├── handlers/              # HTTP request handlers
│   │   ├── static/           # Embedded static assets
│   │   └── templates/        # HTML templates
│   └── models/               # Data models
├── pkg/
│   ├── config/               # Configuration management
│   └── utils/                # Utility functions
├── migrations/               # Database migration files
└── docs/                     # Documentation
```

### Database Schema

The application uses SQLite with the following main tables:

- **users** - User accounts and authentication
- **user_permissions** - Role-based permissions
- **centers** - Educational centers
- **center_working_hours** - Operating schedules
- **classrooms** - Classroom organization
- **materials** - Inventory management
- **activities** - Event and activity tracking

## 🐳 Docker Deployment

### Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'

services:
  figaro:
    image: ghcr.io/euskaditech/figaro:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - PORT=8080
      - HOST=0.0.0.0
      - GIN_MODE=release
    restart: unless-stopped
```

Run with: `docker-compose up -d`

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: figaro
spec:
  replicas: 1
  selector:
    matchLabels:
      app: figaro
  template:
    metadata:
      labels:
        app: figaro
    spec:
      containers:
      - name: figaro
        image: ghcr.io/euskaditech/figaro:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATA_DIR
          value: "/data"
        volumeMounts:
        - name: data-volume
          mountPath: /data
      volumes:
      - name: data-volume
        persistentVolumeClaim:
          claimName: figaro-data-pvc
```

## 🔧 Development

### Prerequisites
- Go 1.21 or higher
- SQLite development libraries
- Git

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/EuskadiTech/Figaro.git
cd Figaro

# Install dependencies
go mod download

# Run in development mode
export GIN_MODE=debug
go run ./cmd/figaro
```

### Database Migrations

Migrations are automatically run on application startup. To manually inspect the database:

```bash
# Connect to SQLite database
sqlite3 data/figaro.db

# List tables
.tables

# Describe table structure
.schema users
```

### Adding New Features

1. **Models** - Add data structures in `internal/models/`
2. **Database** - Update migration files in `internal/database/migrations/`
3. **Handlers** - Add HTTP handlers in `internal/handlers/`
4. **Templates** - Add HTML templates in `internal/handlers/templates/`
5. **Routes** - Register routes in `cmd/figaro/main.go`

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 API Documentation

The application primarily uses server-side rendering, but provides some API endpoints:

### Authentication
- `GET /login` - Login page
- `POST /login` - Process login (form data or JSON)
- `GET /logout` - Logout and redirect

### Static Assets
- `GET /static/*` - Serve embedded static files

### Protected Routes (require authentication)
- `GET /` - Dashboard/homepage
- `GET /materiales` - Materials inventory
- `GET /actividades` - Activities management  
- `GET /admin` - Administration panel
- `GET /elegir_centro` - Center/classroom selection

## 🔒 Security Considerations

### Password Security
- Uses bcrypt for password hashing
- Compatible with PHP password_hash() output
- Supports both `$2y$` and `$2a$` bcrypt variants

### Session Management
- HTTP-only cookies for session data
- Secure session validation
- Automatic session expiration

### Access Control
- Permission-based route protection
- Role-based access control (RBAC)
- Center/classroom isolation

### Data Protection
- SQLite with WAL mode for concurrent access
- Prepared statements prevent SQL injection
- Input validation and sanitization
- CSRF protection on forms

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Ensure all tests pass
6. Submit a pull request

### Code Style
- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `golint` and `go vet`
- Add comments for exported functions
- Write tests for new functionality

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Getting Help
- [GitHub Issues](https://github.com/EuskadiTech/Figaro/issues) - Bug reports and feature requests
- [GitHub Discussions](https://github.com/EuskadiTech/Figaro/discussions) - Questions and community support

### Common Issues

**Q: The application won't start**
A: Check that the data directory is writable and SQLite is available.

**Q: Can't login with existing users**
A: The Go version is compatible with PHP bcrypt hashes. Ensure the database migration completed successfully.

**Q: Static files not loading**
A: Static files are embedded in the binary. Ensure you're using a properly built binary or the correct Docker image.

**Q: Database errors**
A: Check SQLite is installed and the data directory has proper permissions.

## 🎯 Roadmap

- [ ] REST API for mobile applications
- [ ] Advanced reporting and analytics
- [ ] Email notification system
- [ ] Backup and restore functionality
- [ ] Multi-language support (i18n)
- [ ] Advanced file management features
- [ ] Calendar integration
- [ ] Mobile-responsive UI improvements
- [ ] Advanced user management (LDAP/SSO)
- [ ] Performance monitoring and metrics

## 📊 Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.

---

**Made with ❤️ by the EuskadiTech team**