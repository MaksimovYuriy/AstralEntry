package document

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/maksimovyuriy/astralentry/internal/entity"
	"github.com/maksimovyuriy/astralentry/internal/repo"
)

type Repo struct {
	db *sql.DB
}

const listDocumentsQuery = `
	SELECT
		d.id,
		d.owner_id,
		d.name,
		d.mime,
		d.file,
		d.public,
		d.created,
		COALESCE((
			SELECT jsonb_agg(u.login ORDER BY u.login)
			FROM document_users du
			JOIN users u ON u.id = du.user_id
			WHERE du.document_id = d.id
		), '[]'::jsonb)
	FROM documents d
	WHERE d.owner_id = $1
		AND (
			d.owner_id = $2
			OR d.public = TRUE
			OR EXISTS (
				SELECT 1
				FROM document_users du_access
				WHERE du_access.document_id = d.id
					AND du_access.user_id = $2
			)
		)
`

const getDocumentQuery = `
	SELECT
		d.id,
		d.owner_id,
		d.name,
		d.mime,
		d.file,
		d.public,
		d.created,
		dc.json,
		dc.file_path,
		(
			d.owner_id = $2
			OR d.public = TRUE
			OR EXISTS (
				SELECT 1
				FROM document_users du
				WHERE du.document_id = d.id
					AND du.user_id = $2
			)
		) AS allowed
	FROM documents d
	JOIN document_contents dc ON dc.document_id = d.id
	WHERE d.id = $1
`

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
		ON CONFLICT (document_id, user_id) DO NOTHING
	`
	for _, user := range users {
		if _, err := tx.ExecContext(ctx, userQuery, document.ID, user.UserID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repo) List(
	ctx context.Context,
	requesterID string,
	ownerID string,
	key string,
	value string,
	limit int,
	offset int,
) ([]entity.Document, error) {
	query, arguments, err := buildListQuery(ownerID, requesterID, key, value, limit, offset)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, arguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDocuments(rows)
}

func (r *Repo) Get(
	ctx context.Context,
	requesterID string,
	documentID string,
) (entity.Document, entity.DocumentContent, error) {
	var document entity.Document
	var content entity.DocumentContent
	var filePath sql.NullString
	var allowed bool
	err := r.db.QueryRowContext(ctx, getDocumentQuery, documentID, requesterID).Scan(
		&document.ID,
		&document.OwnerID,
		&document.Name,
		&document.Mime,
		&document.File,
		&document.Public,
		&document.Created,
		&content.JSON,
		&filePath,
		&allowed,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Document{}, entity.DocumentContent{}, repo.ErrNotFound
	}
	if err != nil {
		return entity.Document{}, entity.DocumentContent{}, err
	}
	if !allowed {
		return entity.Document{}, entity.DocumentContent{}, repo.ErrForbidden
	}

	content.DocumentID = document.ID
	content.FilePath = filePath.String
	return document, content, nil
}

func buildListQuery(
	ownerID string,
	requesterID string,
	key string,
	value string,
	limit int,
	offset int,
) (string, []any, error) {
	column, filterValue, err := documentFilter(key, value)
	if err != nil {
		return "", nil, err
	}

	query := listDocumentsQuery
	arguments := []any{ownerID, requesterID}
	if column != "" {
		arguments = append(arguments, filterValue)
		query += fmt.Sprintf("\tAND %s = $%d\n", column, len(arguments))
	}

	arguments = append(arguments, limit, offset)
	query += fmt.Sprintf(`
	ORDER BY d.name ASC, d.created DESC, d.id ASC
	LIMIT $%d OFFSET $%d
`, len(arguments)-1, len(arguments))

	return query, arguments, nil
}

func scanDocuments(rows *sql.Rows) ([]entity.Document, error) {
	documents := make([]entity.Document, 0)
	for rows.Next() {
		var document entity.Document
		var grantJSON []byte
		if err := rows.Scan(
			&document.ID,
			&document.OwnerID,
			&document.Name,
			&document.Mime,
			&document.File,
			&document.Public,
			&document.Created,
			&grantJSON,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(grantJSON, &document.Grant); err != nil {
			return nil, err
		}

		documents = append(documents, document)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}

func documentFilter(key, value string) (string, any, error) {
	switch key {
	case "":
		return "", nil, nil
	case "id":
		parsed, err := uuid.Parse(value)
		if err != nil {
			return "", nil, repo.ErrInvalidFilter
		}
		return "d.id", parsed, nil
	case "name":
		return "d.name", value, nil
	case "mime":
		return "d.mime", value, nil
	case "file":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return "", nil, repo.ErrInvalidFilter
		}
		return "d.file", parsed, nil
	case "public":
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return "", nil, repo.ErrInvalidFilter
		}
		return "d.public", parsed, nil
	case "created":
		parsed, err := time.Parse("2006-01-02 15:04:05", value)
		if err != nil {
			return "", nil, repo.ErrInvalidFilter
		}
		return "date_trunc('second', d.created)", parsed, nil
	default:
		return "", nil, repo.ErrInvalidFilter
	}
}
