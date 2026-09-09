package response

import (
	"encoding/json"
	"time"

	"github.com/maksimovyuriy/astralentry/internal/entity"
)

const documentTimeFormat = "2006-01-02 15:04:05"

type CreateDocument struct {
	JSON json.RawMessage `json:"json,omitempty"`
	File string          `json:"file,omitempty"`
}

type Document struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Mime    string   `json:"mime"`
	File    bool     `json:"file"`
	Public  bool     `json:"public"`
	Created string   `json:"created"`
	Grant   []string `json:"grant"`
}

type ListDocuments struct {
	Docs []Document `json:"docs"`
}

func ListDocumentsFromEntities(documents []entity.Document) ListDocuments {
	result := make([]Document, 0, len(documents))
	for _, document := range documents {
		result = append(result, Document{
			ID:      document.ID,
			Name:    document.Name,
			Mime:    document.Mime,
			File:    document.File,
			Public:  document.Public,
			Created: document.Created.In(time.UTC).Format(documentTimeFormat),
			Grant:   document.Grant,
		})
	}

	return ListDocuments{Docs: result}
}
