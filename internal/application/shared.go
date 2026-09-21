package application

import (
	"context"
	"errors"
	"fmt"
	domainCourse "golf-game-kaffip/internal/domain/course"
	"golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
	"log/slog"
)

// validatePlayersExist confirms every given player ID exists, using the
// shared game repository's PlayerExists check. Used by any service that
// creates a game-like aggregate referencing players (GameService,
// WolfGameService, TeamEventService).
func validatePlayersExist(ctx context.Context, games game.Repository, logger *slog.Logger, playerIDs []int64) error {
	for _, pid := range playerIDs {
		exists, err := games.PlayerExists(ctx, pid)
		if err != nil {
			return NewServiceError("internal_error", map[string]any{"underlying": err.Error()})
		}
		if !exists {
			logger.Error("player existence validation failed", "player_id", pid)
			return NewServiceError("player_not_found", map[string]any{"player_id": pid})
		}
	}
	return nil
}

// validateActiveGameConflict confirms none of the given players are
// currently in an unfinished game.
func validateActiveGameConflict(ctx context.Context, games game.Repository, logger *slog.Logger, playerIDs []int64) error {
	blockingPlayer, err := games.PlayersInActiveGame(ctx, playerIDs)
	if err != nil {
		return NewServiceError("internal_error", map[string]any{"underlying": err.Error()})
	}
	if blockingPlayer != 0 {
		logger.Error("active game conflict validation failed", "player_id", blockingPlayer)
		return NewServiceError("player_in_active_game", map[string]any{"player_id": blockingPlayer})
	}
	return nil
}

// fetchCourse centralizes external course lookup + error mapping. Only
// CreateGame should call this — course data is fetched once and stored
// locally, everything else reads it back via LoadGame/ListSummaries.
func fetchCourse(ctx context.Context, ecs *ExternalCourseService, logger *slog.Logger, courseID string) (*domainCourse.Course, error) {
	course, err := ecs.GetExternalCourse(ctx, courseID)
	if err != nil {
		logger.Info("external course lookup failed", "course_id", courseID, "error", err.Error())
		if errors.Is(err, domainCourse.ErrCourseNotFound) {
			return nil, domainCourse.ErrCourseNotFound
		}
		return nil, NewServiceError("external_api_error", map[string]any{"underlying": err.Error()})
	}
	return course, nil
}

// loadPlayers fetches the full Player record for each given ID, in order.
// Used by any service that needs actual player data (not just an
// existence check) to construct a game-like aggregate.
func loadPlayers(ctx context.Context, players player.Repository, ids []int64) ([]*player.Player, error) {
	result := make([]*player.Player, 0, len(ids))
	for _, id := range ids {
		p, err := players.FindByID(ctx, id, false)
		if err != nil {
			return nil, fmt.Errorf("failed to load player with ID %d: %w", id, err)
		}
		result = append(result, p)
	}
	return result, nil
}
