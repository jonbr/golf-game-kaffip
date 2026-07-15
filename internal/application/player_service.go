package application

import (
	"context"
	"errors"
	"golf-game-kaffip/internal/domain/player"
)

type PlayerService struct {
	players player.Repository
}

func NewPlayerService(players player.Repository) *PlayerService {
	return &PlayerService{players: players}
}

func (s *PlayerService) CreatePlayer(ctx context.Context, name string, handicap float64) (*player.Player, error) {
	if name == "" {
		return nil, NewServiceError("validation_error", map[string]any{"field": "name", "reason": "required"})
	}
	if handicap < 0 {
		return nil, NewServiceError("validation_error", map[string]any{"field": "handicap", "reason": "must be non-negative"})
	}

	p := &player.Player{
		Name:     name,
		Handicap: handicap,
	}

	if err := s.players.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *PlayerService) GetPlayers(ctx context.Context) ([]*player.Player, error) {
	return s.players.FindAll(ctx)
}

func (s *PlayerService) GetPlayer(ctx context.Context, id int64) (*player.Player, error) {
	return s.findPlayer(ctx, id, true)
}

func (s *PlayerService) UpdatePlayer(ctx context.Context, id int64, req player.UpdatePlayerParams) (*player.Player, error) {
	if err := s.ensureNotInActiveGame(ctx, id); err != nil {
		return nil, err
	}

	existing, err := s.findPlayer(ctx, id, false)
	if err != nil {
		return nil, err
	}

	// 3. Apply updates
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Handicap != nil {
		existing.Handicap = *req.Handicap
	}

	// 4. Persist
	if err := s.players.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *PlayerService) SoftDeletePlayer(ctx context.Context, id int64) error {
	if err := s.ensureNotInActiveGame(ctx, id); err != nil {
		return err
	}
	if err := s.players.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, player.ErrPlayerNotFound) {
			return NewServiceError("player_not_found", map[string]any{"player_id": id})
		}
		return err
	}
	return nil
}

func (s *PlayerService) ensureNotInActiveGame(ctx context.Context, id int64) error {
	activeGameID, err := s.players.GetActiveGameForPlayer(ctx, id)
	if err != nil {
		return err
	}
	if activeGameID != nil {
		return NewServiceError("player_in_active_game", map[string]any{"player_id": id, "game_id": activeGameID})
	}
	return nil
}

func (s *PlayerService) findPlayer(ctx context.Context, id int64, forUpdate bool) (*player.Player, error) {
	p, err := s.players.FindByID(ctx, id, forUpdate)
	if err != nil {
		if errors.Is(err, player.ErrPlayerNotFound) {
			return nil, NewServiceError("player_not_found", map[string]any{"player_id": id})
		}
		return nil, err
	}
	return p, nil
}
