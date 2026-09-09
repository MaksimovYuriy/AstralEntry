package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Local struct {
	root string
}

var _ Storage = (*Local)(nil)

func NewLocal(root string) (*Local, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("storage root: %w", ErrInvalidKey)
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}
	if err := os.MkdirAll(absoluteRoot, 0o750); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}

	return &Local{root: absoluteRoot}, nil
}

func (s *Local) Save(key string, source io.Reader) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}

	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return fmt.Errorf("create storage directory: %w", err)
	}

	temporary, err := os.CreateTemp(directory, ".upload-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	removeTemporary := true
	defer func() {
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()

	if _, err := io.Copy(temporary, source); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("store file: %w", err)
	}

	removeTemporary = false
	return nil
}

func (s *Local) Open(key string) (ReadSeekCloser, error) {
	path, err := s.path(key)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return file, nil
}

func (s *Local) Delete(key string) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}

	if err := os.Remove(path); errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

func (s *Local) path(key string) (string, error) {
	cleanKey := filepath.Clean(key)
	if cleanKey == "." || filepath.IsAbs(cleanKey) || cleanKey == ".." ||
		strings.HasPrefix(cleanKey, ".."+string(filepath.Separator)) {
		return "", ErrInvalidKey
	}

	return filepath.Join(s.root, cleanKey), nil
}
