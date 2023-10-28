package dal

import "errors"

var (
	ErrAlreadyExists = errors.New("record already exists")
	ErrNotFound      = errors.New("not found")
)
