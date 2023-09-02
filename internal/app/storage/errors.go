package storage

import "errors"

var (
	errorURLNotFound = errors.New("url not found")
	ErrRowConflict   = errors.New("row conflict")
)
