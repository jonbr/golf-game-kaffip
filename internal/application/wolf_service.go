package application

import (
	"context"
	"errors"
	"fmt"
	"golf-game-kaffip/internal/api/dto"
	"golf-game-kaffip/internal/domain/game"
	domainGame "golf-game-kaffip/internal/domain/game"
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

func (s *WolfGameService) CreateGame(ctx context.Context, gameType domainGame.GameType, req dto.CreateWolfGameRequest) (*wolf.Game, error) {
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

	gameID := fmt.Sprintf("game_%d", time.Now().UnixNano())

	g, err := wolf.NewGame(gameID, course, players)
	if err != nil {
		return nil, NewServiceError("invalid_wolf_game", map[string]any{"underlying": err.Error()})
	}

	if err := s.wolfGames.CreateGame(ctx, g); err != nil {
		return nil, fmt.Errorf("failed to save wolf game: %w", err)
	}

	return g, nil
}

func (s *WolfGameService) GetGame(ctx context.Context, id string) (*wolf.Game, error) {
	g, err := s.wolfGames.LoadGame(ctx, id)
	if err != nil {
		if errors.Is(err, wolf.ErrGameNotFound) {
			return nil, NewServiceError("wolf_game_not_found", map[string]any{"wolf_game_id": id})
		}
		return nil, err
	}
	return g, nil
}

func (s *WolfGameService) SetHoleScore(ctx context.Context, gameID string, holeNumber int, req dto.SetWolfHoleScoreRequest) (*wolf.Game, error) {
	logger := logging.FromCtx(ctx)

	g, err := s.wolfGames.LoadGame(ctx, gameID)
	if err != nil {
		if errors.Is(err, wolf.ErrGameNotFound) {
			return nil, NewServiceError("wolf_game_not_found", map[string]any{"wolf_game_id": gameID})
		}
		return nil, err
	}

	if len(req.Scores) != 4 {
		return nil, NewServiceError("invalid_score_count", map[string]any{
			"expected": 4,
			"got":      len(req.Scores),
		})
	}

	mode := wolf.WolfMode(req.Mode)
	if mode != wolf.WolfModePartnered && mode != wolf.WolfModeLone {
		return nil, NewServiceError("invalid_wolf_mode", map[string]any{"mode": req.Mode})
	}

	inputs := make([]wolf.PlayerScoreInput, 0, 4)
	for _, s := range req.Scores {
		inputs = append(inputs, wolf.PlayerScoreInput{PlayerID: s.PlayerID, Gross: s.Gross})
	}

	if err := g.SetHoleScore(holeNumber, req.WolfPlayerID, mode, req.PartnerID, inputs); err != nil {
		logger.Error("failed to set wolf hole score", "wolf_game_id", gameID, "hole_number", holeNumber, "error", err)
		return nil, NewServiceError("invalid_wolf_hole_score", map[string]any{
			"wolf_game_id": gameID,
			"hole_number":  holeNumber,
			"underlying":   err.Error(),
		})
	}

	if err := s.wolfGames.SaveHoleResult(ctx, g, holeNumber); err != nil {
		return nil, fmt.Errorf("failed to persist wolf hole %d result: %w", holeNumber, err)
	}

	return g, nil
}
