package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/maksimovyuriy/astralentry/internal/usecase"
)

type Authorizer interface {
	Authorize(ctx context.Context, token string) (string, error)
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type authorizationContextKey struct{}

type authorization struct {
	userID string
	token  string
}

func Authorize(auth Authorizer, writeError ErrorHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				writeError(w, r, usecase.ErrInvalidSession)
				return
			}

			userID, err := auth.Authorize(r.Context(), token)
			if err != nil {
				writeError(w, r, err)
				return
			}

			ctx := context.WithValue(r.Context(), authorizationContextKey{}, authorization{
				userID: userID,
				token:  token,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(ctx context.Context) string {
	return authorizationFromContext(ctx).userID
}

func Token(ctx context.Context) string {
	return authorizationFromContext(ctx).token
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	return parts[1], true
}

func authorizationFromContext(ctx context.Context) authorization {
	value, _ := ctx.Value(authorizationContextKey{}).(authorization)
	return value
}
