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

	rows, err := r.db.Query(ctx, `
        SELECT player_id FROM game_players
        WHERE game_id = $1
        ORDER BY seat
    `, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load wolf players: %w", err)
	}
	defer rows.Close()

	var playerIDs []int64
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			return nil, err
		}
		playerIDs = append(playerIDs, pid)
	}
	if len(playerIDs) != 4 {
		return nil, fmt.Errorf("wolf game %s has %d players, expected 4", id, len(playerIDs))
	}

	var players [4]*player.Player
	for i, pid := range playerIDs {
		p, err := r.playerRepo.FindByID(ctx, pid, false)
		if err != nil {
			return nil, fmt.Errorf("failed to load player %d: %w", pid, err)
		}
		players[i] = p
	}

	holeRows, err := r.db.Query(ctx, `
        SELECT hole_number, par, handicap_index FROM game_course_holes
        WHERE game_id = $1 ORDER BY hole_number
    `, id)
	if err != nil {
		return nil, err
	}
	defer holeRows.Close()

	c := &course.Course{ID: courseID, Name: courseName}
	for holeRows.Next() {
		var h course.Hole
		if err := holeRows.Scan(&h.Number, &h.Par, &h.HandicapIndex); err != nil {
			return nil, err
		}
		c.HolesData = append(c.HolesData, h)
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
		HoleResults: make(map[int]*domainWolf.HoleResult), // empty until scoring is built
	}, nil
}
