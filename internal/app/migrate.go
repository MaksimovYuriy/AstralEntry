package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/maksimovyuriy/astralentry/internal/config"
	"github.com/maksimovyuriy/astralentry/pkg/postgres"
	"github.com/pressly/goose/v3"
)

const migrationsDirectory = "migrations"

func Migrate() error {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := postgres.New(cfg.DB)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, migrationsDirectory); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
