package restapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	rediscache "github.com/maksimovyuriy/astralentry/internal/cache/redis"
	"github.com/maksimovyuriy/astralentry/internal/entity"
	documentusecase "github.com/maksimovyuriy/astralentry/internal/usecase/document"
	redislib "github.com/redis/go-redis/v9"
)

const redisTestDB = 15

type e2eAuth struct{}

func (e2eAuth) Register(context.Context, string, string, string) (entity.User, error) {
	return entity.User{}, nil
}

func (e2eAuth) Authenticate(context.Context, string, string) (string, error) {
	return "", nil
}

func (e2eAuth) Authorize(context.Context, string) (string, error) {
	return "user-id", nil
}

func (e2eAuth) Logout(context.Context, string) error {
	return nil
}

type e2eDocuments struct {
	listCalls int
}

func (s *e2eDocuments) Create(
	context.Context,
	entity.Document,
	entity.DocumentContent,
	[]string,
	io.Reader,
) (entity.Document, error) {
	return entity.Document{}, nil
}

func (s *e2eDocuments) List(
	context.Context,
	string,
	string,
	string,
	string,
	int,
	int,
) ([]entity.Document, error) {
	s.listCalls++
	return []entity.Document{{
		ID:    "550e8400-e29b-41d4-a716-446655440000",
		Name:  "cached.json",
		Mime:  "application/json",
		Grant: []string{},
	}}, nil
}

func (s *e2eDocuments) Get(
	context.Context,
	string,
	string,
) (entity.Document, entity.DocumentContent, io.ReadSeekCloser, error) {
	return entity.Document{}, entity.DocumentContent{}, nil, nil
}

func (s *e2eDocuments) Delete(context.Context, string, string) error {
	return nil
}

func TestRedisCacheEndToEnd(t *testing.T) {
	address := os.Getenv("REDIS_TEST_ADDRESS")
	if address == "" {
		t.Skip("set REDIS_TEST_ADDRESS to run Redis end-to-end test")
	}

	ctx := context.Background()
	client := redislib.NewClient(&redislib.Options{Addr: address, DB: redisTestDB})
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("connect to test Redis: %v", err)
	}

	documentCache := rediscache.NewDocument(client, 5*time.Minute)
	if err := documentCache.Invalidate(ctx); err != nil {
		t.Fatalf("clean test cache: %v", err)
	}
	t.Cleanup(func() { _ = documentCache.Invalidate(ctx) })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	next := &e2eDocuments{}
	documents := documentusecase.NewCached(next, documentCache, logger, 1024)
	handler := NewRouter(NewController(e2eAuth{}, documents, logger), logger)

	request := func(method, target string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, target, nil)
		r.Header.Set("Authorization", "Bearer session-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}

	for range 2 {
		if response := request(http.MethodGet, "/api/docs"); response.Code != http.StatusOK {
			t.Fatalf("GET /api/docs status = %d, body = %s", response.Code, response.Body.String())
		}
	}
	if next.listCalls != 1 {
		t.Fatalf("underlying List() calls after cache hit = %d, want 1", next.listCalls)
	}

	documentID := "550e8400-e29b-41d4-a716-446655440000"
	if response := request(http.MethodDelete, "/api/docs/"+documentID); response.Code != http.StatusOK {
		t.Fatalf("DELETE /api/docs/{id} status = %d, body = %s", response.Code, response.Body.String())
	}
	if response := request(http.MethodGet, "/api/docs"); response.Code != http.StatusOK {
		t.Fatalf("GET after invalidation status = %d, body = %s", response.Code, response.Body.String())
	}
	if next.listCalls != 2 {
		t.Fatalf("underlying List() calls after invalidation = %d, want 2", next.listCalls)
	}

	response := request(http.MethodPut, "/api/docs")
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("PUT /api/docs status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if !strings.Contains(response.Body.String(), `"code":405`) {
		t.Fatalf("405 response is not JSON error envelope: %s", response.Body.String())
	}
}
