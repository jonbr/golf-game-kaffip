package application

import (
	"context"
	"fmt"
	"golf-game-kaffip/internal/domain/cup"
	"golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/wolf"
)

type CupService struct {
	cups      cup.Repository
	games     game.Repository
	wolfGames wolf.Repository
}

// CupDetail is everything GetCup needs to render: the cup itself, plus
// its live-computed aggregate score. Score is nver stored - always
// derived fresh from each match's current state.
type CupDetail struct {
	Cup   *cup.Cup
	Score cup.Score
}

func NewCupService(cups cup.Repository, games game.Repository, wolfGames wolf.Repository) *CupService {
	return &CupService{cups: cups, games: games, wolfGames: wolfGames}
}

func (s *CupService) GetCup(ctx context.Context, id string) (*CupDetail, error) {
	c, err := s.cups.LoadCup(ctx, id)
	if err != nil {
		return nil, err
	}

	var contributions []cup.MatchContribution

	for _, gameID := range c.MatchIDs {
		contribution, ok, err := s.resolveMatch(ctx, gameID, c.Players)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve match %s: %w", gameID, err)
		}
		if ok {
			contributions = append(contributions, contribution)
		}
	}

	return &CupDetail{
		Cup:   c,
		Score: cup.ComputeScore(contributions),
	}, nil
}

// resolveMatch figures out whether gameID belongs to the two-sided game
// family (points_play/match_play) or wolf, loads it from the right
// repository, and calls the matching resolver.
func (s *CupService) resolveMatch(ctx context.Context, gameID string, roster map[int64]cup.Side) (cup.MatchContribution, bool, error) {
	gameType, err := s.games.GetGameType(ctx, gameID)
	if err != nil {
		return cup.MatchContribution{}, false, fmt.Errorf("failed to resolve game for match %s: %w", gameID, err)
	}

	if gameType == game.GameTypeWolf {
		wg, err := s.wolfGames.LoadGame(ctx, gameID)
		if err != nil {
			return cup.MatchContribution{}, false, err
		}
		c, ok := resolveWolfContribution(wg, wg.Standings(), roster)
		return c, ok, nil
	}

	g, err := s.games.LoadGame(ctx, gameID)
	if err != nil {
		return cup.MatchContribution{}, false, err
	}
	c, ok := resolveGameContribution(g, roster)
	return c, ok, nil
}
