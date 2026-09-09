package restapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/middleware"
)

func NewRouter(controller *Controller, logger *slog.Logger) http.Handler {
	router := http.NewServeMux()
	apiRouter := http.NewServeMux()

	apiRouter.HandleFunc("POST /register", controller.register)
	apiRouter.HandleFunc("POST /auth", controller.authenticate)
	authorize := middleware.Authorize(controller.auth, controller.writeError)
	apiRouter.Handle("DELETE /auth", authorize(http.HandlerFunc(controller.logout)))
	apiRouter.Handle("POST /docs", authorize(http.HandlerFunc(controller.createDocument)))
	apiRouter.Handle("GET /docs", authorize(http.HandlerFunc(controller.listDocuments)))
	apiRouter.Handle("HEAD /docs", authorize(http.HandlerFunc(controller.listDocuments)))

	router.HandleFunc("GET /healthz", health)
	router.Handle("/api/", http.StripPrefix("/api", apiRouter))

	return requestLogger(logger, router)
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodHead {
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()

		next.ServeHTTP(w, r)

		route := r.Pattern
		if route == "" {
			route = r.URL.Path
		}

		logger.Info(
			"HTTP request",
			"method", r.Method,
			"route", route,
			"duration", time.Since(startedAt),
		)
	})
}
