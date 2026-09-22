package wolf

import "context"

type Repository interface {
	CreateGame(ctx context.Context, game *Game) error
	LoadGame(ctx context.Context, id string) (*Game, error)
}
