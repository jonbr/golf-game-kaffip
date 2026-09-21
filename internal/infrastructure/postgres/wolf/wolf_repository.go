package wolf

import (
	"context"
	"fmt"
	"golf-game-kaffip/internal/domain/wolf"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WolfRepository struct {
	db *pgxpool.Pool // Assuming you're using pgxpool for database connection
}

func NewWolfRepository(db *pgxpool.Pool) *WolfRepository {
	return &WolfRepository{
		db: db, // Initialize your database connection here
	}
}

// CreateGame persists a new Wolf game: the game row, its 4-player fixed
// rotation, and the course's hole snapshot (par, handicap index), all
// within a single transaction — so a partial failure never leaves an
// inconsistent game record, and SetHoleScore/GetGame never need to call
// the external course API again.
func (r *WolfRepository) CreateGame(ctx context.Context, g *wolf.Game) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
        INSERT INTO wolf_games (id, course_id, course_name, current_hole, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
    `, g.ID, g.Course.ID, g.Course.Name, g.CurrentHole)
	if err != nil {
		return fmt.Errorf("failed to insert wolf game: %w", err)
	}

	for seat, p := range g.Players {
		if _, err := tx.Exec(ctx, `
            INSERT INTO wolf_game_players (wolf_game_id, player_id, seat)
            VALUES ($1, $2, $3)
        `, g.ID, p.ID, seat); err != nil {
			return fmt.Errorf("failed to insert wolf player seat %d: %w", seat, err)
		}
	}

	for _, h := range g.Course.HolesData {
		if _, err := tx.Exec(ctx, `
            INSERT INTO wolf_game_course_holes (wolf_game_id, hole_number, par, handicap_index)
            VALUES ($1, $2, $3, $4)
        `, g.ID, h.Number, h.Par, h.HandicapIndex); err != nil {
			return fmt.Errorf("failed to insert wolf course hole %d: %w", h.Number, err)
		}
	}

	return tx.Commit(ctx)
}
