package document

import (
	"context"
	"database/sql"

	"github.com/maksimovyuriy/astralentry/internal/entity"
	"github.com/maksimovyuriy/astralentry/internal/repo"
)

type Repo struct {
	db *sql.DB
}

var _ repo.DocumentRepo = (*Repo)(nil)

func New(db *sql.DB) repo.DocumentRepo {
	return &Repo{db: db}
}

func (r *Repo) Create(
	ctx context.Context,
	document entity.Document,
	content entity.DocumentContent,
	users []entity.DocumentUser,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const documentQuery = `
		INSERT INTO documents (id, owner_id, name, mime, file, public, created)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if _, err := tx.ExecContext(
		ctx,
		documentQuery,
		document.ID,
		document.OwnerID,
		document.Name,
		document.Mime,
		document.File,
		document.Public,
		document.Created,
	); err != nil {
		return err
	}

	const contentQuery = `
		INSERT INTO document_contents (document_id, json, file_path)
		VALUES ($1, $2, $3)
	`
	var jsonContent any
	if len(content.JSON) != 0 {
		jsonContent = content.JSON
	}
	var filePath any
	if content.FilePath != "" {
		filePath = content.FilePath
	}
	if _, err := tx.ExecContext(ctx, contentQuery, document.ID, jsonContent, filePath); err != nil {
		return err
	}

	const userQuery = `
		INSERT INTO document_users (document_id, user_id)
		VALUES ($1, $2)
	`
	for _, user := range users {
		if _, err := tx.ExecContext(ctx, userQuery, document.ID, user.UserID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
