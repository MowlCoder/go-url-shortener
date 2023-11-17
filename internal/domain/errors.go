package domain

import "errors"

// All available domain errors. They can occur during service working.
var (
	ErrURLConflict      = errors.New("url conflict")
	ErrURLNotFound      = errors.New("url not found")
	ErrShortURLConflict = errors.New("provided short url already exists")
)
