package game

import (
	"context"
)

type ListOptions struct {
	FinishedOnly bool
	ActiveOnly   bool
}

// GameRepository defines persistence operations for Game aggregates.
type Repository interface {
	LoadGame(ctx context.Context, id string) (*Game, error)
	ListSummaries(ctx context.Context, opts ListOptions) ([]*GameSummary, error)
	FinishGame(ctx context.Context, id string) error
	CreateGame(ctx context.Context, g *Game) error
	SaveHoleResult(ctx context.Context, g *Game, holeNumber int) error
	PlayersInActiveGame(ctx context.Context, playersIDs []int64) (int64, error)
	PlayerExists(ctx context.Context, id int64) (bool, error)
}
