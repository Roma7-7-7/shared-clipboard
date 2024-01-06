package dal

import "errors"

const (
	pgConflictErrorCode = "23505"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrConflictUnique = errors.New("conflict unique")
)
