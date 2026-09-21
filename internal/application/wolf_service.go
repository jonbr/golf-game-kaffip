package application

import (
	"context"
	"fmt"
	"golf-game-kaffip/internal/api/dto"
	"golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
	"golf-game-kaffip/internal/domain/wolf"
	"golf-game-kaffip/internal/infrastructure/external/opengolfapi"
	"golf-game-kaffip/internal/logging"
	"time"
)

type WolfGameService struct {
	games                 game.Repository
	wolfGames             wolf.Repository
	players               player.Repository
	externalCourseService *ExternalCourseService
}

func NewWolfGameService(
	games game.Repository,
	wolfGames wolf.Repository,
	players player.Repository,
	externalAPI opengolfapi.ClientInterface,
) *WolfGameService {
	return &WolfGameService{
		games:                 games,
		wolfGames:             wolfGames,
		players:               players,
		externalCourseService: NewExternalCourseService(externalAPI),
	}
}

func (s *WolfGameService) CreateGame(ctx context.Context, req dto.CreateWolfGameRequest) (*wolf.Game, error) {
	logger := logging.FromCtx(ctx)

	if len(req.PlayerIDs) != 4 {
		return nil, NewServiceError("invalid_wolf_players", map[string]any{"expected": 4, "got": len(req.PlayerIDs)})
	}

	if err := validatePlayersExist(ctx, s.games, logger, req.PlayerIDs); err != nil {
		return nil, err
	}
	if err := validateActiveGameConflict(ctx, s.games, logger, req.PlayerIDs); err != nil {
		return nil, err
	}

	resolvedPlayers, err := loadPlayers(ctx, s.players, req.PlayerIDs)
	if err != nil {
		return nil, err
	}

	var players [4]*player.Player
	copy(players[:], resolvedPlayers)

	course, err := fetchCourse(ctx, s.externalCourseService, logger, req.CourseID)
	if err != nil {
		return nil, err
	}

	gameID := fmt.Sprintf("wolf_%d", time.Now().UnixNano())

	g, err := wolf.NewGame(gameID, course, players)
	if err != nil {
		return nil, NewServiceError("invalid_wolf_game", map[string]any{"underlying": err.Error()})
	}

	if err := s.wolfGames.CreateGame(ctx, g); err != nil {
		return nil, fmt.Errorf("failed to save wolf game: %w", err)
	}

	return g, nil
}
