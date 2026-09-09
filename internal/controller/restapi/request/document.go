package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidDocument = errors.New("invalid document")

const (
	defaultDocumentsLimit = 20
	maxDocumentsLimit     = 100
)

type CreateDocument struct {
	Meta CreateDocumentMeta
	JSON json.RawMessage
	File *multipart.FileHeader
}

type CreateDocumentMeta struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}

type ListDocuments struct {
	Login  string
	Key    string
	Value  string
	Limit  string
	Offset string
}

type GetDocument struct {
	ID string
}

type DeleteDocument struct {
	ID string
}

func (r *CreateDocument) Normalize() {
	r.Meta.Name = strings.TrimSpace(r.Meta.Name)
	r.Meta.Mime = strings.TrimSpace(r.Meta.Mime)

	grant := make([]string, 0, len(r.Meta.Grant))
	seen := make(map[string]struct{}, len(r.Meta.Grant))
	for _, login := range r.Meta.Grant {
		login = strings.TrimSpace(login)
		if _, exists := seen[login]; exists {
			continue
		}

		seen[login] = struct{}{}
		grant = append(grant, login)
	}
	r.Meta.Grant = grant
}

func (r CreateDocument) Validate() error {
	if r.Meta.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidDocument)
	}
	if r.Meta.Mime == "" {
		return fmt.Errorf("%w: mime is required", ErrInvalidDocument)
	}
	if len(r.JSON) == 0 && r.File == nil {
		return fmt.Errorf("%w: json or file is required", ErrInvalidDocument)
	}
	if r.Meta.File != (r.File != nil) {
		return fmt.Errorf("%w: file flag does not match multipart file", ErrInvalidDocument)
	}
	if len(r.JSON) != 0 && !json.Valid(r.JSON) {
		return fmt.Errorf("%w: json is malformed", ErrInvalidDocument)
	}
	for _, login := range r.Meta.Grant {
		if login == "" {
			return fmt.Errorf("%w: grant contains an empty login", ErrInvalidDocument)
		}
	}

	return nil
}

func (r *ListDocuments) Normalize() {
	r.Login = strings.TrimSpace(r.Login)
	r.Key = strings.TrimSpace(r.Key)
	r.Value = strings.TrimSpace(r.Value)
	r.Limit = strings.TrimSpace(r.Limit)
	r.Offset = strings.TrimSpace(r.Offset)
}

func (r ListDocuments) Validate() error {
	if (r.Key == "") != (r.Value == "") {
		return fmt.Errorf("%w: key and value must be provided together", ErrInvalidDocument)
	}

	return nil
}

func (r ListDocuments) ParsePagination() (int, int, error) {
	limit := defaultDocumentsLimit
	if r.Limit != "" {
		parsed, err := strconv.Atoi(r.Limit)
		if err != nil || parsed < 1 || parsed > maxDocumentsLimit {
			return 0, 0, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidDocument, maxDocumentsLimit)
		}
		limit = parsed
	}

	offset, err := parseOffset(r.Offset)
	if err != nil {
		return 0, 0, err
	}

	return limit, offset, nil
}

func parseOffset(value string) (int, error) {
	if value == "" {
		return 0, nil
	}

	offset, err := strconv.Atoi(value)
	if err != nil || offset < 0 {
		return 0, fmt.Errorf("%w: offset must be a non-negative integer", ErrInvalidDocument)
	}

	return offset, nil
}

func (r GetDocument) Validate() error {
	return validateDocumentID(r.ID)
}

func (r DeleteDocument) Validate() error {
	return validateDocumentID(r.ID)
}

func validateDocumentID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: id must be a valid UUID", ErrInvalidDocument)
	}

	return nil
}
