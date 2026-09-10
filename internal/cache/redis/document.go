package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/maksimovyuriy/astralentry/internal/cache"
	"github.com/maksimovyuriy/astralentry/internal/entity"
	redislib "github.com/redis/go-redis/v9"
)

const (
	documentKeyPattern = "astralentry:documents:*"
	scanBatchSize      = 100
)

type Document struct {
	client *redislib.Client
	ttl    time.Duration
}

var _ cache.Document = (*Document)(nil)

func NewDocument(client *redislib.Client, ttl time.Duration) *Document {
	return &Document{client: client, ttl: ttl}
}

func (c *Document) GetList(ctx context.Context, key cache.DocumentKey) ([]entity.Document, error) {
	cacheKey, err := listKey(key)
	if err != nil {
		return nil, err
	}

	data, err := c.get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}

	var documents []entity.Document
	if err := json.Unmarshal(data, &documents); err != nil {
		return nil, fmt.Errorf("decode document list cache: %w", err)
	}

	return documents, nil
}

func (c *Document) SetList(ctx context.Context, key cache.DocumentKey, documents []entity.Document) error {
	cacheKey, err := listKey(key)
	if err != nil {
		return err
	}

	return c.set(ctx, cacheKey, documents)
}

func (c *Document) GetDocument(
	ctx context.Context,
	requesterID string,
	documentID string,
) (cache.DocumentValue, error) {
	cacheKey := documentKey(requesterID, documentID)
	data, err := c.get(ctx, cacheKey)
	if err != nil {
		return cache.DocumentValue{}, err
	}

	var value cache.DocumentValue
	if err := json.Unmarshal(data, &value); err != nil {
		return cache.DocumentValue{}, fmt.Errorf("decode document cache: %w", err)
	}

	return value, nil
}

func (c *Document) SetDocument(
	ctx context.Context,
	requesterID string,
	documentID string,
	value cache.DocumentValue,
) error {
	return c.set(ctx, documentKey(requesterID, documentID), value)
}

func (c *Document) Invalidate(ctx context.Context) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(
			ctx,
			cursor,
			documentKeyPattern,
			scanBatchSize,
		).Result()
		if err != nil {
			return fmt.Errorf("scan document cache: %w", err)
		}
		if len(keys) != 0 {
			if err := c.client.Unlink(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("unlink document cache: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			return nil
		}
	}
}

func listKey(key cache.DocumentKey) (string, error) {
	data, err := json.Marshal(key)
	if err != nil {
		return "", fmt.Errorf("encode document list cache key: %w", err)
	}
	digest := sha256.Sum256(data)

	return "astralentry:documents:list:" + hex.EncodeToString(digest[:]), nil
}

func documentKey(requesterID, documentID string) string {
	return fmt.Sprintf(
		"astralentry:documents:item:%s:%s",
		requesterID,
		documentID,
	)
}

func (c *Document) get(ctx context.Context, key string) ([]byte, error) {
	data, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redislib.Nil) {
		return nil, cache.ErrMiss
	}
	if err != nil {
		return nil, fmt.Errorf("get cache value: %w", err)
	}

	return data, nil
}

func (c *Document) set(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode cache value: %w", err)
	}
	if c.ttl <= 0 {
		return fmt.Errorf("set cache value: ttl must be positive")
	}
	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("set cache value: %w", err)
	}

	return nil
}
