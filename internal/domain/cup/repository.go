package cup

import "context"

type Repository interface {
	CreateCup(ctx context.Context, c *Cup) error
	LoadCup(ctx context.Context, id string) (*Cup, error)
	FinishCup(ctx context.Context, id string) error
}
