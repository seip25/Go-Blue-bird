# Go Blue Bird Starter Kit

A production-ready, ultra-high-performance base architecture for Go web applications and REST APIs. Built on the **Gin Framework**, **Air** live-reloading, **Go `embed.FS`**, resilient multi-driver database connections (with SQLite WAL mode), in-memory RAM caching, structured validation, and an automated deployment CLI tool.

---

## ⚡ Key Highlights & Benchmarks

- **Blazing Fast Throughput:** Capable of **55,000+ requests/sec** on SQLite WAL queries and **57,000+ req/s** on in-memory RAM cache with average latencias under **0.05 ms** ($c=20$).
- **Single Static Binary (21 MB / 13 MB stripped):** All HTML templates (`templates/`) and static CSS/JS (`public/`) are embedded into a standalone executable via `embed.FS`. Zero external file dependencies in production.
- **Micro Memory Footprint:** Consumes merely **21 MB** at idle and **~39 MB** under peak stress of 57,000 req/s.
- **Automated VPS Deployment CLI:** Generate binaries and ready-to-paste `systemd_service.txt`, `nginx_reverse_proxy.txt`, and deployment checklists with a single command.

---

## 🚀 Features

- **Gin Web Framework:** Clean segregation between `/` (SSR HTML pages) and `/api` (JSON REST APIs).
- **Environment Modes:** Automatically switches between `gin.DebugMode` (with colored request logging in development) and `gin.ReleaseMode` (silent, maximum throughput in production) via `APP_ENV`.
- **Live Reload with Air:** Instant server updates on source code modifications during development.
- **Resilient Multi-Driver Database Core (`core/database.go`):**
  - **SQLite:** Pre-configured with `_journal_mode=WAL`, `_busy_timeout=5000`, and `_synchronous=NORMAL` to prevent database locks under high concurrency.
  - **MySQL & PostgreSQL:** Built-in connection pool settings, UTF-8 (`utf8mb4`), and automatic connection retry logic with backoff.
- **In-Memory RAM Cache (`core/cache.go`):** Thread-safe (`sync.RWMutex`) sub-millisecond memory cache with configurable TTL expiration.
- **Validation & Request Binding (`core/validate.go`):** Schema validation using `go-playground/validator/v10` tags with standardized error formatting.
- **Unified JSON Responses (`core/responses.go`):** Consistent API response contracts (`RespondSuccess`, `RespondCreated`, `RespondError`, `RespondValidationError`).
- **Security & Password Hashing (`core/hash.go`):** Bcrypt hashing and verification.
- **Smart CLI Tool (`cli/main.go`):** Module renaming, database schema inspector, automated binary compilation, and systemd/Nginx config generator.

---

## 📦 Quick Start

### 1. Environment Setup
Copy the `.env.example` file to `.env`:
```bash
cp .env.example .env
```

Configuration variables:
```ini
APP_NAME=GoBlueBird
APP_ENV=development      # Use 'production' for maximum performance
PORT=8080

DB_DRIVER=sqlite         # sqlite, mysql, postgres
DB_NAME=app_db.sqlite
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME_MIN=5

API_KEY=sample_secret_key_12345
```

### 2. Run in Development Mode with Air (Live Reload)
```bash
air
```

### 3. Run with Standard Go
```bash
go run main.go
```

The application will be accessible at:
- **Web UI:** `http://localhost:8080/`
- **About Page:** `http://localhost:8080/about`
- **API Health:** `http://localhost:8080/api/health`
- **API Info:** `http://localhost:8080/api/info`

---

## 🛠️ In-Memory RAM Cache Usage

The framework includes a high-speed, thread-safe RAM cache in `core/cache.go`:

```go
import (
    "time"
    "github.com/seip25/Go-Blue-bird/core"
)

// Store a value in memory with a 5-minute TTL
core.CacheSet("user:123:profile", userData, 5*time.Minute)

// Retrieve a value
if value, found := core.CacheGet("user:123:profile"); found {
    profile := value.(*UserProfile)
    // Use cached profile...
}

// Invalidate / Delete a key
core.CacheDelete("user:123:profile")

// Clear entire cache
core.CacheClear()
```

---

## 🎨 HTML Rendering with Layouts (`core.Render`)

Go Blue Bird automatically discovers all `.html` pages inside `templates/pages/` (including subdirectories like `pages/auth/login.html`), pairs them with `layouts/base.html` in isolated scopes to avoid block collisions, and pre-compiles them at startup:

```go
package web

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/seip25/Go-Blue-bird/core"
)

func HomeHandler(c *gin.Context) {
    core.Render(c, http.StatusOK, "index", gin.H{
        "Title": "Go Blue Bird - Home",
    })
}

func AboutHandler(c *gin.Context) {
    core.Render(c, http.StatusOK, "about", gin.H{
        "Title": "Go Blue Bird - About",
    })
}
```

You can pass clean template names (`"index"`, `"about"`, `"auth/login"`) or filenames with extensions (`"index.html"`).

---

## 🛡️ Request Validation & API Responses

### Defining and Validating Schemas

