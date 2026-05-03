package storage

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("storage: not found")
	ErrAlreadyExists = errors.New("storage: already exists")
	ErrInvalidInput  = errors.New("storage: invalid input")
	ErrTimeout       = errors.New("storage: operation timeout")
)

type StorageError struct {
	Op  string
	Err error
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage.%s: %v", e.Op, e.Err)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}
