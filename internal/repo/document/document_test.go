package document

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/maksimovyuriy/astralentry/internal/repo"
)

func TestDocumentFilter(t *testing.T) {
	valid := map[string]string{
		"":        "",
		"id":      "550e8400-e29b-41d4-a716-446655440000",
		"name":    "photo.jpg",
		"mime":    "image/jpeg",
		"file":    "true",
		"public":  "false",
		"created": "2026-09-09 12:30:00",
	}
	for key, value := range valid {
		if _, _, err := documentFilter(key, value); err != nil {
			t.Errorf("documentFilter(%q, %q) error = %v", key, value, err)
		}
	}

	invalid := map[string]string{
		"unknown": "value",
		"id":      "not-a-uuid",
		"file":    "yes",
		"public":  "no",
		"created": "not-a-date",
	}
	for key, value := range invalid {
		if _, _, err := documentFilter(key, value); !errors.Is(err, repo.ErrInvalidFilter) {
			t.Errorf("documentFilter(%q, %q) error = %v, want %v", key, value, err, repo.ErrInvalidFilter)
		}
	}
}

func TestBuildListQuery(t *testing.T) {
	t.Run("without filter", func(t *testing.T) {
		query, arguments, err := buildListQuery("owner", "requester", "", "", 20, 40)
		if err != nil {
			t.Fatalf("buildListQuery() error = %v", err)
		}
		if !strings.Contains(query, "LIMIT $3 OFFSET $4") {
			t.Errorf("buildListQuery() query has incorrect pagination placeholders")
		}
		want := []any{"owner", "requester", 20, 40}
		if !reflect.DeepEqual(arguments, want) {
			t.Errorf("buildListQuery() arguments = %#v, want %#v", arguments, want)
		}
	})

	t.Run("with filter", func(t *testing.T) {
		query, arguments, err := buildListQuery("owner", "requester", "name", "photo.jpg", 10, 5)
		if err != nil {
			t.Fatalf("buildListQuery() error = %v", err)
		}
		if !strings.Contains(query, "AND d.name = $3") {
			t.Errorf("buildListQuery() query does not contain filter placeholder")
		}
		if !strings.Contains(query, "LIMIT $4 OFFSET $5") {
			t.Errorf("buildListQuery() query has incorrect pagination placeholders")
		}
		want := []any{"owner", "requester", "photo.jpg", 10, 5}
		if !reflect.DeepEqual(arguments, want) {
			t.Errorf("buildListQuery() arguments = %#v, want %#v", arguments, want)
		}
	})
}
