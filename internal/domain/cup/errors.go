package cup

import "errors"

var (
	errEmptyID         = errors.New("cup id cannot be empty")
	errNoPlayers       = errors.New("cup requires at least one player")
	errNoMatches       = errors.New("cup requires at least one match")
	errIncompleteSides = errors.New("cup requires at least one player on each side")
	ErrCupNotFound     = errors.New("cup not found")
)
