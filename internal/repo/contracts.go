package repo

import (
	"context"

	"github.com/maksimovyuriy/astralentry/internal/entity"
)

type (
	UserRepo interface {
		Create(ctx context.Context, user entity.User) error
		FindByLogin(ctx context.Context, login string) (entity.User, error)
	}

	SessionRepo interface {
		Create(ctx context.Context, session entity.Session) error
		FindByTokenHash(ctx context.Context, tokenHash []byte) (entity.Session, error)
		DeleteByTokenHash(ctx context.Context, tokenHash []byte) error
	}
)
