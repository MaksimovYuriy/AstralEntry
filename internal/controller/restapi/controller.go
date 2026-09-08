package restapi

import (
	"log/slog"

	"github.com/maksimovyuriy/astralentry/internal/usecase"
)

type Controller struct {
	auth   usecase.Auth
	logger *slog.Logger
}

func NewController(
	auth usecase.Auth,
	logger *slog.Logger,
) *Controller {
	return &Controller{
		auth:   auth,
		logger: logger,
	}
}
