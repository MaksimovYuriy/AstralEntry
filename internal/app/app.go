package app

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/maksimovyuriy/astralentry/internal/config"
	"github.com/maksimovyuriy/astralentry/internal/controller/restapi"
	"github.com/maksimovyuriy/astralentry/pkg/httpserver"
	"github.com/maksimovyuriy/astralentry/pkg/logger"
	"github.com/maksimovyuriy/astralentry/pkg/postgres"
)

func Run() error {
	appCtx, appCancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer appCancel()

	appCfg, err := config.Load()
	if err != nil {
		return err
	}

	appLogger := logger.New(appCfg.App.Env)
	appLogger.Info("Logger started")

	appDatabase, err := postgres.New(appCfg.DB)
	if err != nil {
		return err
	}
	defer appDatabase.Close()

	repositories := initRepositories(appDatabase)
	useCases := initUseCases(repositories)
	appServer := initServer(appCfg.HTTP, useCases, appLogger)
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- appServer.Run()
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-appCtx.Done():
		appLogger.Info("App shutting down", "reason", appCtx.Err())
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		appCfg.HTTP.ShutdownTimeout,
	)
	defer shutdownCancel()

	if err := appServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	if err := <-serverErrors; err != nil {
		return err
	}

	appLogger.Info("App stopped")
	return nil
}

type repositories struct{}

type useCases struct{}

func initRepositories(_ *sql.DB) repositories {
	// TODO: initialize persistent repository implementations.
	return repositories{}
}

func initUseCases(_ repositories) useCases {
	// TODO: initialize application use cases.
	return useCases{}
}

func initServer(cfg config.HTTPConfig, _ useCases, logger *slog.Logger) *httpserver.Server {
	router := restapi.NewRouter(logger)

	return httpserver.New(
		net.JoinHostPort(cfg.Address, cfg.Port),
		router,
		httpserver.WithReadTimeout(cfg.ReadTimeout),
		httpserver.WithWriteTimeout(cfg.WriteTimeout),
		httpserver.WithReadHeaderTimeout(cfg.ReadHeaderTimeout),
		httpserver.WithIdleTimeout(cfg.IdleTimeout),
	)
}
