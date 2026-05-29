package config

import (
	"os"
	"strings"
	"testing"
)

func TestBuildDatabaseURL(t *testing.T) {
	dsn, err := BuildDatabaseURL("localhost", "5432", "loggerdb", "postgres", "postgres", "disable")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dsn, "localhost:5432") {
		t.Fatalf("unexpected host in dsn: %s", dsn)
	}
	if !strings.Contains(dsn, "/loggerdb") {
		t.Fatalf("unexpected db name in dsn: %s", dsn)
	}
	if !strings.Contains(dsn, "sslmode=disable") {
		t.Fatalf("missing sslmode: %s", dsn)
	}
}

func TestBuildDatabaseURL_EncodesSpecialPassword(t *testing.T) {
	dsn, err := BuildDatabaseURL("localhost", "5432", "loggerdb", "postgres", "p@ss:word", "disable")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dsn, "p%40ss") {
		t.Fatalf("expected encoded password in dsn: %s", dsn)
	}
}

func TestBuildDatabaseURL_Validation(t *testing.T) {
	_, err := BuildDatabaseURL("", "5432", "loggerdb", "postgres", "postgres", "disable")
	if err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestBuildTestDatabaseURL_Fallbacks(t *testing.T) {
	os.Unsetenv("TEST_DB_HOST")
	os.Unsetenv("TEST_DB_NAME")
	os.Setenv("DB_HOST", "dbhost")
	os.Setenv("DB_NAME", "loggerdb")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_NAME")
	}()

	dsn, err := BuildTestDatabaseURL()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dsn, "dbhost") {
		t.Fatalf("expected DB_HOST fallback: %s", dsn)
	}
	if !strings.Contains(dsn, "loggerdb_test") {
		t.Fatalf("expected default test db name: %s", dsn)
	}
}

func TestLoad_DefaultSQLite(t *testing.T) {
	clearDBEnv(t)
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("SQLITE_PATH", ":memory:")
	defer clearDBEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DBDriver != DriverSQLite {
		t.Fatalf("driver: got %q want %q", cfg.DBDriver, DriverSQLite)
	}
	if cfg.SQLitePath != ":memory:" {
		t.Fatalf("sqlite path: %s", cfg.SQLitePath)
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("expected empty DatabaseURL for sqlite, got %q", cfg.DatabaseURL)
	}
}

func TestLoad_PostgresBuildsURL(t *testing.T) {
	clearDBEnv(t)
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "loggerdb")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	defer clearDBEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DBDriver != DriverPostgres {
		t.Fatalf("driver: %q", cfg.DBDriver)
	}
	if !strings.Contains(cfg.DatabaseURL, "postgres://") {
		t.Fatalf("expected postgres url: %s", cfg.DatabaseURL)
	}
}

func TestLoad_UnsupportedDriver(t *testing.T) {
	clearDBEnv(t)
	os.Setenv("DB_DRIVER", "mysql")
	defer clearDBEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}

func TestLoad_EnableSwaggerUI_DefaultFalse(t *testing.T) {
	clearDBEnv(t)
	os.Unsetenv("ENABLE_SWAGGER_UI")
	defer clearDBEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnableSwaggerUI {
		t.Fatal("expected EnableSwaggerUI false by default")
	}
}

func TestLoad_EnableSwaggerUI_True(t *testing.T) {
	clearDBEnv(t)
	os.Setenv("ENABLE_SWAGGER_UI", "true")
	defer clearDBEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EnableSwaggerUI {
		t.Fatal("expected EnableSwaggerUI true")
	}
}

func TestLoad_DefaultAdminCredentials(t *testing.T) {
	clearDBEnv(t)
	os.Unsetenv("ADMIN_USERNAME")
	os.Unsetenv("ADMIN_PASSWORD")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AdminUsername != "admin" {
		t.Fatalf("username: %q", cfg.AdminUsername)
	}
	if cfg.AdminPassword != "12345678" {
		t.Fatalf("password: %q", cfg.AdminPassword)
	}
}

func clearDBEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"DB_DRIVER", "SQLITE_PATH", "DB_HOST", "DB_PORT", "DB_NAME",
		"DB_USER", "DB_PASSWORD", "DB_SSLMODE",
		"ADMIN_USERNAME", "ADMIN_PASSWORD", "ENABLE_SWAGGER_UI",
	} {
		os.Unsetenv(k)
	}
}
