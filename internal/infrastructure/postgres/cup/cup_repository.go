package cup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainCup "golf-game-kaffip/internal/domain/cup"
)

type CupRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *CupRepository {
	return &CupRepository{db: db}
}

// Create a cup row, its roster, and its match links, all
// within a single transaction
func (r *CupRepository) CreateCup(ctx context.Context, c *domainCup.Cup) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO cups (id, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`, c.ID, c.Name)
	if err != nil {
		return fmt.Errorf("failed to insert cup: %w", err)
	}

	for playerID, side := range c.Players {
		if _, err := tx.Exec(ctx, `
			INSERT INTO cup_players (cup_id, player_id, side)
			VALUES ($1, $2, $3)
		`, c.ID, playerID, string(side)); err != nil {
			return fmt.Errorf("failed to insert cup player %d: %w", playerID, err)
		}
	}

	for i, gameID := range c.MatchIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO cup_matches (cup_id, game_id, cup_position)
			VALUES ($1, $2, $3)
		`, c.ID, gameID, i); err != nil {
			return fmt.Errorf("failed to insert cup match %s: %w", gameID, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *CupRepository) LoadCup(ctx context.Context, id string) (*domainCup.Cup, error) {
	var name *string
	var createdAt, updatedAt time.Time
	var finishedAt *time.Time

	err := r.db.QueryRow(ctx, `
		SELECT name, created_at, updated_at, finished_at
		FROM cups
		WHERE id = $1
	`, id).Scan(&name, &createdAt, &updatedAt, &finishedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainCup.ErrCupNotFound
		}
		return nil, err
	}

	players, err := r.loadPlayers(ctx, id)
	if err != nil {
		return nil, err
	}
	matchIDs, err := r.loadMatchIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	cupName := ""
	if name != nil {
		cupName = *name
	}

	return &domainCup.Cup{
		ID:         id,
		Name:       cupName,
		Players:    players,
		MatchIDs:   matchIDs,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		FinishedAt: finishedAt,
	}, nil
}

func (r *CupRepository) FinishCup(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE cups SET finished_at = NOW(), update_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *CupRepository) loadPlayers(ctx context.Context, cupID string) (map[int64]domainCup.Side, error) {
	rows, err := r.db.Query(ctx, `
		SELECT player_id, side FROM cup_players WHERE cup_id = $1
	`, cupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cup players: %w", err)
	}
	defer rows.Close()

	players := make(map[int64]domainCup.Side)
	for rows.Next() {
		var playerID int64
		var side string
		if err := rows.Scan(&playerID, &side); err != nil {
			return nil, err
		}
		players[playerID] = domainCup.Side(side)
	}
	return players, rows.Err()
}

func (r *CupRepository) loadMatchIDs(ctx context.Context, cupID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT game_id FROM cup_matches WHERE cup_id = $1 ORDER BY cup_position
	`, cupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cup matches: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var gameID string
		if err := rows.Scan(&gameID); err != nil {
			return nil, err
		}
		ids = append(ids, gameID)
	}
	return ids, rows.Err()
}
