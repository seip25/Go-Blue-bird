package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppEnv  string
	Port    string
	DB      DatabaseConfig
	APIKey  string
}

type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	Charset         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

var globalConfig *Config

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		AppName: Get("APP_NAME", "GoApp"),
		AppEnv:  strings.ToLower(Get("APP_ENV", "development")),
		Port:    Get("PORT", "8080"),
		APIKey:  Get("API_KEY", ""),
		DB: DatabaseConfig{
			Driver:          strings.ToLower(Get("DB_DRIVER", "sqlite")),
			Host:            Get("DB_HOST", "127.0.0.1"),
			Port:            Get("DB_PORT", "3306"),
			User:            Get("DB_USER", "root"),
			Password:        Get("DB_PASSWORD", ""),
			Name:            Get("DB_NAME", "app_db.sqlite"),
			Charset:         Get("DB_CHARSET", "utf8mb4"),
			MaxOpenConns:    GetInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    GetInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: GetInt("DB_CONN_MAX_LIFETIME_MIN", 5),
		},
	}

	globalConfig = cfg
	return cfg
}

func GetGlobal() *Config {
	if globalConfig == nil {
		return Load()
	}
	return globalConfig
}

func Get(key string, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func GetInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func GetBool(key string, defaultValue bool) bool {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func IsDev() bool {
	env := strings.ToLower(Get("APP_ENV", "development"))
	return env == "development" || env == "dev"
}

func IsProd() bool {
	env := strings.ToLower(Get("APP_ENV", "development"))
	return env == "production" || env == "prod"
}
