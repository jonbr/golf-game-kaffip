package application

import (
	"context"
	"fmt"
	"golf-game-kaffip/internal/api/dto"
	"golf-game-kaffip/internal/domain/cup"
)

type WolfCupService struct {
	cups     cup.Repository
	wolfGame *WolfGameService
}

func NewWolfGameCupService(cups cup.Repository, wolfGame *WolfGameService) *WolfCupService {
	return &WolfCupService{cups: cups, wolfGame: wolfGame}
}

func (s *WolfCupService) CreateCup(ctx context.Context, req dto.CreateWolfCupRequest) (*cup.Cup, error) {
	roster := make(map[int64]cup.Side, len(req.Roster))
	for _, entry := range req.Roster {
		roster[entry.PlayerID] = cup.Side(entry.Side)
	}

	matchIDs := make([]string, 0, len(req.Matches))
	for _, spec := range req.Matches {
		wg, err := s.wolfGame.CreateGame(ctx, dto.CreateWolfGameRequest{
			CourseID:  spec.CourseID,
			PlayerIDs: spec.PlayerIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create match %d of %d: %w", len(matchIDs)+1, len(req.Matches), err)

		}
		matchIDs = append(matchIDs, wg.ID)
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
