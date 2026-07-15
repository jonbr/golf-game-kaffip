package game

import (
	"context"
	"errors"
	"fmt"
	domainGame "golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameRepository struct {
	db         *pgxpool.Pool
	playerRepo player.Repository
}

func NewRepository(db *pgxpool.Pool, playerRepo player.Repository) domainGame.Repository {
	return &GameRepository{
		db:         db,
		playerRepo: playerRepo,
	}
}

func (r *GameRepository) ListSummaries(ctx context.Context, opts domainGame.ListOptions) ([]*domainGame.GameSummary, error) {
	query := `
        SELECT g.id, g.course_id, g.course_name, g.variant, g.current_hole,
               COUNT(gch.hole_number) AS total_holes,
               g.match_team_a, g.match_team_b,
               g.created_at, g.updated_at, g.finished_at
        FROM games g
        LEFT JOIN game_course_holes gch ON gch.game_id = g.id`

	if opts.ActiveOnly {
		query += ` WHERE g.finished_at IS NULL`
	} else if opts.FinishedOnly {
		query += ` WHERE g.finished_at IS NOT NULL`
	}

	query += `
        GROUP BY g.id
        ORDER BY g.created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*domainGame.GameSummary
	for rows.Next() {
		var row GameSummaryRow
		if err := rows.Scan(
			&row.ID, &row.CourseID, &row.CourseName, &row.Variant, &row.CurrentHole,
			&row.TotalHoles, &row.MatchTeamA, &row.MatchTeamB,
			&row.CreatedAt, &row.UpdatedAt, &row.FinishedAt,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, MapGameSummaryRowToDomain(&row))
	}

	return summaries, rows.Err()
}

func (r *GameRepository) LoadGame(ctx context.Context, id string) (*domainGame.Game, error) {
	row, err := r.findGameRow(ctx, id)
	if err != nil {
		return nil, err
	}

	holes, err := r.findCourseHoles(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load course holes: %w", err)
	}

	results, scoresByResult, err := r.findHoleResults(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load hole results: %w", err)
	}

	g := MapGameRowToDomain(row, holes, results, scoresByResult)

	teamA, teamB, err := r.playerRepo.GetTeamsForGame(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load teams for game %s: %w", id, err)
	}
	g.TeamA = teamA
	g.TeamB = teamB

	return g, nil
}

func (r *GameRepository) FinishGame(ctx context.Context, id string) error {
	const q = `
        UPDATE games
        SET finished_at = NOW(), updated_at = NOW()
        WHERE id = $1
    `
	_, err := r.db.Exec(ctx, q, id)
	return err
}

// Save upserts a game. On first insert, it also snapshots the course's
// name and hole data (par, handicap index) and writes the team rosters,
// all within a single transaction — so a partial failure never leaves an
// inconsistent game record, and SetHoleScore/GetGame never need to call
// the external course API again.
func (r *GameRepository) CreateGame(ctx context.Context, g *domainGame.Game) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
        UPDATE games
        SET current_hole=$2, match_team_a=$3, match_team_b=$4, updated_at=NOW()
        WHERE id=$1
    `,
		g.ID, g.CurrentHole, g.MatchScore.TeamA, g.MatchScore.TeamB,
	)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	if res.RowsAffected() == 0 {
		_, err = tx.Exec(ctx, `
            INSERT INTO games (id, course_id, course_name, variant, starting_lead,
                                current_hole, match_team_a, match_team_b, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
        `,
			g.ID, g.Course.ID, g.Course.Name, string(g.Variant), g.StartingLead,
			g.CurrentHole, g.MatchScore.TeamA, g.MatchScore.TeamB,
		)
		if err != nil {
			return fmt.Errorf("failed to insert game: %w", err)
		}

		for _, h := range g.Course.HolesData {
			if _, err := tx.Exec(ctx, `
                INSERT INTO game_course_holes (game_id, hole_number, par, handicap_index)
                VALUES ($1, $2, $3, $4)
            `, g.ID, h.Number, h.Par, h.HandicapIndex); err != nil {
				return fmt.Errorf("failed to insert course hole %d: %w", h.Number, err)
			}
		}

		for _, p := range g.TeamA {
			if _, err := tx.Exec(ctx, `
                INSERT INTO game_players (game_id, player_id, team, created_at)
                VALUES ($1, $2, 'A', NOW())
            `, g.ID, p.ID); err != nil {
				return fmt.Errorf("failed to insert team A player: %w", err)
			}
		}

		for _, p := range g.TeamB {
			if _, err := tx.Exec(ctx, `
                INSERT INTO game_players (game_id, player_id, team, created_at)
                VALUES ($1, $2, 'B', NOW())
            `, g.ID, p.ID); err != nil {
				return fmt.Errorf("failed to insert team B player: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

// SaveHoleResult persists the result of scoring (or correcting) a single
// hole, along with the game's updated current_hole and match score,
// all within a single transaction.
func (r *GameRepository) SaveHoleResult(ctx context.Context, g *domainGame.Game, holeNumber int) error {
	result, ok := g.HoleResults[holeNumber]
	if !ok {
		return fmt.Errorf("no hole result found for hole %d", holeNumber)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
        UPDATE games
        SET current_hole=$2, match_team_a=$3, match_team_b=$4, updated_at=NOW()
        WHERE id=$1
    `, g.ID, g.CurrentHole, g.MatchScore.TeamA, g.MatchScore.TeamB); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	var lowWinner, teamWinner *string
	if result.LowScoreWinnerTeam != "" {
		lowWinner = &result.LowScoreWinnerTeam
	}
	if result.TeamTotalWinnerTeam != "" {
		teamWinner = &result.TeamTotalWinnerTeam
	}

	var holeResultID int64
	err = tx.QueryRow(ctx, `
        INSERT INTO hole_results (game_id, hole_number, points_a, points_b,
                                   low_score_winner_team, team_total_winner_team, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
        ON CONFLICT (game_id, hole_number) DO UPDATE
        SET points_a=$3, points_b=$4, low_score_winner_team=$5,
            team_total_winner_team=$6, updated_at=NOW()
        RETURNING id
    `, g.ID, holeNumber, result.PointsA, result.PointsB, lowWinner, teamWinner).Scan(&holeResultID)
	if err != nil {
		return fmt.Errorf("failed to upsert hole result: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM hole_result_scores WHERE hole_result_id=$1`, holeResultID); err != nil {
		return fmt.Errorf("failed to clear old hole scores: %w", err)
	}

	for _, s := range result.Scores {
		if _, err := tx.Exec(ctx, `
            INSERT INTO hole_result_scores (hole_result_id, player_id, gross, net, strokes, gross_bonus)
            VALUES ($1, $2, $3, $4, $5, $6)
        `, holeResultID, s.PlayerID, s.Gross, s.Net, s.Strokes, s.GrossBonus); err != nil {
			return fmt.Errorf("failed to insert score for player %d: %w", s.PlayerID, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *GameRepository) PlayersInActiveGame(ctx context.Context, playersIDs []int64) (int64, error) {
	const query = `
        SELECT gp.player_id
        FROM game_players gp
        JOIN games g ON g.id = gp.game_id
        WHERE gp.player_id = ANY($1)
          AND g.finished_at IS NULL
        LIMIT 1
    `

	var pid int64
	err := r.db.QueryRow(ctx, query, playersIDs).Scan(&pid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return pid, nil
}

func (r *GameRepository) PlayerExists(ctx context.Context, id int64) (bool, error) {
	query := `
		SELECT 1
		FROM players
		WHERE id = $1
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *GameRepository) findGameRow(ctx context.Context, id string) (*GameRow, error) {
	var row GameRow
	err := r.db.QueryRow(ctx, `
        SELECT id, course_id, course_name, variant, starting_lead, current_hole,
               match_team_a, match_team_b, created_at, updated_at, finished_at
        FROM games
        WHERE id = $1
    `, id).Scan(
		&row.ID, &row.CourseID, &row.CourseName, &row.Variant, &row.StartingLead, &row.CurrentHole,
		&row.MatchTeamA, &row.MatchTeamB, &row.CreatedAt, &row.UpdatedAt, &row.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainGame.ErrGameNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (r *GameRepository) findCourseHoles(ctx context.Context, gameID string) ([]CourseHoleRow, error) {
	rows, err := r.db.Query(ctx, `
        SELECT hole_number, par, handicap_index
        FROM game_course_holes
        WHERE game_id = $1
        ORDER BY hole_number
    `, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holes []CourseHoleRow
	for rows.Next() {
		var h CourseHoleRow
		if err := rows.Scan(&h.HoleNumber, &h.Par, &h.HandicapIndex); err != nil {
			return nil, err
		}
		holes = append(holes, h)
	}
	return holes, rows.Err()
}

func (r *GameRepository) findHoleResults(ctx context.Context, gameID string) ([]HoleResultRow, map[int64][]HoleResultScoreRow, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, game_id, hole_number, points_a, points_b,
               low_score_winner_team, team_total_winner_team
        FROM hole_results
        WHERE game_id = $1
        ORDER BY hole_number
    `, gameID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []HoleResultRow
	var resultIDs []int64
	for rows.Next() {
		var hr HoleResultRow
		if err := rows.Scan(
			&hr.ID, &hr.GameID, &hr.HoleNumber, &hr.PointsA, &hr.PointsB,
			&hr.LowScoreWinnerTeam, &hr.TeamTotalWinnerTeam,
		); err != nil {
			return nil, nil, err
		}
		results = append(results, hr)
		resultIDs = append(resultIDs, hr.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	scoresByResult, err := r.findHoleResultScores(ctx, resultIDs)
	if err != nil {
		return nil, nil, err
	}

	return results, scoresByResult, nil
}

func (r *GameRepository) findHoleResultScores(ctx context.Context, resultIDs []int64) (map[int64][]HoleResultScoreRow, error) {
	scoresByResult := make(map[int64][]HoleResultScoreRow)
	if len(resultIDs) == 0 {
		return scoresByResult, nil
	}

	rows, err := r.db.Query(ctx, `
        SELECT hole_result_id, player_id, gross, net, strokes, gross_bonus
        FROM hole_result_scores
        WHERE hole_result_id = ANY($1)
    `, resultIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var holeResultID int64
		var s HoleResultScoreRow
		if err := rows.Scan(&holeResultID, &s.PlayerID, &s.Gross, &s.Net, &s.Strokes, &s.GrossBonus); err != nil {
			return nil, err
		}
		scoresByResult[holeResultID] = append(scoresByResult[holeResultID], s)
	}
	return scoresByResult, rows.Err()
}
