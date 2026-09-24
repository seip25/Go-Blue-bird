package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/seip25/Go-Blue-bird/config"
	"github.com/seip25/Go-Blue-bird/core"
)

const defaultModule = "github.com/seip25/Go-Blue-bird"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "rename":
		if len(os.Args) < 3 {
			log.Fatal("Error: New module name is required. Example: go run cli/main.go rename github.com/user/my-repo")
		}
		newModule := os.Args[2]
		renameModule(defaultModule, newModule)
	case "db":
		if len(os.Args) < 3 {
			log.Fatal("Error: DB sub-command required (tables, columns <table>, query <table>)")
		}
		handleDBCommand(os.Args[2:])
	case "build":
		handleBuildCommand(os.Args[2:])
	case "service":
		handleServiceCommand(os.Args[2:])
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("⚡ Go Blue Bird CLI Tool")
	fmt.Println("Usage:")
	fmt.Println("  go run cli/main.go rename <new-module-name>          # Rename project module path")
	fmt.Println("  go run cli/main.go db tables                         # List database tables")
	fmt.Println("  go run cli/main.go db columns <table>               # List columns for a table")
	fmt.Println("  go run cli/main.go db query <table> [limit]          # Select rows from a table")
	fmt.Println("  go run cli/main.go build [os] [arch] [--dir=<path>]  # Compile binary into builds/ and generate VPS deploy configs")
	fmt.Println("  go run cli/main.go service [--install]               # Generate or install systemd service on Linux VPS")
}

