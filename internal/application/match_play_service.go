package application

import (
	"context"
	"fmt"
	"golf-game-kaffip/internal/api/dto"
	"golf-game-kaffip/internal/domain/game"
	domainGame "golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/matchplay"
	"golf-game-kaffip/internal/domain/player"
	"golf-game-kaffip/internal/infrastructure/external/opengolfapi"
	"golf-game-kaffip/internal/logging"
	"time"
)

type MatchPlayService struct {
	games                 game.Repository
	players               player.Repository
	externalCourseService *ExternalCourseService
}

func NewMatchPlayService(
	games game.Repository,
	players player.Repository,
	externalAPI opengolfapi.ClientInterface,
) *MatchPlayService {
	return &MatchPlayService{
		games:                 games,
		players:               players,
		externalCourseService: NewExternalCourseService(externalAPI),
	}
}

func (s *MatchPlayService) CreateGame(ctx context.Context, gameType domainGame.GameType, req dto.CreateMatchPlayRequest) (*domainGame.Game, error) {
	logger := logging.FromCtx(ctx)

	if err := validateTeamSize(gameType, req.TeamA, req.TeamB); err != nil {
		return nil, err
	}

	playersIDs := append(req.TeamA, req.TeamB...)

	if err := validatePlayersExist(ctx, s.games, logger, playersIDs); err != nil {
		return nil, err
	}
	if err := validateActiveGameConflict(ctx, s.games, logger, playersIDs); err != nil {
		return nil, err
	}

	teamAPlayers, err := loadPlayers(ctx, s.players, req.TeamA)
	if err != nil {
		return nil, err
	}
	teamBPlayers, err := loadPlayers(ctx, s.players, req.TeamB)
	if err != nil {
		return nil, err
	}

	course, err := fetchCourse(ctx, s.externalCourseService, logger, req.CourseID)
	if err != nil {
		return nil, err
	}

	gameID := fmt.Sprintf("game_%d", time.Now().UnixNano())

	// 5. Create domain game
	g, err := domainGame.NewGame(gameID, course, teamAPlayers, teamBPlayers, domainGame.GameType(gameType), domainGame.Variant(req.Variant))
	if err != nil {
		return nil, NewServiceError("invalid_game_params", map[string]any{"underlying": err.Error()})
	}

	// 4. Persist
	if err := s.games.CreateGame(ctx, g); err != nil {
		return nil, fmt.Errorf("failed to save game: %w", err)
	}

	return g, nil
}

func (s *MatchPlayService) SetHoleScore(ctx context.Context, gameID string, holeNumber int, req dto.SetHoleScoreRequest) (*domainGame.Game, error) {
	g, err := s.games.LoadGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	inputs, err := buildScoreInputs(g, req.Scores) // existing shared helper
	if err != nil {
		return nil, err
	}

	hole, err := g.HoleInfo(holeNumber)
	if err != nil {
		return nil, NewServiceError("invalid_hole_score", map[string]any{"underlying": err.Error()})
	}

	result, err := matchplay.CalculateHoleResult(hole, inputs, g.Variant, g.PlayersByID(), g.TotalHoles())
	if err != nil {
		return nil, NewServiceError("invalid_hole_score", map[string]any{"underlying": err.Error()})
	}

	if err := g.SetHoleScore(holeNumber, result); err != nil {
		return nil, NewServiceError("invalid_hole_score", map[string]any{"underlying": err.Error()})
	}

	if err := s.games.SaveHoleResult(ctx, g, holeNumber); err != nil {
		return nil, fmt.Errorf("failed to persist hole %d result: %w", holeNumber, err)
	}

	return g, nil
}
