package restapi

import (
	"log/slog"
	"net/http"
	"time"
)

func NewRouter(logger *slog.Logger) http.Handler {
	router := http.NewServeMux()
	apiRouter := http.NewServeMux()

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

		logger.Info(
			"HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(startedAt),
		)
	})
}