func renameModule(oldMod, newMod string) {
	fmt.Printf("Renaming module from '%s' to '%s'...\n", oldMod, newMod)

	count := 0
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (path == ".git" || path == "tmp" || path == "vendor" || path == "bin" || path == "builds") {
			return filepath.SkipDir
		}

		if !d.IsDir() && (strings.HasSuffix(path, ".go") || path == "go.mod" || path == "README.md") {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			if strings.Contains(string(content), oldMod) {
				updated := strings.ReplaceAll(string(content), oldMod, newMod)
				writeErr := os.WriteFile(path, []byte(updated), 0644)
				if writeErr != nil {
					return writeErr
				}
				fmt.Printf(" Updated: %s\n", path)
				count++
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Rename failed: %v\n", err)
	}

	fmt.Printf("Successfully updated %d files to module '%s'.\n", count, newMod)
}

func handleBuildCommand(args []string) {
	cfg := config.Load()
	appName := strings.ToLower(cfg.AppName)
	if appName == "" {
		appName = "goapp"
	}
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH
	vpsDir := "/var/www/" + appName

	for _, arg := range args {
		if strings.HasPrefix(arg, "--dir=") {
			vpsDir = strings.TrimPrefix(arg, "--dir=")
		} else if !strings.HasPrefix(arg, "-") {
			if arg == "linux" || arg == "windows" || arg == "darwin" {
				targetOS = arg
			} else if arg == "amd64" || arg == "arm64" || arg == "386" {
				targetArch = arg
			}
		}
	}

	buildDir := "builds"
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		log.Fatalf("Failed to create %s directory: %v", buildDir, err)
	}

	binName := appName
	if targetOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(buildDir, binName)

	fmt.Printf("📦 Compiling Go Blue Bird for %s/%s...\n", targetOS, targetArch)
	fmt.Printf(" - Binary target: %s\n", binPath)

	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", binPath, ".")
	cmd.Env = append(os.Environ(), "GOOS="+targetOS, "GOARCH="+targetArch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Compilation error:\n%s\n", string(output))
		fmt.Println("\n💡 Tip: When cross-compiling SQLite with CGO, compiling natively on target OS or using a cross-compiler is recommended.")
		return
	}

	fi, err := os.Stat(binPath)
	if err == nil {
		fmt.Printf("✅ Binary compiled successfully! Size: %.2f MB\n", float64(fi.Size())/(1024*1024))
	}

	// 1. Generate systemd service configuration file
	serviceFile := filepath.Join(buildDir, "systemd_service.txt")
	serviceContent := fmt.Sprintf(`# Systemd Service for %s
# Target path on VPS: /etc/systemd/system/%s.service
#
# Deployment commands on Ubuntu/Debian VPS:
#   sudo cp systemd_service.txt /etc/systemd/system/%s.service
#   sudo systemctl daemon-reload
#   sudo systemctl enable --now %s
#   sudo systemctl status %s
#   sudo journalctl -u %s -f

[Unit]
Description=%s Go Web Service
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=%s
ExecStart=%s/%s
Restart=always
RestartSec=5s
EnvironmentFile=%s/.env
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`, cfg.AppName, appName, appName, appName, appName, appName, cfg.AppName, vpsDir, vpsDir, binName, vpsDir)

	_ = os.WriteFile(serviceFile, []byte(serviceContent), 0644)
	fmt.Printf("📄 Generated Systemd config: %s\n", serviceFile)

	// 2. Generate Nginx reverse proxy configuration file
	nginxFile := filepath.Join(buildDir, "nginx_reverse_proxy.txt")
	nginxContent := fmt.Sprintf(`# Nginx Reverse Proxy for %s
# Target path on VPS: /etc/nginx/sites-available/%s
#
# Activation commands on VPS:
#   sudo cp nginx_reverse_proxy.txt /etc/nginx/sites-available/%s
#   sudo ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/
#   sudo nginx -t && sudo systemctl reload nginx
#   sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    client_max_body_size 50M;

    location / {
        proxy_pass http://127.0.0.1:%s;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_buffering off;
        proxy_read_timeout 300s;
    }
}
`, cfg.AppName, appName, appName, appName, port)

	_ = os.WriteFile(nginxFile, []byte(nginxContent), 0644)
	fmt.Printf("📄 Generated Nginx config:   %s\n", nginxFile)

	// 3. Generate step-by-step deploy guide
	guideFile := filepath.Join(buildDir, "deploy_guide.txt")
	guideContent := fmt.Sprintf(`=======================================================
🚀 GUIA RAPIDA DE DESPLIEGUE EN VPS - %s
=======================================================

1. Preparar directorio en tu VPS:
   sudo mkdir -p %s
   sudo chown -R www-data:www-data %s

2. Subir binario y configuracion:
   scp %s user@vps:%s/
   scp .env user@vps:%s/
   ssh user@vps "chmod +x %s/%s"

3. Configurar servicio Systemd:
   scp %s user@vps:/tmp/%s.service
   ssh user@vps "sudo mv /tmp/%s.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl enable --now %s"

4. Configurar Nginx (Reverse Proxy):
   scp %s user@vps:/tmp/%s
   ssh user@vps "sudo mv /tmp/%s /etc/nginx/sites-available/ && sudo ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/ && sudo nginx -t && sudo systemctl reload nginx"

5. Certificado SSL gratuito (Opcional):
   ssh user@vps "sudo certbot --nginx -d yourdomain.com"

=======================================================
Listo! Tu aplicacion estara corriendo en produccion.
=======================================================
`, cfg.AppName, vpsDir, vpsDir, binPath, vpsDir, vpsDir, vpsDir, binName, serviceFile, appName, appName, appName, nginxFile, appName, appName, appName)

	_ = os.WriteFile(guideFile, []byte(guideContent), 0644)
	fmt.Printf("📄 Generated Deploy guide:   %s\n", guideFile)
	fmt.Println("\n✨ Build completed! Everything you need for deployment is inside builds/ folder.")
}

