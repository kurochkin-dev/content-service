package article

import "errors"

var (
	ErrNotFound   = errors.New("article not found")
	ErrForbidden  = errors.New("forbidden: you can only manage your own articles")
	ErrValidation = errors.New("validation error")
)
