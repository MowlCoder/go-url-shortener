package domain

import "errors"

var (
	ErrURLConflict      = errors.New("url conflict")
	ErrURLNotFound      = errors.New("url not found")
	ErrShortURLConflict = errors.New("provided short url already exists")
)
