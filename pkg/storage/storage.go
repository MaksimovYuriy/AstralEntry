package storage

import (
	"errors"
	"io"
)

var (
	ErrInvalidKey = errors.New("invalid storage key")
	ErrNotFound   = errors.New("file not found")
)

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type Storage interface {
	Save(key string, source io.Reader) error
	Open(key string) (ReadSeekCloser, error)
	Delete(key string) error
}
