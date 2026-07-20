package player

import "errors"

var (
	ErrPlayerNotFound     = errors.New("player not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)
