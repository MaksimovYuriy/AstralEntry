package document

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maksimovyuriy/astralentry/internal/entity"
	"github.com/maksimovyuriy/astralentry/internal/repo"
	"github.com/maksimovyuriy/astralentry/internal/usecase"
)

type UseCase struct {
	documents repo.DocumentRepo
	users     repo.UserRepo
}

var _ usecase.Document = (*UseCase)(nil)

func New(documents repo.DocumentRepo, users repo.UserRepo) usecase.Document {
	return &UseCase{documents: documents, users: users}
}

func (uc *UseCase) Create(
	ctx context.Context,
	document entity.Document,
	content entity.DocumentContent,
	grant []string,
) (entity.Document, error) {
	document.Name = strings.TrimSpace(document.Name)
	document.Mime = strings.TrimSpace(document.Mime)
	if document.OwnerID == "" || document.Name == "" || document.Mime == "" {
		return entity.Document{}, usecase.ErrInvalidDocument
	}
	if len(content.JSON) == 0 && content.FilePath == "" {
		return entity.Document{}, usecase.ErrInvalidDocument
	}
	if document.File != (content.FilePath != "") {
		return entity.Document{}, usecase.ErrInvalidDocument
	}
	if len(content.JSON) != 0 && !json.Valid(content.JSON) {
		return entity.Document{}, usecase.ErrInvalidDocument
	}

	documentID, err := generateUUID()
	if err != nil {
		return entity.Document{}, err
	}
	document.ID = documentID
	document.Created = time.Now().UTC()
	content.DocumentID = documentID

	documentUsers := make([]entity.DocumentUser, 0, len(grant))
	seen := make(map[string]struct{}, len(grant))
	for _, login := range grant {
		login = strings.TrimSpace(login)
		if login == "" {
			return entity.Document{}, usecase.ErrGrantUserNotFound
		}
		if _, exists := seen[login]; exists {
			continue
		}
		seen[login] = struct{}{}

		user, err := uc.users.FindByLogin(ctx, login)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return entity.Document{}, usecase.ErrGrantUserNotFound
			}
			return entity.Document{}, err
		}
		if user.ID == document.OwnerID {
			continue
		}

		documentUsers = append(documentUsers, entity.DocumentUser{
			DocumentID: documentID,
			UserID:     user.ID,
		})
	}

	if err := uc.documents.Create(ctx, document, content, documentUsers); err != nil {
		return entity.Document{}, err
	}

	return document, nil
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
		id[0:4], id[4:6], id[6:8], id[8:10], id[10:16],
	), nil
}
