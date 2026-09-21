package wolf

import "errors"

var (
	errEmptyID       = errors.New("wolf game id cannot be empty")
	errMissingPlayer = errors.New("wolf game must have 4 players")
	ErrGameNotFound  = errors.New("wolf game not found")
)
