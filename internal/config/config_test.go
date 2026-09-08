package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromDotEnv(t *testing.T) {
	workingDirectory := changeWorkingDirectory(t, t.TempDir())
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	t.Setenv("APP_ENV", "before_dotenv")
	t.Setenv("HTTP_PORT", "39091")
	t.Setenv("DB_NAME", "before_dotenv_database")

	dotEnv := []byte("APP_ENV=dotenv\nHTTP_PORT=19091\nDB_NAME=dotenv_database\n")
	if err := os.WriteFile(".env", dotEnv, 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.App.Env != "dotenv" {
		t.Errorf("App.Env = %q, want %q", cfg.App.Env, "dotenv")
	}
	if cfg.HTTP.Port != "19091" {
		t.Errorf("HTTP.Port = %q, want %q", cfg.HTTP.Port, "19091")
	}
	if cfg.DB.Name != "dotenv_database" {
		t.Errorf("DB.Name = %q, want %q", cfg.DB.Name, "dotenv_database")
	}
}

func TestLoadFromEnvironmentWhenDotEnvDoesNotExist(t *testing.T) {
	workingDirectory := changeWorkingDirectory(t, t.TempDir())
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	t.Setenv("APP_ENV", "environment")
	t.Setenv("HTTP_PORT", "29091")
	t.Setenv("DB_NAME", "environment_database")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.App.Env != "environment" {
		t.Errorf("App.Env = %q, want %q", cfg.App.Env, "environment")
	}
	if cfg.HTTP.Port != "29091" {
		t.Errorf("HTTP.Port = %q, want %q", cfg.HTTP.Port, "29091")
	}
	if cfg.DB.Name != "environment_database" {
		t.Errorf("DB.Name = %q, want %q", cfg.DB.Name, "environment_database")
	}
}

func changeWorkingDirectory(t *testing.T, directory string) string {
	t.Helper()

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(filepath.Clean(directory)); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	return workingDirectory
}
