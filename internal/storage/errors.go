package storage

import "errors"

var (
	errorURLNotFound    = errors.New("url not found")
	ErrRowConflict      = errors.New("row conflict")
	ErrNotFound         = errors.New("row not found")
	ErrShortURLConflict = errors.New("provided short url already in database")
)

var (
	PgUniqueIndexErrorCode = "23505"
)
