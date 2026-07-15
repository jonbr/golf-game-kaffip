package player

import (
	"context"
)

type Repository interface {
	GetTeamsForGame(ctx context.Context, gameID string) ([]*Player, []*Player, error)

	Create(ctx context.Context, p *Player) error
	FindAll(ctx context.Context) ([]*Player, error)
	FindByID(ctx context.Context, id int64, includeDeleted bool) (*Player, error)
	SoftDelete(ctx context.Context, id int64) error
	Update(ctx context.Context, p *Player) error

	GetActiveGameForPlayer(ctx context.Context, id int64) (*string, error)
}
