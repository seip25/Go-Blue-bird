package core

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/seip25/Go-Blue-bird/config"
)

var globalDB *sql.DB

func ConnectDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	driver := strings.ToLower(cfg.Driver)
	dsn := buildDSN(cfg)

	maxRetries := 1
	if driver == "mysql" || driver == "postgres" || driver == "postgresql" {
		maxRetries = 5
	}

	var db *sql.DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = sql.Open(getDriverName(driver), dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Printf("Successfully connected to %s database", driver)
				break
			}
		}

		log.Printf("Failed database connection attempt %d/%d (%s): %v", attempt, maxRetries, driver, err)
		if attempt < maxRetries {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("unable to connect to database after %d attempts: %w", maxRetries, err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	globalDB = db
	return db, nil
}

func getDriverName(driver string) string {
	switch driver {
	case "sqlite", "sqlite3":
		return "sqlite3"
	case "postgres", "postgresql":
		return "postgres"
	case "mysql":
		return "mysql"
	default:
		return driver
	}
}

func buildDSN(cfg config.DatabaseConfig) string {
	driver := strings.ToLower(cfg.Driver)
	switch driver {
	case "sqlite", "sqlite3":
		dsn := cfg.Name
		if !strings.Contains(dsn, "?") {
			dsn += "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL"
		}
		return dsn
	case "mysql":
		charset := cfg.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, charset)
	case "postgres", "postgresql":
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable client_encoding=UTF8",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	default:
		return cfg.Name
	}
}

func GetDB() *sql.DB {
	return globalDB
}

func PingDB() error {
	if globalDB == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return globalDB.Ping()
}

func ExecuteQuery(query string, args ...interface{}) (sql.Result, error) {
	if globalDB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	return globalDB.Exec(query, args...)
}

func QueryRow(query string, args ...interface{}) *sql.Row {
	if globalDB == nil {
		return nil
	}
	return globalDB.QueryRow(query, args...)
}

func QueryRows(query string, args ...interface{}) (*sql.Rows, error) {
	if globalDB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	return globalDB.Query(query, args...)
}

func CloseDB() error {
	if globalDB != nil {
		return globalDB.Close()
	}
	return nil
}