func handleServiceCommand(args []string) {
	cfg := config.Load()
	appName := strings.ToLower(cfg.AppName)
	if appName == "" {
		appName = "goapp"
	}

	install := false
	for _, a := range args {
		if a == "--install" {
			install = true
		}
	}

	if install {
		if runtime.GOOS != "linux" {
			log.Fatalf("Error: --install only runs natively on a Linux VPS.")
		}
		if os.Geteuid() != 0 {
			log.Fatalf("Error: Please run with sudo: sudo go run cli/main.go service --install")
		}

		vpsDir := "/var/www/" + appName
		servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", appName)
		serviceContent := fmt.Sprintf(`[Unit]
Description=%s Go Web Service
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=%s
ExecStart=%s/%s
Restart=always
RestartSec=5s
EnvironmentFile=%s/.env
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`, cfg.AppName, vpsDir, vpsDir, appName, vpsDir)

		if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
			log.Fatalf("Failed to write %s: %v", servicePath, err)
		}

		_ = exec.Command("systemctl", "daemon-reload").Run()
		_ = exec.Command("systemctl", "enable", "--now", appName).Run()

		fmt.Printf("✅ Service %s installed and started successfully!\n", appName)
		fmt.Printf("Check status with: systemctl status %s\n", appName)
		return
	}

	fmt.Println("To compile and generate deployment files, run:")
	fmt.Println("  go run cli/main.go build")
	fmt.Println("To install systemd service automatically on a Linux VPS:")
	fmt.Println("  sudo go run cli/main.go service --install")
}

func handleDBCommand(args []string) {
	cfg := config.Load()
	db, err := core.ConnectDB(cfg.DB)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	subCmd := args[0]
	driver := strings.ToLower(cfg.DB.Driver)

	switch subCmd {
	case "tables":
		listTables(db, driver)
	case "columns":
		if len(args) < 2 {
			log.Fatal("Error: Table name required for columns command")
		}
		listColumns(db, driver, args[1])
	case "query":
		if len(args) < 2 {
			log.Fatal("Error: Table name required for query command")
		}
		limit := "10"
		if len(args) >= 3 {
			limit = args[2]
		}
		queryTable(db, args[1], limit)
	default:
		log.Fatalf("Unknown db subcommand: %s", subCmd)
	}
}

func listTables(db *sql.DB, driver string) {
	var query string
	switch driver {
	case "mysql":
		query = "SHOW TABLES;"
	case "postgres", "postgresql":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema='public';"
	default:
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';"
	}

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Printf("Tables in database (%s):\n", driver)
	fmt.Println("----------------------------------------")
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			fmt.Printf(" - %s\n", name)
		}
	}
}

func listColumns(db *sql.DB, driver string, tableName string) {
	var query string
	switch driver {
	case "mysql":
		query = fmt.Sprintf("SHOW COLUMNS FROM `%s`;", tableName)
	case "postgres", "postgresql":
		query = fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name='%s';", tableName)
	default:
		query = fmt.Sprintf("PRAGMA table_info(`%s`);", tableName)
	}

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	fmt.Printf("Columns for table '%s':\n", tableName)
	fmt.Println("----------------------------------------")

	cols, err := rows.Columns()
	if err != nil {
		log.Fatalf("Failed to read columns: %v", err)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err == nil {
			rowStr := ""
			for i, val := range values {
				if val != nil {
					rowStr += fmt.Sprintf("%s: %v | ", cols[i], string(fmt.Sprintf("%v", val)))
				}
			}
			fmt.Println(rowStr)
		}
	}
}

func queryTable(db *sql.DB, tableName string, limit string) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %s;", tableName, limit)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Fatalf("Failed to get columns: %v", err)
	}

	fmt.Printf("Results for query: %s\n", query)
	fmt.Println(strings.Join(cols, "\t| "))
	fmt.Println(strings.Repeat("-", 60))

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err == nil {
			var rowValues []string
			for _, val := range values {
				if val == nil {
					rowValues = append(rowValues, "NULL")
				} else {
					rowValues = append(rowValues, fmt.Sprintf("%v", val))
				}
			}
			fmt.Println(strings.Join(rowValues, "\t| "))
		}
	}
}
