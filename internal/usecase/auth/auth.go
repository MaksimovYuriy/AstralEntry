package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"github.com/maksimovyuriy/astralentry/internal/entity"
	"github.com/maksimovyuriy/astralentry/internal/repo"
	"github.com/maksimovyuriy/astralentry/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

const tokenSize = 32

var loginPattern = regexp.MustCompile(`^[A-Za-z0-9]{8,}$`)

type UseCase struct {
	users      repo.UserRepo
	sessions   repo.SessionRepo
	adminToken string
	tokenTTL   time.Duration
}

var _ usecase.Auth = (*UseCase)(nil)

func New(
	users repo.UserRepo,
	sessions repo.SessionRepo,
	adminToken string,
	tokenTTL time.Duration,
) usecase.Auth {
	return &UseCase{
		users:      users,
		sessions:   sessions,
		adminToken: adminToken,
		tokenTTL:   tokenTTL,
	}
}

func (uc *UseCase) Register(
	ctx context.Context,
	adminToken string,
	login string,
	password string,
) (entity.User, error) {
	if !tokensEqual(adminToken, uc.adminToken) {
		return entity.User{}, errors.New("invalid admin token")
	}
	if err := validateLogin(login); err != nil {
		return entity.User{}, err
	}
	if err := validatePassword(password); err != nil {
		return entity.User{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, err
	}

	userID, err := generateUUID()
	if err != nil {
		return entity.User{}, err
	}

	user := entity.User{
		ID:           userID,
		Login:        login,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now().UTC(),
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (uc *UseCase) Authenticate(
	ctx context.Context,
	login string,
	password string,
) (string, error) {
	user, err := uc.users.FindByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", err
	}

	token, err := generateToken()
	if err != nil {
		return "", err
	}

	createdAt := time.Now().UTC()
	session := entity.Session{
		TokenHash: tokenHash(token),
		UserID:    user.ID,
		CreatedAt: createdAt,
		ExpiresAt: createdAt.Add(uc.tokenTTL),
	}

	if err := uc.sessions.Create(ctx, session); err != nil {
		return "", err
	}

	return token, nil
}

func (uc *UseCase) Authorize(ctx context.Context, token string) (string, error) {
	session, err := uc.sessions.FindByTokenHash(ctx, tokenHash(token))
	if err != nil {
		return "", err
	}

	return session.UserID, nil
}

func (uc *UseCase) Logout(ctx context.Context, token string) error {
	return uc.sessions.DeleteByTokenHash(ctx, tokenHash(token))
}

func validateLogin(login string) error {
	if !loginPattern.MatchString(login) {
		return errors.New("login must contain at least 8 latin letters or digits")
	}

	return nil
}

func validatePassword(password string) error {
	if len([]rune(password)) < 8 {
		return errors.New("password must contain at least 8 characters")
	}

	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, character := range password {
		switch {
		case unicode.IsLower(character):
			hasLower = true
		case unicode.IsUpper(character):
			hasUpper = true
		case unicode.IsDigit(character):
			hasDigit = true
		case !unicode.IsLetter(character) && !unicode.IsDigit(character):
			hasSymbol = true
		}
	}

	if !hasLower || !hasUpper || !hasDigit || !hasSymbol {
		return errors.New("password must contain lower-case, upper-case, digit and symbol characters")
	}

	return nil
}

func tokensEqual(left, right string) bool {
	leftHash := sha256.Sum256([]byte(left))
	rightHash := sha256.Sum256([]byte(right))
	return subtle.ConstantTimeCompare(leftHash[:], rightHash[:]) == 1
}

func generateToken() (string, error) {
	buffer := make([]byte, tokenSize)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func tokenHash(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func generateUUID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}

	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		id[0:4],
		id[4:6],
		id[6:8],
		id[8:10],
		id[10:16],
	), nil
}
