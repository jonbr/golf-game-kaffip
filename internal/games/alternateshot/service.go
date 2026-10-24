package alternateshot

import (
	"context"
	"fmt"
	"log/slog"

	"golf-game-kaffip/internal/application"
	domainGame "golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
)

// Repository is everything the Alternate Shot service needs from storage.
// The postgres implementation lives in repo.go.
type Repository interface {
	CreateGame(ctx context.Context, g *Game) error
	LoadGame(ctx context.Context, gameID string) (*Game, error)
	SaveHoleResult(ctx context.Context, g *Game, holeNumber int) error

	PlayerExists(ctx context.Context, playerID int64) (bool, error)
	PlayersInActiveGame(ctx context.Context, playerIDs []int64) (int64, error)
}

type Service struct {
	games   Repository
	players player.Repository
	courses application.ExternalCourseService
	logger  *slog.Logger
}

func NewService(games Repository, players player.Repository, courses application.ExternalCourseService, logger *slog.Logger) *Service {
	return &Service{games: games, players: players, courses: courses, logger: logger}
}

func (s *Service) CreateGame(ctx context.Context, req CreateRequest) (string, error) {
	if len(req.TeamA) != 2 || len(req.TeamB) != 2 {
		return "", application.NewServiceError("invalid_team_size", map[string]any{
			"team_a_size": len(req.TeamA),
			"team_b_size": len(req.TeamB),
		})
	}

	playerIDs := append(append([]int64{}, req.TeamA...), req.TeamB...)

	seen := make(map[int64]struct{}, len(playerIDs))
	for _, id := range playerIDs {
		if _, dup := seen[id]; dup {
			return "", application.NewServiceError("duplicate_player", map[string]any{"player_id": id})
		}
		seen[id] = struct{}{}
	}

	if err := application.ValidatePlayersExist(ctx, s.games, s.logger, playerIDs); err != nil {
		return "", err
	}
	if err := application.ValidateActiveGameConflict(ctx, s.games, s.logger, playerIDs); err != nil {
		return "", err
	}

	players, err := application.LoadPlayers(ctx, s.players, playerIDs)
	if err != nil {
		return "", fmt.Errorf("load players: %w", err)
	}

	c, err := application.FetchCourse(ctx, &s.courses, s.logger, req.CourseID)
	if err != nil {
		return "", err
	}

	g, err := NewGame(application.NewEntityID("game"), c, players[:2], players[2:], domainGame.Variant(req.Variant))
	if err != nil {
		return "", application.NewServiceError("invalid_game", map[string]any{"underlying": err.Error()})
	}

	if err := s.games.CreateGame(ctx, g); err != nil {
		return "", fmt.Errorf("create alternate shot game: %w", err)
	}
	return g.ID, nil
}

func (s *Service) SetHoleScore(ctx context.Context, gameID string, holeNumber int, req SetHoleScoreRequest) (*Game, error) {
	g, err := s.games.LoadGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	hole, err := g.HoleInfo(holeNumber)
	if err != nil {
		return nil, application.NewServiceError("invalid_hole_score", map[string]any{"underlying": err.Error()})
	}

	result, err := g.CalculateHoleResult(hole, req.TeamAGross, req.TeamBGross)
	if err != nil {
		return nil, application.NewServiceError("invalid_hole_score", map[string]any{"underlying": err.Error()})
	}

	if err := g.SetHoleScore(holeNumber, result); err != nil {
		return nil, application.NewServiceError("invalid_hole_score", map[string]any{"underlying": err.Error()})
	}

	if err := s.games.SaveHoleResult(ctx, g, holeNumber); err != nil {
		return nil, fmt.Errorf("save hole result: %w", err)
	}
	return g, nil
}

func (s *Service) GetGame(ctx context.Context, gameID string) (*Game, error) {
	return s.games.LoadGame(ctx, gameID)
}
