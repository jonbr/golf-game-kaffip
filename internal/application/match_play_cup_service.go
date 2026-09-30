package application

import (
	"context"
	"fmt"
	"golf-game-kaffip/internal/api/dto"
	"golf-game-kaffip/internal/domain/cup"
)

type MatchPlayCupService struct {
	cups      cup.Repository
	matchPlay *MatchPlayService
}

func NewMatchPlayCupService(cups cup.Repository, matchPlay *MatchPlayService) *MatchPlayCupService {
	return &MatchPlayCupService{cups: cups, matchPlay: matchPlay}
}

func (s *MatchPlayCupService) CreateCup(ctx context.Context, req dto.CreateMatchPlayCupRequest) (*cup.Cup, error) {
	roster := make(map[int64]cup.Side, len(req.Roster))
	for _, entry := range req.Roster {
		roster[entry.PlayerID] = cup.Side(entry.Side)
	}

	matchIDs := make([]string, 0, len(req.Matches))
	for _, spec := range req.Matches {
		g, err := s.matchPlay.CreateGame(ctx, dto.CreateMatchPlayRequest{
			CourseID: spec.CourseID,
			Variant:  spec.Variant,
			PlayerA:  spec.PlayerA,
			PlayerB:  spec.PlayerB,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create match %d og %d: %w", len(matchIDs)+1, len(req.Matches), err)
		}
		matchIDs = append(matchIDs, g.ID)
	}

	c, err := cup.NewCup(newEntityID("cup"), req.Name, roster, matchIDs)
	if err != nil {
		return nil, NewServiceError("invalid_cup", map[string]any{"underlying": err.Error()})
	}

	if err := s.cups.CreateCup(ctx, c); err != nil {
		return nil, fmt.Errorf("failed to save cup: %w", err)
	}

	return c, nil
}
