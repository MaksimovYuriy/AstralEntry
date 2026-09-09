package usecase

import (
	"context"
	"io"

	"github.com/maksimovyuriy/astralentry/internal/entity"
)

type (
	Auth interface {
		Register(ctx context.Context, adminToken, login, password string) (entity.User, error)
		Authenticate(ctx context.Context, login, password string) (string, error)
		Authorize(ctx context.Context, token string) (string, error)
		Logout(ctx context.Context, token string) error
	}

	Document interface {
		Create(
			ctx context.Context,
			document entity.Document,
			content entity.DocumentContent,
			grant []string,
			file io.Reader,
		) (entity.Document, error)
	}
)
