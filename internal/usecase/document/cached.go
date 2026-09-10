package document

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/maksimovyuriy/astralentry/internal/cache"
	"github.com/maksimovyuriy/astralentry/internal/entity"
	"github.com/maksimovyuriy/astralentry/internal/usecase"
)

type Cached struct {
	next        usecase.Document
	cache       cache.Document
	logger      *slog.Logger
	maxFileSize int64
}

var _ usecase.Document = (*Cached)(nil)

func NewCached(
	next usecase.Document,
	documentCache cache.Document,
	logger *slog.Logger,
	maxFileSize int64,
) usecase.Document {
	return &Cached{
		next:        next,
		cache:       documentCache,
		logger:      logger,
		maxFileSize: maxFileSize,
	}
}

func (uc *Cached) Create(
	ctx context.Context,
	document entity.Document,
	content entity.DocumentContent,
	grant []string,
	file io.Reader,
) (entity.Document, error) {
	created, err := uc.next.Create(ctx, document, content, grant, file)
	if err != nil {
		return entity.Document{}, err
	}

	uc.invalidate(ctx)
	return created, nil
}

func (uc *Cached) List(
	ctx context.Context,
	requesterID string,
	login string,
	key string,
	value string,
	limit int,
	offset int,
) ([]entity.Document, error) {
	cacheKey := cache.DocumentKey{
		RequesterID: requesterID,
		Login:       login,
		Key:         key,
		Value:       value,
		Limit:       limit,
		Offset:      offset,
	}

	documents, err := uc.cache.GetList(ctx, cacheKey)
	if err == nil {
		return documents, nil
	}
	uc.logCacheError("get document list", err)

	documents, err = uc.next.List(ctx, requesterID, login, key, value, limit, offset)
	if err != nil {
		return nil, err
	}

	if err := uc.cache.SetList(ctx, cacheKey, documents); err != nil {
		uc.logCacheError("set document list", err)
	}

	return documents, nil
}

func (uc *Cached) Get(
	ctx context.Context,
	requesterID string,
	documentID string,
) (entity.Document, entity.DocumentContent, io.ReadSeekCloser, error) {
	cached, err := uc.cache.GetDocument(ctx, requesterID, documentID)
	if err == nil {
		return cached.Document, cached.Content, cachedFile(cached), nil
	}
	uc.logCacheError("get document", err)

	document, content, file, err := uc.next.Get(ctx, requesterID, documentID)
	if err != nil {
		return entity.Document{}, entity.DocumentContent{}, nil, err
	}
	if file == nil {
		uc.setDocument(ctx, requesterID, documentID, cache.DocumentValue{
			Document: document,
			Content:  content,
		})
		return document, content, nil, nil
	}

	return uc.cacheFile(ctx, requesterID, documentID, document, content, file)
}

func (uc *Cached) Delete(ctx context.Context, requesterID, documentID string) error {
	if err := uc.next.Delete(ctx, requesterID, documentID); err != nil {
		return err
	}

	uc.invalidate(ctx)
	return nil
}

func (uc *Cached) cacheFile(
	ctx context.Context,
	requesterID string,
	documentID string,
	document entity.Document,
	content entity.DocumentContent,
	file io.ReadSeekCloser,
) (entity.Document, entity.DocumentContent, io.ReadSeekCloser, error) {
	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		uc.logCacheError("measure document file", err)
		return document, content, file, nil
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		return entity.Document{}, entity.DocumentContent{}, nil, err
	}
	if uc.maxFileSize <= 0 || size > uc.maxFileSize {
		return document, content, file, nil
	}

	data, err := io.ReadAll(file)
	if err != nil {
		file.Close()
		return entity.Document{}, entity.DocumentContent{}, nil, err
	}
	if err := file.Close(); err != nil {
		return entity.Document{}, entity.DocumentContent{}, nil, err
	}

	cached := cache.DocumentValue{Document: document, Content: content, File: data}
	uc.setDocument(ctx, requesterID, documentID, cached)
	return document, content, cachedFile(cached), nil
}

func (uc *Cached) setDocument(
	ctx context.Context,
	requesterID string,
	documentID string,
	value cache.DocumentValue,
) {
	if err := uc.cache.SetDocument(ctx, requesterID, documentID, value); err != nil {
		uc.logCacheError("set document", err)
	}
}

func (uc *Cached) invalidate(ctx context.Context) {
	if err := uc.cache.Invalidate(ctx); err != nil {
		uc.logCacheError("invalidate documents", err)
	}
}

func (uc *Cached) logCacheError(operation string, err error) {
	if errors.Is(err, cache.ErrMiss) {
		return
	}
	uc.logger.Warn("Document cache unavailable", "operation", operation, "error", err)
}

func cachedFile(value cache.DocumentValue) io.ReadSeekCloser {
	if !value.Document.File {
		return nil
	}

	return &readSeekCloser{Reader: bytes.NewReader(value.File)}
}

type readSeekCloser struct {
	*bytes.Reader
}

func (readSeekCloser) Close() error {
	return nil
}
