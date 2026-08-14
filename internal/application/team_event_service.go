package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golf-game-kaffip/internal/api/dto"
	domainGame "golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
	"golf-game-kaffip/internal/infrastructure/external/opengolfapi"
	"golf-game-kaffip/internal/logging"
)

type TeamEventService struct {
	events                domainGame.TeamEventRepository
	players               player.Repository
	games                 domainGame.Repository
	externalCourseService *ExternalCourseService
}

func NewTeamEventService(
	events domainGame.TeamEventRepository,
	playersRepo player.Repository,
	gamesRepo domainGame.Repository,
	externalAPI opengolfapi.ClientInterface,
) *TeamEventService {
	return &TeamEventService{
		events:                events,
		players:               playersRepo,
		games:                 gamesRepo,
		externalCourseService: NewExternalCourseService(externalAPI),
	}
}

func (s *TeamEventService) CreateEvent(ctx context.Context, req dto.CreateEventRequest) (*domainGame.TeamEvent, error) {
	logger := logging.FromCtx(ctx)

	if len(req.Matches) == 0 {
		return nil, NewServiceError("invalid_event", map[string]any{"reason": "at least one match required"})
	}

	var allPlayerIDs []int64
	for _, m := range req.Matches {
		allPlayerIDs = append(allPlayerIDs, m.TeamA...)
		allPlayerIDs = append(allPlayerIDs, m.TeamB...)
	}

	if err := validateNoDuplicatePlayersInEvent(allPlayerIDs); err != nil {
		return nil, err
	}
	if err := s.validatePlayersExist(ctx, logger, allPlayerIDs); err != nil {
		return nil, err
	}
	if err := s.validateActiveGameConflict(ctx, allPlayerIDs); err != nil {
		return nil, err
	}

	course, err := fetchCourse(ctx, s.externalCourseService, logger, req.CourseID)
	if err != nil {
		return nil, err
	}

	eventID := fmt.Sprintf("event_%d", time.Now().UnixNano())

	matches := make([]*domainGame.Game, 0, len(req.Matches))
	for i, m := range req.Matches {
		gameType := domainGame.GameType(m.GameType)

		teamAPlayers, err := s.loadPlayers(ctx, m.TeamA)
		if err != nil {
			return nil, err
		}
		teamBPlayers, err := s.loadPlayers(ctx, m.TeamB)
		if err != nil {
			return nil, err
		}

		matchID := fmt.Sprintf("game_%d_%d", time.Now().UnixNano(), i)
		match, err := domainGame.NewGame(matchID, course, teamAPlayers, teamBPlayers, gameType, domainGame.Variant(req.Variant))
		if err != nil {
			return nil, NewServiceError("invalid_game_params", map[string]any{"underlying": err.Error()})
		}
		matches = append(matches, match)
	}

	event, err := domainGame.NewTeamEvent(eventID, course, domainGame.Variant(req.Variant), matches)
	if err != nil {
		return nil, NewServiceError("invalid_event", map[string]any{"underlying": err.Error()})
	}

	if err := s.events.CreateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	return event, nil
}

func (s *TeamEventService) GetEvent(ctx context.Context, id string) (*domainGame.TeamEvent, error) {
	event, err := s.events.LoadEvent(ctx, id)
	if err != nil {
		if errors.Is(err, domainGame.ErrTeamEventNotFound) {
			return nil, NewServiceError("team_event_not_found", map[string]any{"event_id": id})
		}
		return nil, err
	}
	return event, nil
}

func (s *TeamEventService) FinishEvent(ctx context.Context, id string) error {
	if _, err := s.GetEvent(ctx, id); err != nil {
		return err
	}
	return s.events.FinishEvent(ctx, id)
}

func (s *TeamEventService) loadPlayers(ctx context.Context, ids []int64) ([]*player.Player, error) {
	players := make([]*player.Player, 0, len(ids))
	for _, id := range ids {
		p, err := s.players.FindByID(ctx, id, false)
		if err != nil {
			return nil, fmt.Errorf("failed to load player %d: %w", id, err)
		}
		players = append(players, p)
	}
	return players, nil
}

func (s *TeamEventService) validatePlayersExist(ctx context.Context, logger *slog.Logger, playerIDs []int64) error {
	for _, pid := range playerIDs {
		exists, err := s.games.PlayerExists(ctx, pid)
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

func (s *TeamEventService) validateActiveGameConflict(ctx context.Context, playerIDs []int64) error {
	blockingPlayer, err := s.games.PlayersInActiveGame(ctx, playerIDs)
	if err != nil {
		return NewServiceError("internal_error", map[string]any{"underlying": err.Error()})
	}
	if blockingPlayer != 0 {
		return NewServiceError("player_in_active_game", map[string]any{"player_id": blockingPlayer})
	}
	return nil
}

func validateNoDuplicatePlayersInEvent(playerIDs []int64) error {
	seen := make(map[int64]bool, len(playerIDs))
	for _, pid := range playerIDs {
		if seen[pid] {
			return NewServiceError("duplicate_player_in_event", map[string]any{"player_id": pid})
		}
		seen[pid] = true
	}
	return nil
}
