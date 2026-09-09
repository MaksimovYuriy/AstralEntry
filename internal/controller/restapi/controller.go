package restapi

import (
	"log/slog"

	"github.com/maksimovyuriy/astralentry/internal/usecase"
)

type Controller struct {
	auth      usecase.Auth
	documents usecase.Document
	logger    *slog.Logger
}

func NewController(
	auth usecase.Auth,
	documents usecase.Document,
	logger *slog.Logger,
) *Controller {
	return &Controller{
		auth:      auth,
		documents: documents,
		logger:    logger,
	}
}