Define your DTO struct with validator tags:

```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=30"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}
```

In your handler, use `core.BindJSONAndValidate`:

```go
func CreateUserHandler(c *gin.Context) {
    var req CreateUserRequest
    
    // Automatically binds JSON and returns 422 with formatted errors if validation fails
    if !core.BindJSONAndValidate(c, &req) {
        return
    }

    hashedPassword, err := core.HashPassword(req.Password)
    if err != nil {
        core.RespondError(c, http.StatusInternalServerError, "Failed to hash password", err.Error())
        return
    }

    core.RespondCreated(c, "User created successfully", gin.H{
        "username": req.Username,
        "email":    req.Email,
    })
}
```

### Standardized JSON Responses

```go
// 200 OK
core.RespondSuccess(c, "Data retrieved successfully", data)

// 201 Created
core.RespondCreated(c, "Resource created", newResource)

// 400 / 500 Error
core.RespondError(c, http.StatusBadRequest, "Invalid request", "Detailed error message")

// 422 Validation Error
core.RespondValidationError(c, validationErrors)
```

---

## ⚡ Smart CLI Tool (`cli/main.go`)

### 1. Automated Production Build & VPS Config Generation
Compile a standalone binary into the `builds/` directory and generate ready-to-use VPS configurations:

```bash
# Compile for current OS / Architecture
go run cli/main.go build

# Cross-compile for Linux VPS (from Mac/Windows)
go run cli/main.go build linux amd64

# Specify custom VPS deploy directory
go run cli/main.go build linux amd64 --dir=/var/www/my-app
```

The CLI automatically outputs into `builds/`:
1. `builds/gobluebird` (Optimized, stripped binary with `-s -w`).
2. `builds/systemd_service.txt` (Complete `/etc/systemd/system/gobluebird.service` unit file).
3. `builds/nginx_reverse_proxy.txt` (Complete Nginx reverse proxy block with WebSocket & SSL Certbot setup).
4. `builds/deploy_guide.txt` (3-minute step-by-step VPS deployment checklist).

### 2. Automatic Systemd Installation (On Linux VPS)
When running directly on your Ubuntu/Debian server:
```bash
sudo go run cli/main.go service --install
```

### 3. Renaming the Module for a New Project
To rebrand the module path across all files:
```bash
go run cli/main.go rename github.com/your-username/your-new-repo
```

### 4. Database Inspector Commands
- **List tables:**
  ```bash
  go run cli/main.go db tables
  ```
- **Inspect table columns:**
  ```bash
  go run cli/main.go db columns users
  ```
- **Query table rows:**
  ```bash
  go run cli/main.go db query users 10
  ```

---

## 🚀 3-Minute VPS Deployment Guide

1. **Build locally:**
   ```bash
   go run cli/main.go build linux amd64
   ```

2. **Copy files to VPS:**
   ```bash
   scp builds/gobluebird user@vps:/var/www/gobluebird/
   scp .env user@vps:/var/www/gobluebird/
   ```

3. **Configure Systemd & Nginx:**
   Copy the prepared text from `builds/systemd_service.txt` into `/etc/systemd/system/gobluebird.service` and `builds/nginx_reverse_proxy.txt` into `/etc/nginx/sites-available/gobluebird`.

4. **Start the service:**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now gobluebird
   sudo systemctl reload nginx
   ```

---

## 📂 Project Directory Layout

```
.
├── .air.toml               # Air live-reload configuration
├── .env.example            # Environment variables template
├── .env                    # Local environment variables
├── .gitignore              # Git ignore rules
├── go.mod                  # Go module definition
├── main.go                 # Application entry point & mode switch
├── config/
│   └── config.go           # Dotenv environment variables loader
├── core/
│   ├── cache.go            # Thread-safe in-memory RAM cache (sync.RWMutex)
│   ├── database.go         # Resilient DB wrapper (SQLite WAL, MySQL, Postgres)
│   ├── hash.go             # Bcrypt password hashing
│   ├── helpers.go          # General utility functions
│   ├── render.go           # Automated template scanner & core.Render helper
│   ├── request.go          # Query, path & pagination extractors
│   ├── responses.go        # Unified JSON response contracts
│   └── validate.go         # Request validation & error formatting
├── routes/
│   ├── routes.go           # Router setup, CORS & recovery pipeline
│   ├── api/
│   │   └── api.go          # REST API endpoints (/api/)
│   └── web/
│       └── web.go          # Web endpoints rendering templates
├── templates/
│   ├── embed.go            # Embedded HTML template system (embed.FS)
│   ├── layouts/
│   │   └── base.html       # Base layout template
│   └── pages/
│       ├── index.html      # Home page
│       └── about.html      # About page
├── public/
│   ├── embed.go            # Embedded static asset system (embed.FS)
│   ├── css/
│   │   └── bluebird.css    # Blue Bird modern CSS framework
│   └── js/
│       └── bluebird.js     # Client-side JavaScript
└── cli/
    └── main.go             # Smart CLI tool (rename, db, build, service)
```

---

## 📜 License

MIT License. Built with Go & Gin Framework.
