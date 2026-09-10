package cache

import (
	"context"
	"errors"

	"github.com/maksimovyuriy/astralentry/internal/entity"
)

var ErrMiss = errors.New("cache miss")

type DocumentKey struct {
	RequesterID string
	Login       string
	Key         string
	Value       string
	Limit       int
	Offset      int
}

type DocumentValue struct {
	Document entity.Document        `json:"document"`
	Content  entity.DocumentContent `json:"content"`
	File     []byte                 `json:"file,omitempty"`
}

type Document interface {
	GetList(ctx context.Context, key DocumentKey) ([]entity.Document, error)
	SetList(ctx context.Context, key DocumentKey, documents []entity.Document) error
	GetDocument(ctx context.Context, requesterID, documentID string) (DocumentValue, error)
	SetDocument(ctx context.Context, requesterID, documentID string, value DocumentValue) error
	Invalidate(ctx context.Context) error
}
