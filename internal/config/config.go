package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	DriverSQLite   = "sqlite"
	DriverPostgres = "postgres"
)

type Config struct {
	AppEnv                string
	AppPort               string
	DBDriver              string
	SQLitePath            string
	DatabaseURL           string
	DBHost                string
	DBPort                string
	DBName                string
	DBUser                string
	DBPassword            string
	DBSSLMode             string
	DBAutoMigrate         bool
	ServiceName           string
	EnableConsoleLogging  bool
	EnableDatabaseLogging bool
	LogLevel              string
	AdminUsername         string
	AdminPassword         string
	AdminAPIKey           string
	JWTSecret             string
	JWTExpiryHours        int
	CORSOrigins           string
	EnableSwaggerUI       bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:                getEnv("APP_ENV", "local"),
		AppPort:               getEnv("APP_PORT", "8080"),
		DBDriver:              strings.ToLower(getEnv("DB_DRIVER", DriverSQLite)),
		SQLitePath:            getEnv("SQLITE_PATH", "./data/logger.db"),
		DBHost:                getEnv("DB_HOST", "localhost"),
		DBPort:                getEnv("DB_PORT", "5432"),
		DBName:                getEnv("DB_NAME", "loggerdb"),
		DBUser:                getEnv("DB_USER", "postgres"),
		DBPassword:            getEnv("DB_PASSWORD", "postgres"),
		DBSSLMode:             getEnv("DB_SSLMODE", "disable"),
		DBAutoMigrate:         getEnvBool("DB_AUTO_MIGRATE", true),
		ServiceName:           getEnv("SERVICE_NAME", "application-logging-audit-module"),
		EnableConsoleLogging:  getEnvBool("ENABLE_CONSOLE_LOGGING", true),
		EnableDatabaseLogging: getEnvBool("ENABLE_DATABASE_LOGGING", true),
		LogLevel:              getEnv("LOG_LEVEL", "INFO"),
		AdminUsername:         getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:         getEnv("ADMIN_PASSWORD", "12345678"),
		AdminAPIKey:           getEnv("ADMIN_API_KEY", "super-secret-admin-key"),
		JWTSecret:             getEnv("JWT_SECRET", "change-this-secret"),
		JWTExpiryHours:        getEnvInt("JWT_EXPIRY_HOURS", 24),
		CORSOrigins:           getEnv("CORS_ORIGINS", "http://localhost:5173"),
		EnableSwaggerUI:       getEnvBool("ENABLE_SWAGGER_UI", false),
	}

	switch cfg.DBDriver {
	case DriverSQLite:
		if cfg.SQLitePath == "" {
			return nil, fmt.Errorf("SQLITE_PATH is required when DB_DRIVER=sqlite")
		}
	case DriverPostgres:
		var err error
		cfg.DatabaseURL, err = BuildDatabaseURL(
			cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword, cfg.DBSSLMode,
		)
		if err != nil {
			return nil, fmt.Errorf("build database url: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q (use %q or %q)", cfg.DBDriver, DriverSQLite, DriverPostgres)
	}

	return cfg, nil
}

func BuildDatabaseURL(host, port, name, user, password, sslMode string) (string, error) {
	if host == "" {
		return "", fmt.Errorf("DB_HOST is required")
	}
	if port == "" {
		return "", fmt.Errorf("DB_PORT is required")
	}
	if name == "" {
		return "", fmt.Errorf("DB_NAME is required")
	}
	if user == "" {
		return "", fmt.Errorf("DB_USER is required")
	}
	if sslMode == "" {
		sslMode = "disable"
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   "/" + name,
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func BuildTestDatabaseURL() (string, error) {
	host := getEnv("TEST_DB_HOST", getEnv("DB_HOST", "localhost"))
	port := getEnv("TEST_DB_PORT", getEnv("DB_PORT", "5432"))
	name := getEnv("TEST_DB_NAME", "loggerdb_test")
	user := getEnv("TEST_DB_USER", getEnv("DB_USER", "postgres"))
	password := getEnv("TEST_DB_PASSWORD", getEnv("DB_PASSWORD", "postgres"))
	sslMode := getEnv("TEST_DB_SSLMODE", getEnv("DB_SSLMODE", "disable"))

	return BuildDatabaseURL(host, port, name, user, password, sslMode)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
