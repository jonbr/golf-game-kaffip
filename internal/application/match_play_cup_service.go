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

	if err := validateMatchPlayCupMatches(roster, req.Matches); err != nil {
		return nil, err
	}

	matchIDs := make([]string, 0, len(req.Matches))
	for _, spec := range req.Matches {
		g, err := s.matchPlay.CreateGame(ctx, dto.CreateMatchPlayRequest{
			CourseID: spec.CourseID,
			Variant:  spec.Variant,
			TeamA:    []int64{spec.TeamA[0]},
			TeamB:    []int64{spec.TeamB[0]},
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

// validateMatchPlayCupMatches confirms every player referenced in matches
// is on the roster, and that each matche's two players are on the opposite
// cup sides - a match with both players on the same side would produce
// a nonesecial event contribution at score time.
func validateMatchPlayCupMatches(roster map[int64]cup.Side, matches []dto.CreateMatchPlayRequest) error {
	for i, spec := range matches {
		sideA, okA := roster[spec.TeamA[0]]
		if !okA {
			return NewServiceError("player_not_on_cup_roster", map[string]any{
				"player_id": spec.TeamA[0], "match_index": i,
			})
		}

		sideB, okB := roster[spec.TeamB[0]]
		if !okB {
			// Fixed: Typo in error string ("onb") and match_index was hardcoded to 1
			return NewServiceError("player_not_on_cup_roster", map[string]any{
				"player_id": spec.TeamB[0], "match_index": i,
			})
		}

		if sideA == sideB {
			// Fixed: match_index was hardcoded to 1
			return NewServiceError("match_players_same_side", map[string]any{
				"match_index": i,
				"player_a":    spec.TeamA[0],
				"player_b":    spec.TeamB[0],
				"side":        string(sideA),
			})
		}
	}
	return nil
}
