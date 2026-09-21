package wolf

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	domainWolf "golf-game-kaffip/internal/domain/wolf"
	gamedb "golf-game-kaffip/internal/infrastructure/postgres/game"
)

type WolfRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *WolfRepository {
	return &WolfRepository{db: db}
}

// CreateGame persists a new Wolf game: the game row, its 4-player fixed
// rotation, and the course's hole snapshot (par, handicap index), all
// within a single transaction — so a partial failure never leaves an
// inconsistent game record, and SetHoleScore/GetGame never need to call
// the external course API again.
func (r *WolfRepository) CreateGame(ctx context.Context, g *domainWolf.Game) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := gamedb.InsertGameRow(ctx, tx, gamedb.GameInsertParams{
		ID:          g.ID,
		GameType:    "wolf",
		CourseID:    g.Course.ID,
		CourseName:  g.Course.Name,
		Variant:     "gross", // unused by wolf, column default
		CurrentHole: g.CurrentHole,
	}); err != nil {
		return err
	}

	for seat, p := range g.Players {
		s := seat
		if err := gamedb.InsertGamePlayer(ctx, tx, g.ID, p.ID, nil, &s); err != nil {
			return err
		}
	}

	for _, h := range g.Course.HolesData {
		if err := gamedb.InsertCourseHole(ctx, tx, g.ID, h.Number, h.Par, h.HandicapIndex); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
