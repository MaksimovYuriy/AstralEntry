package document

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/maksimovyuriy/astralentry/internal/cache"
	"github.com/maksimovyuriy/astralentry/internal/entity"
)

type documentUseCaseStub struct {
	listCalls int
	documents []entity.Document
}

func (s *documentUseCaseStub) Create(
	context.Context,
	entity.Document,
	entity.DocumentContent,
	[]string,
	io.Reader,
) (entity.Document, error) {
	return entity.Document{}, nil
}

func (s *documentUseCaseStub) List(
	context.Context,
	string,
	string,
	string,
	string,
	int,
	int,
) ([]entity.Document, error) {
	s.listCalls++
	return s.documents, nil
}

func (s *documentUseCaseStub) Get(
	context.Context,
	string,
	string,
) (entity.Document, entity.DocumentContent, io.ReadSeekCloser, error) {
	return entity.Document{}, entity.DocumentContent{}, nil, nil
}

func (s *documentUseCaseStub) Delete(context.Context, string, string) error {
	return nil
}

type documentCacheStub struct {
	lists         map[cache.DocumentKey][]entity.Document
	invalidations int
}

func (s *documentCacheStub) GetList(_ context.Context, key cache.DocumentKey) ([]entity.Document, error) {
	documents, ok := s.lists[key]
	if !ok {
		return nil, cache.ErrMiss
	}
	return documents, nil
}

func (s *documentCacheStub) SetList(
	_ context.Context,
	key cache.DocumentKey,
	documents []entity.Document,
) error {
	s.lists[key] = documents
	return nil
}

func (s *documentCacheStub) GetDocument(context.Context, string, string) (cache.DocumentValue, error) {
	return cache.DocumentValue{}, cache.ErrMiss
}

func (s *documentCacheStub) SetDocument(context.Context, string, string, cache.DocumentValue) error {
	return nil
}

func (s *documentCacheStub) Invalidate(context.Context) error {
	s.invalidations++
	clear(s.lists)
	return nil
}

func TestCachedDocumentLifecycle(t *testing.T) {
	next := &documentUseCaseStub{documents: []entity.Document{{ID: "document-id"}}}
	documentCache := &documentCacheStub{lists: make(map[cache.DocumentKey][]entity.Document)}
	useCase := NewCached(next, documentCache, slog.New(slog.NewTextHandler(io.Discard, nil)), 1024)
	ctx := context.Background()

	for range 2 {
		documents, err := useCase.List(ctx, "user-id", "", "", "", 20, 0)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(documents) != 1 || documents[0].ID != "document-id" {
			t.Fatalf("List() documents = %#v", documents)
		}
	}

	if next.listCalls != 1 {
		t.Errorf("underlying List() calls = %d, want 1", next.listCalls)
	}

	if _, err := useCase.Create(
		ctx,
		entity.Document{},
		entity.DocumentContent{},
		nil,
		nil,
	); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := useCase.List(ctx, "user-id", "", "", "", 20, 0); err != nil {
		t.Fatalf("List() after invalidation error = %v", err)
	}
	if next.listCalls != 2 {
		t.Errorf("underlying List() calls after invalidation = %d, want 2", next.listCalls)
	}

	if err := useCase.Delete(ctx, "user-id", "document-id"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if documentCache.invalidations != 2 {
		t.Errorf("Invalidate() calls = %d, want 2", documentCache.invalidations)
	}
}
