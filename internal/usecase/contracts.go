package usecase

import (
	"context"

	"github.com/maksimovyuriy/astralentry/internal/entity"
)

type (
	Auth interface {
		Register(ctx context.Context, adminToken, login, password string) (entity.User, error)
		Authenticate(ctx context.Context, login, password string) (string, error)
		Authorize(ctx context.Context, token string) (string, error)
		Logout(ctx context.Context, token string) error
	}
)
