package storage

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestLocalSaveOpenDelete(t *testing.T) {
	storage, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	const key = "documents/document-id"
	const content = "file content"
	if err := storage.Save(key, strings.NewReader(content)); err != nil {
		t.Fatalf("save file: %v", err)
	}

	file, err := storage.Open(key)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}
	if string(data) != content {
		t.Fatalf("content = %q, want %q", data, content)
	}

	if err := storage.Delete(key); err != nil {
		t.Fatalf("delete file: %v", err)
	}
	if _, err := storage.Open(key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("open deleted file error = %v, want %v", err, ErrNotFound)
	}
}

func TestLocalRejectsUnsafeKeys(t *testing.T) {
	storage, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	for _, key := range []string{"", ".", "..", "../outside", "/absolute"} {
		if err := storage.Save(key, strings.NewReader("content")); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("Save(%q) error = %v, want %v", key, err, ErrInvalidKey)
		}
	}
}
