package player

import (
	"context"
	"errors"
	"fmt"
	"golf-game-kaffip/internal/domain/player"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlayerRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *PlayerRepository {
	return &PlayerRepository{db: db}
}

func (r *PlayerRepository) GetTeamsForGame(
	ctx context.Context,
	gameID string,
) ([]*player.Player, []*player.Player, error) {

	const q = `
        SELECT player_id, team
        FROM game_players
        WHERE game_id = $1
        ORDER BY team, player_id
    `

	rows, err := r.db.Query(ctx, q, gameID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var teamA []*player.Player
	var teamB []*player.Player

	for rows.Next() {
		var pid int64
		var team string

		if err := rows.Scan(&pid, &team); err != nil {
			return nil, nil, err
		}

		p, err := r.FindByID(ctx, pid, true)
		if err != nil {
			return nil, nil, err
		}

		if team == "A" {
			teamA = append(teamA, p)
		} else {
			teamB = append(teamB, p)
		}
	}

	return teamA, teamB, nil
}

func (r *PlayerRepository) Create(ctx context.Context, p *player.Player) error {
	query := `
        INSERT INTO players (name, email, handicap)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at;
    `

	err := r.db.QueryRow(ctx, query, p.Name, p.Email, p.Handicap).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return player.ErrEmailAlreadyExists
		}
		return fmt.Errorf("failed to insert player: %w", err)
	}

	return nil
}

func (r *PlayerRepository) FindAll(ctx context.Context) ([]*player.Player, error) {
	query := `SELECT id, name, email, handicap, created_at, updated_at FROM players WHERE deleted_at IS NULL`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []*player.Player
	for rows.Next() {
		p := &player.Player{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Email, &p.Handicap, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		players = append(players, p)
	}

	return players, nil
}

func (r *PlayerRepository) FindByID(ctx context.Context, id int64, includeDeleted bool) (*player.Player, error) {
	var query string

	if includeDeleted {
		query = `
            SELECT id, name, email, handicap, created_at, updated_at, deleted_at
            FROM players
            WHERE id = $1
        `
	} else {
		query = `
            SELECT id, name, email, handicap, created_at, updated_at, deleted_at
            FROM players
            WHERE id = $1
              AND deleted_at IS NULL
        `
	}

	p := &player.Player{}
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.Name, &p.Email, &p.Handicap, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, player.ErrPlayerNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PlayerRepository) Update(ctx context.Context, p *player.Player) error {
	cmd, err := r.db.Exec(ctx,
		`UPDATE players
         SET name = $1, email = $2, handicap = $3, updated_at = NOW()
         WHERE id = $4`,
		p.Name, p.Email, p.Handicap, p.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return player.ErrEmailAlreadyExists
		}
		return err
	}

	if cmd.RowsAffected() == 0 {
		return player.ErrPlayerNotFound
	}

	return nil
}

func (r *PlayerRepository) SoftDelete(ctx context.Context, id int64) error {
	cmd, err := r.db.Exec(ctx, `UPDATE players SET deleted_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return player.ErrPlayerNotFound
	}

	return nil
}

func (r *PlayerRepository) GetActiveGameForPlayer(ctx context.Context, id int64) (*string, error) {
	var gameID string

	err := r.db.QueryRow(ctx, `
        SELECT gp.game_id
        FROM game_players gp
        JOIN games g ON g.id = gp.game_id
        WHERE gp.player_id = $1
          AND g.finished_at IS NULL
        LIMIT 1;
    `, id).Scan(&gameID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &gameID, nil
}
