package request

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"strings"
)

var ErrInvalidDocument = errors.New("invalid document")

type CreateDocument struct {
	Meta CreateDocumentMeta
	JSON json.RawMessage
	File *multipart.FileHeader
}

type CreateDocumentMeta struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
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
	if r.Meta.Name == "" || r.Meta.Mime == "" {
		return ErrInvalidDocument
	}
	if len(r.JSON) == 0 && r.File == nil {
		return ErrInvalidDocument
	}
	if r.Meta.File != (r.File != nil) {
		return ErrInvalidDocument
	}
	if len(r.JSON) != 0 && !json.Valid(r.JSON) {
		return ErrInvalidDocument
	}
	for _, login := range r.Meta.Grant {
		if login == "" {
			return ErrInvalidDocument
		}
	}

	return nil
}
