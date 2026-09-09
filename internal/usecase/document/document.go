package document

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/maksimovyuriy/astralentry/internal/entity"
	"github.com/maksimovyuriy/astralentry/internal/repo"
	"github.com/maksimovyuriy/astralentry/internal/usecase"
	"github.com/maksimovyuriy/astralentry/pkg/storage"
)

type UseCase struct {
	documents repo.DocumentRepo
	users     repo.UserRepo
	storage   storage.Storage
}

var _ usecase.Document = (*UseCase)(nil)

func New(
	documents repo.DocumentRepo,
	users repo.UserRepo,
	storage storage.Storage,
) usecase.Document {
	return &UseCase{documents: documents, users: users, storage: storage}
}

func (uc *UseCase) Create(
	ctx context.Context,
	document entity.Document,
	content entity.DocumentContent,
	grant []string,
	file io.Reader,
) (entity.Document, error) {
	documentID, err := generateUUID()
	if err != nil {
		return entity.Document{}, err
	}
	document.ID = documentID
	document.Created = time.Now().UTC()
	content.DocumentID = documentID

	documentUsers := make([]entity.DocumentUser, 0, len(grant))
	for _, login := range grant {
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

	if file != nil {
		content.FilePath = documentID
		if err := uc.storage.Save(content.FilePath, file); err != nil {
			return entity.Document{}, err
		}
	}

	if err := uc.documents.Create(ctx, document, content, documentUsers); err != nil {
		return entity.Document{}, uc.cleanupFile(content.FilePath, err)
	}

	return document, nil
}

func (uc *UseCase) cleanupFile(filePath string, cause error) error {
	if filePath == "" {
		return cause
	}

	if err := uc.storage.Delete(filePath); err != nil {
		return errors.Join(cause, fmt.Errorf("delete file after database error: %w", err))
	}

	return cause
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
