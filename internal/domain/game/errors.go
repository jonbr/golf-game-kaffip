package game

import (
	"errors"
	"fmt"
)

var ErrGameNotFound = errors.New("game not found")

type PlayersInActiveGame struct {
	PlayerID string
	GameID   string
}

// Returned when a player is already in an active game.
func (e PlayersInActiveGame) Error() string {
	return fmt.Sprintf("player %s in another active game %s", e.PlayerID, e.GameID)
}

// Returned when a player attempts to score a hole but is not part of the game.
type ErrPlayerNotInGame struct {
	PlayerID string
	GameID   string
}

func (e ErrPlayerNotInGame) Error() string {
	return fmt.Sprintf("player %s not in game %s", e.PlayerID, e.GameID)
}

// Returned when a gross score is invalid (<= 0 or otherwise rejected).
type ErrInvalidGrossScore struct {
	PlayerID string
	GameID   string
	Gross    int
}

func (e ErrInvalidGrossScore) Error() string {
	return fmt.Sprintf("invalid gross score %d for player %s in game %s", e.Gross, e.PlayerID, e.GameID)
}

// Returned when the requested hole number is outside the course range.
type ErrHoleOutOfRange struct {
	Hole   int
	GameID string
}

func (e ErrHoleOutOfRange) Error() string {
	return fmt.Sprintf("hole %d out of range in game %s", e.Hole, e.GameID)
}

// Returned when the client tries to score a hole that is not the current hole.
type ErrWrongHole struct {
	Expected int
	Got      int
	GameID   string
}

func (e ErrWrongHole) Error() string {
	return fmt.Sprintf("wrong hole: expected %d, got %d in game %s", e.Expected, e.Got, e.GameID)
}

// Returned when scoring is attempted after the game has already finished.
type ErrGameFinished struct {
	GameID string
}

func (e ErrGameFinished) Error() string {
	return fmt.Sprintf("game %s already finished", e.GameID)
}
