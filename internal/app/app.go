package app

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maksimovyuriy/astralentry/internal/cache"
	redisCache "github.com/maksimovyuriy/astralentry/internal/cache/redis"
	"github.com/maksimovyuriy/astralentry/internal/config"
	"github.com/maksimovyuriy/astralentry/internal/controller/restapi"
	"github.com/maksimovyuriy/astralentry/internal/repo"
	documentrepo "github.com/maksimovyuriy/astralentry/internal/repo/document"
	sessionrepo "github.com/maksimovyuriy/astralentry/internal/repo/session"
	userrepo "github.com/maksimovyuriy/astralentry/internal/repo/user"
	"github.com/maksimovyuriy/astralentry/internal/usecase"
	authusecase "github.com/maksimovyuriy/astralentry/internal/usecase/auth"
	documentusecase "github.com/maksimovyuriy/astralentry/internal/usecase/document"
	"github.com/maksimovyuriy/astralentry/pkg/httpserver"
	"github.com/maksimovyuriy/astralentry/pkg/logger"
	"github.com/maksimovyuriy/astralentry/pkg/postgres"
	redisClient "github.com/maksimovyuriy/astralentry/pkg/redis"
	"github.com/maksimovyuriy/astralentry/pkg/storage"
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
	appLogger.Info("Database connected")

	fileStorage, err := storage.NewLocal(appCfg.Storage.Path)
	if err != nil {
		return err
	}
	appLogger.Info("File storage initialized")

	redisConnection := redisClient.New(appCfg.Redis)
	defer redisConnection.Close()
	redisCtx, redisCancel := context.WithTimeout(appCtx, 2*time.Second)
	if err := redisConnection.Ping(redisCtx).Err(); err != nil {
		appLogger.Warn("Redis cache unavailable", "error", err)
	} else {
		appLogger.Info("Redis cache connected")
	}
	redisCancel()
	documentCache := redisCache.NewDocument(redisConnection, appCfg.Redis.TTL)

	repositories := initRepositories(appDatabase)
	useCases := initUseCases(repositories, fileStorage, documentCache, appCfg, appLogger)
	appServer := initServer(appCfg.HTTP, useCases, appLogger)
	serverErrors := make(chan error, 1)
	appLogger.Info("Server started")

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
	users     repo.UserRepo
	sessions  repo.SessionRepo
	documents repo.DocumentRepo
}

type useCases struct {
	auth      usecase.Auth
	documents usecase.Document
}

func initRepositories(db *sql.DB) repositories {
	return repositories{
		users:     userrepo.New(db),
		sessions:  sessionrepo.New(db),
		documents: documentrepo.New(db),
	}
}

func initUseCases(
	repositories repositories,
	fileStorage storage.Storage,
	documentCache cache.Document,
	cfg *config.Config,
	logger *slog.Logger,
) useCases {
	documents := documentusecase.New(
		repositories.documents,
		repositories.users,
		fileStorage,
	)

	return useCases{
		auth: authusecase.New(
			repositories.users,
			repositories.sessions,
			cfg.Auth.AdminToken,
			cfg.Auth.TokenTTL,
		),
		documents: documentusecase.NewCached(
			documents,
			documentCache,
			logger,
			cfg.Redis.MaxFileSize,
		),
	}
}

func initServer(cfg config.HTTPConfig, useCases useCases, logger *slog.Logger) *httpserver.Server {
	controller := restapi.NewController(useCases.auth, useCases.documents, logger)
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
