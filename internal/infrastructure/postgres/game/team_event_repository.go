package game

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainGame "golf-game-kaffip/internal/domain/game"
)

type TeamEventRow struct {
	ID         string
	CourseID   string
	CourseName string
	Variant    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	FinishedAt *time.Time
}

type TeamEventRepository struct {
	db       *pgxpool.Pool
	gameRepo domainGame.Repository
}

func NewTeamEventRepository(db *pgxpool.Pool, gameRepo domainGame.Repository) *TeamEventRepository {
	return &TeamEventRepository{db: db, gameRepo: gameRepo}
}

// CreateEvent persists the event row and every match's game row, course
// hole snapshot, and player roster, all within a single transaction.
func (r *TeamEventRepository) CreateEvent(ctx context.Context, event *domainGame.TeamEvent) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
        INSERT INTO team_events (id, course_id, course_name, variant, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
    `, event.ID, event.Course.ID, event.Course.Name, string(event.Variant))
	if err != nil {
		return fmt.Errorf("failed to insert team event: %w", err)
	}

	for i, m := range event.Matches {
		_, err = tx.Exec(ctx, `
            INSERT INTO games (id, course_id, course_name, game_type, variant, starting_lead,
                                current_hole, match_team_a, match_team_b, team_event_id,
                                event_position, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
        `,
			m.ID, m.Course.ID, m.Course.Name, string(m.GameType), string(m.Variant), m.StartingLead,
			m.CurrentHole, m.MatchScore.TeamA, m.MatchScore.TeamB, event.ID, i,
		)
		if err != nil {
			return fmt.Errorf("failed to insert match %s: %w", m.ID, err)
		}

		for _, h := range m.Course.HolesData {
			if _, err := tx.Exec(ctx, `
                INSERT INTO game_course_holes (game_id, hole_number, par, handicap_index)
                VALUES ($1, $2, $3, $4)
            `, m.ID, h.Number, h.Par, h.HandicapIndex); err != nil {
				return fmt.Errorf("failed to insert course hole %d for match %s: %w", h.Number, m.ID, err)
			}
		}

		for _, p := range m.TeamA {
			if _, err := tx.Exec(ctx, `
                INSERT INTO game_players (game_id, player_id, team, created_at)
                VALUES ($1, $2, 'A', NOW())
            `, m.ID, p.ID); err != nil {
				return fmt.Errorf("failed to insert team A player for match %s: %w", m.ID, err)
			}
		}
		for _, p := range m.TeamB {
			if _, err := tx.Exec(ctx, `
                INSERT INTO game_players (game_id, player_id, team, created_at)
                VALUES ($1, $2, 'B', NOW())
            `, m.ID, p.ID); err != nil {
				return fmt.Errorf("failed to insert team B player for match %s: %w", m.ID, err)
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *TeamEventRepository) LoadEvent(ctx context.Context, id string) (*domainGame.TeamEvent, error) {
	var row TeamEventRow
	err := r.db.QueryRow(ctx, `
        SELECT id, course_id, course_name, variant, created_at, updated_at, finished_at
        FROM team_events
        WHERE id = $1
    `, id).Scan(&row.ID, &row.CourseID, &row.CourseName, &row.Variant, &row.CreatedAt, &row.UpdatedAt, &row.FinishedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainGame.ErrTeamEventNotFound
		}
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
        SELECT id FROM games WHERE team_event_id = $1 ORDER BY event_position
    `, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load match ids: %w", err)
	}
	var matchIDs []string
	for rows.Next() {
		var mid string
		if err := rows.Scan(&mid); err != nil {
			rows.Close()
			return nil, err
		}
		matchIDs = append(matchIDs, mid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(matchIDs) == 0 {
		return nil, fmt.Errorf("event %s has no matches", id)
	}

	matches := make([]*domainGame.Game, 0, len(matchIDs))
	for _, mid := range matchIDs {
		m, err := r.gameRepo.LoadGame(ctx, mid)
		if err != nil {
			return nil, fmt.Errorf("failed to load match %s: %w", mid, err)
		}
		matches = append(matches, m)
	}

	return &domainGame.TeamEvent{
		ID:         row.ID,
		Course:     matches[0].Course,
		Variant:    domainGame.Variant(row.Variant),
		MatchIDs:   matchIDs,
		Matches:    matches,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		FinishedAt: row.FinishedAt,
	}, nil
}

func (r *TeamEventRepository) FinishEvent(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
        UPDATE team_events SET finished_at = NOW(), updated_at = NOW() WHERE id = $1
    `, id)
	return err
}
