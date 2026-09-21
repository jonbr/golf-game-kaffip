package game

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// InsertGameRow writes one row into the shared games table. Used by every
// game type's repository (points_play, match_play, wolf, team_event) —
// callers that don't use a given column (e.g. wolf has no match score)
// pass its zero value.
func InsertGameRow(ctx context.Context, tx pgx.Tx, params GameInsertParams) error {
	_, err := tx.Exec(ctx, `
        INSERT INTO games (id, game_type, course_id, course_name, variant, starting_lead,
                            current_hole, match_team_a, match_team_b, team_event_id,
                            event_position, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
    `,
		params.ID, params.GameType, params.CourseID, params.CourseName, params.Variant,
		params.StartingLead, params.CurrentHole, params.MatchTeamA, params.MatchTeamB,
		params.TeamEventID, params.EventPosition,
	)
	if err != nil {
		return fmt.Errorf("failed to insert game row (%s): %w", params.ID, err)
	}
	return nil
}

type GameInsertParams struct {
	ID            string
	GameType      string
	CourseID      string
	CourseName    string
	Variant       string // "gross" default if the game type doesn't use variant
	StartingLead  int
	CurrentHole   int
	MatchTeamA    int
	MatchTeamB    int
	TeamEventID   *string
	EventPosition *int
}

// InsertGamePlayer writes one row into the shared game_players table.
// Pass team for two-sided formats (points_play/match_play/team_event
// matches), or seat for wolf's fixed rotation — leave the other nil.
func InsertGamePlayer(ctx context.Context, tx pgx.Tx, gameID string, playerID int64, team *string, seat *int) error {
	_, err := tx.Exec(ctx, `
        INSERT INTO game_players (game_id, player_id, team, seat, created_at)
        VALUES ($1, $2, $3, $4, NOW())
    `, gameID, playerID, team, seat)
	if err != nil {
		return fmt.Errorf("failed to insert game player %d for game %s: %w", playerID, gameID, err)
	}
	return nil
}

// InsertCourseHole writes one row into the shared game_course_holes table.
func InsertCourseHole(ctx context.Context, tx pgx.Tx, gameID string, holeNumber, par, handicapIndex int) error {
	_, err := tx.Exec(ctx, `
        INSERT INTO game_course_holes (game_id, hole_number, par, handicap_index)
        VALUES ($1, $2, $3, $4)
    `, gameID, holeNumber, par, handicapIndex)
	if err != nil {
		return fmt.Errorf("failed to insert course hole %d for game %s: %w", holeNumber, gameID, err)
	}
	return nil
}
