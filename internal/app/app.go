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
	"github.com/maksimovyuriy/astralentry/internal/repo"
	sessionrepo "github.com/maksimovyuriy/astralentry/internal/repo/session"
	userrepo "github.com/maksimovyuriy/astralentry/internal/repo/user"
	"github.com/maksimovyuriy/astralentry/internal/usecase"
	authusecase "github.com/maksimovyuriy/astralentry/internal/usecase/auth"
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
	useCases := initUseCases(repositories, appCfg.Auth)
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

type repositories struct {
	users    repo.UserRepo
	sessions repo.SessionRepo
}

type useCases struct {
	auth usecase.Auth
}

func initRepositories(db *sql.DB) repositories {
	return repositories{
		users:    userrepo.New(db),
		sessions: sessionrepo.New(db),
	}
}

func initUseCases(repositories repositories, cfg config.AuthConfig) useCases {
	return useCases{
		auth: authusecase.New(
			repositories.users,
			repositories.sessions,
			cfg.AdminToken,
			cfg.TokenTTL,
		),
	}
}

func initServer(cfg config.HTTPConfig, useCases useCases, logger *slog.Logger) *httpserver.Server {
	controller := restapi.NewController(useCases.auth, logger)
	router := restapi.NewRouter(controller, logger)

	return httpserver.New(
		net.JoinHostPort(cfg.Address, cfg.Port),
		router,
		httpserver.WithReadTimeout(cfg.ReadTimeout),
		httpserver.WithWriteTimeout(cfg.WriteTimeout),
		httpserver.WithReadHeaderTimeout(cfg.ReadHeaderTimeout),
		httpserver.WithIdleTimeout(cfg.IdleTimeout),
	)
}
