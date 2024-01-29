package dal

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrConflictUnique = errors.New("conflict unique")
)
