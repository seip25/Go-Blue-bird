package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Go Blue Bird CLI Tool")
	fmt.Println("Usage:")
	fmt.Println("  go run cli/main.go rename <new-module-name>   # Rename project module path")
	fmt.Println("  go run cli/main.go db tables                  # List database tables")
	fmt.Println("  go run cli/main.go db columns <table>        # List columns for a table")
	fmt.Println("  go run cli/main.go db query <table> [limit]   # Select rows from a table")
}

func renameModule(oldMod, newMod string) {
	fmt.Printf("Renaming module from '%s' to '%s'...\n", oldMod, newMod)

	count := 0
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (path == ".git" || path == "tmp" || path == "vendor" || path == "bin") {
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
