# Go Blue Bird Starter Kit

A production-ready base template for Go web applications and REST APIs using Gin Framework, Air live-reloading, Dotenv configuration, Go `embed.FS`, resilient database connections, and a CLI utility.

---

## Features

- **Gin Web Framework:** High-performance HTTP server with segregated `/` (Web HTML) and `/api` (JSON REST API) routes.
- **Live Reload with Air:** Instant server updates on source code modifications.
- **Embedded Assets (`embed.FS`):** Compiles static files (`public/`) and HTML templates (`templates/`) into a single executable binary.
- **Modular Core Architecture:**
  - `config/config.go`: Environment variables with helper functions (`Get`, `GetInt`, `GetBool`, `IsDev`, `IsProd`).
  - `core/database.go`: Connection retry wrapper for MySQL, PostgreSQL, and SQLite with UTF-8 (`utf8mb4`) support and health check pinging.
  - `core/responses.go`: Standardized JSON API responses (`Success`, `Created`, `Error`, `ValidationError`, `Paginated`).
  - `core/request.go`: Extractor functions for query strings, integers, path parameters, and pagination offsets.
  - `core/validate.go`: Request binding wrapper formatting `go-playground/validator/v10` validation error messages.
  - `core/hash.go`: Password hashing and comparison using Bcrypt.
- **Smart CLI Tool (`cli/main.go`):** Module renaming utility and quick database table/column/query inspector.

---

## Installing Air (Live Reload)

Air is a live-reloading utility for Go applications.

### Windows (PowerShell)
```powershell
go install github.com/air-verse/air@latest
```
Ensure your Go binary path (`$env:USERPROFILE\go\bin`) is added to your PATH environment variable.

### Linux & macOS (Terminal)
```bash
# Option 1: Via Go install
go install github.com/air-verse/air@latest

# Option 2: Via binary installer script
curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# macOS via Homebrew
brew install air
```

---

## Quick Start

### 1. Environment Setup
Copy the `.env.example` file to `.env`:
```bash
cp .env.example .env
```

### 2. Run in Development Mode with Air
```bash
air
```

### 3. Run with Standard Go
```bash
go run main.go
```

The application will be accessible at:
- Web Interface: `http://localhost:8080`
- API Health Check: `http://localhost:8080/api/health`
- API Info Endpoint: `http://localhost:8080/api/info`

---

## Production Build

To compile a single, standalone binary containing all HTML templates and static assets:
```bash
go build -o bin/app main.go
```

To run the production binary:
```bash
./bin/app
```

---

## CLI Utility

The included CLI tool in `cli/main.go` provides commands for project management and database inspection.

### Renaming the Module for a New Project
When starting a new project derived from this template, run:
```bash
go run cli/main.go rename github.com/your-username/your-new-repo
```

### Database Inspection Commands
- **List all tables:**
  ```bash
  go run cli/main.go db tables
  ```
- **List table columns:**
  ```bash
  go run cli/main.go db columns users
  ```
- **Run quick select query:**
  ```bash
  go run cli/main.go db query users 5
  ```

---

## Project Directory Layout

```
.
├── .air.toml               # Air live-reload configuration
├── .env.example            # Environment variables template
├── .env                    # Local environment variables
├── .gitignore              # Git ignore rules
├── go.mod                  # Go module definition
├── main.go                 # Application entry point
├── config/
│   └── config.go           # Environment variables loader
├── core/
│   ├── database.go         # Database connection wrapper with retry logic
│   ├── hash.go             # Bcrypt password hashing
│   ├── helpers.go          # General utility functions
│   ├── request.go          # Query & path parameter helpers
│   ├── responses.go        # Unified API JSON responses
│   └── validate.go         # Request validation & error formatting
├── routes/
│   ├── routes.go           # Router setup & middleware pipeline
│   ├── api/
│   │   └── api.go          # REST API endpoints (/api/)
│   └── web/
│       └── web.go          # Web endpoints rendering templates
├── templates/
│   ├── embed.go            # Embedded HTML template system
│   ├── layouts/
│   │   └── base.html       # Base layout template
│   └── pages/
│       ├── index.html      # Home page
│       └── about.html      # About page
├── public/
│   ├── embed.go            # Embedded static asset system
│   ├── css/
│   │   └── style.css       # CSS styling
│   └── js/
│       └── app.js          # Client-side JavaScript
└── cli/
    └── main.go             # Smart CLI tool
```
