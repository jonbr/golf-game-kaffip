package wolf

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"golf-game-kaffip/internal/domain/course"
	"golf-game-kaffip/internal/domain/player"
	domainWolf "golf-game-kaffip/internal/domain/wolf"
	gamedb "golf-game-kaffip/internal/infrastructure/postgres/game"
)

type WolfRepository struct {
	db         *pgxpool.Pool
	playerRepo player.Repository
}

type wolfScoreRow struct {
	playerID int64
	gross    int
	net      int
	strokes  int
	points   int
}

func NewRepository(db *pgxpool.Pool, playerRepo player.Repository) *WolfRepository {
	return &WolfRepository{
		db:         db,
		playerRepo: playerRepo,
	}
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
		Variant:     "net", // unused by wolf, column default
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

func (r *WolfRepository) LoadGame(ctx context.Context, id string) (*domainWolf.Game, error) {
	var courseID, courseName string
	var currentHole int
	var createdAt, updatedAt time.Time
	var finishedAt *time.Time

	err := r.db.QueryRow(ctx, `
        SELECT course_id, course_name, current_hole, created_at, updated_at, finished_at
        FROM games
        WHERE id = $1 AND game_type = 'wolf'
    `, id).Scan(&courseID, &courseName, &currentHole, &createdAt, &updatedAt, &finishedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainWolf.ErrGameNotFound
		}
		return nil, err
	}

	playerIDs, err := r.loadPlayerSeats(ctx, id)
	if err != nil {
		return nil, err
	}

	var players [4]*player.Player
	for i, pid := range playerIDs {
		p, err := r.playerRepo.FindByID(ctx, pid, false)
		if err != nil {
			return nil, fmt.Errorf("failed to load player %d: %w", pid, err)
		}
		players[i] = p
	}

	c, err := r.loadCourse(ctx, id, courseID, courseName)
	if err != nil {
		return nil, err
	}

	holeResults, err := r.loadHoleResults(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domainWolf.Game{
		ID:          id,
		GameType:    "wolf",
		Course:      c,
		Players:     players,
		CurrentHole: currentHole,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		FinishedAt:  finishedAt,
		HoleResults: holeResults,
	}, nil
}

func (r *WolfRepository) SaveHoleResult(ctx context.Context, g *domainWolf.Game, holeNumber int) error {
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
        SET current_hole=$2, updated_at=NOW()
        WHERE id=$1
    `, g.ID, g.CurrentHole); err != nil {
		return fmt.Errorf("failed to update wolf game: %w", err)
	}

	var winningSide *string
	if result.WinningSide != "" {
		s := string(result.WinningSide)
		winningSide = &s
	}

	var holeResultID int64
	err = tx.QueryRow(ctx, `
        INSERT INTO wolf_hole_results (game_id, hole_number, wolf_player_id, mode, partner_id, winning_side, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
        ON CONFLICT (game_id, hole_number) DO UPDATE
        SET wolf_player_id=$3, mode=$4, partner_id=$5, winning_side=$6, updated_at=NOW()
        RETURNING id
    `, g.ID, holeNumber, result.WolfPlayerID, string(result.Mode), result.PartnerID, winningSide).Scan(&holeResultID)
	if err != nil {
		return fmt.Errorf("failed to upsert wolf hole result: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM wolf_hole_scores WHERE wolf_hole_result_id=$1`, holeResultID); err != nil {
		return fmt.Errorf("failed to clear old wolf hole scores: %w", err)
	}

	for _, s := range result.Scores {
		points := result.PointsAwarded[s.PlayerID]
		if _, err := tx.Exec(ctx, `
            INSERT INTO wolf_hole_scores (wolf_hole_result_id, player_id, gross, net, strokes, points)
            VALUES ($1, $2, $3, $4, $5, $6)
        `, holeResultID, s.PlayerID, s.Gross, s.Net, s.Strokes, points); err != nil {
			return fmt.Errorf("failed to insert wolf score for player %d: %w", s.PlayerID, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *WolfRepository) loadPlayerSeats(ctx context.Context, gameID string) ([]int64, error) {
	rows, err := r.db.Query(ctx, `
        SELECT player_id FROM game_players
        WHERE game_id = $1
        ORDER BY seat
    `, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to load wolf players: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			return nil, err
		}
		ids = append(ids, pid)
	}
	if len(ids) != 4 {
		return nil, fmt.Errorf("wolf game %s has %d players, expected 4", gameID, len(ids))
	}
	return ids, nil
}

func (r *WolfRepository) loadCourse(ctx context.Context, gameID, courseID, courseName string) (*course.Course, error) {
	rows, err := r.db.Query(ctx, `
        SELECT hole_number, par, handicap_index FROM game_course_holes
        WHERE game_id = $1 ORDER BY hole_number
    `, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	c := &course.Course{ID: courseID, Name: courseName}
	for rows.Next() {
		var h course.Hole
		if err := rows.Scan(&h.Number, &h.Par, &h.HandicapIndex); err != nil {
			return nil, err
		}
		c.HolesData = append(c.HolesData, h)
	}
	c.Holes = len(c.HolesData)
	return c, nil
}

func (r *WolfRepository) loadHoleResults(ctx context.Context, gameID string) (map[int]*domainWolf.HoleResult, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, hole_number, wolf_player_id, mode, partner_id, winning_side
        FROM wolf_hole_results
        WHERE game_id = $1
        ORDER BY hole_number
    `, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to load wolf hole results: %w", err)
	}

	type resultRow struct {
		id           int64
		holeNumber   int
		wolfPlayerID int64
		mode         string
		partnerID    *int64
		winningSide  *string
	}

	var rowsData []resultRow
	var resultIDs []int64
	for rows.Next() {
		var rr resultRow
		if err := rows.Scan(&rr.id, &rr.holeNumber, &rr.wolfPlayerID, &rr.mode, &rr.partnerID, &rr.winningSide); err != nil {
			rows.Close()
			return nil, err
		}
		rowsData = append(rowsData, rr)
		resultIDs = append(resultIDs, rr.id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	scoresByResult, err := r.loadHoleScores(ctx, resultIDs)
	if err != nil {
		return nil, err
	}

	holeResults := make(map[int]*domainWolf.HoleResult, len(rowsData))
	for _, rr := range rowsData {
		winningSide := domainWolf.WolfSide("")
		if rr.winningSide != nil {
			winningSide = domainWolf.WolfSide(*rr.winningSide)
		}

		scores := scoresByResult[rr.id]
		points := make(map[int64]int, len(scores))
		playerHoleResults := make([]domainWolf.PlayerHoleResult, 0, len(scores))
		for _, s := range scores {
			playerHoleResults = append(playerHoleResults, domainWolf.PlayerHoleResult{
				PlayerID: s.playerID, Gross: s.gross, Net: s.net, Strokes: s.strokes,
			})
			points[s.playerID] = s.points
		}

		holeResults[rr.holeNumber] = &domainWolf.HoleResult{
			Hole:          domainWolf.HoleInfo{Number: rr.holeNumber},
			WolfPlayerID:  rr.wolfPlayerID,
			Mode:          domainWolf.WolfMode(rr.mode),
			PartnerID:     rr.partnerID,
			Scores:        playerHoleResults,
			WinningSide:   winningSide,
			PointsAwarded: points,
		}
	}

	return holeResults, nil
}

func (r *WolfRepository) loadHoleScores(ctx context.Context, resultIDs []int64) (map[int64][]wolfScoreRow, error) {
	scoresByResult := make(map[int64][]wolfScoreRow)
	if len(resultIDs) == 0 {
		return scoresByResult, nil
	}

	rows, err := r.db.Query(ctx, `
        SELECT wolf_hole_result_id, player_id, gross, net, strokes, points
        FROM wolf_hole_scores
        WHERE wolf_hole_result_id = ANY($1)
    `, resultIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var resultID int64
		var s wolfScoreRow
		if err := rows.Scan(&resultID, &s.playerID, &s.gross, &s.net, &s.strokes, &s.points); err != nil {
			return nil, err
		}
		scoresByResult[resultID] = append(scoresByResult[resultID], s)
	}
	return scoresByResult, rows.Err()
}
